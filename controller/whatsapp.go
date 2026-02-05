package controller

import (
	"fmt"
	"strings"
	"whatsapp-gmail-bot/constants"
	"whatsapp-gmail-bot/models"
	"whatsapp-gmail-bot/service"

	"github.com/gofiber/fiber/v2"
)

type CommandHandler func(ctx *fiber.Ctx, msg, sender string) error

type WhatsappController struct {
	pendingDrafts map[string]string
	commands      map[string]CommandHandler
}

func NewWhatsappController() *WhatsappController {
	wppc := &WhatsappController{
		pendingDrafts: make(map[string]string),
		commands:      make(map[string]CommandHandler),
	}
	wppc.registerCommands()
	return wppc
}

func (wppc *WhatsappController) registerCommands() {
	wppc.commands[constants.COMMAND_CANCEL] = wppc.handleCancel
	wppc.commands[constants.COMMAND_CONFIRM] = wppc.handleConfirm
}

func (wppc *WhatsappController) SendWhatsappController(ctx *fiber.Ctx) error {
	msg := ctx.FormValue("Body")
	sender := ctx.FormValue("From")

	if sender != constants.ONLY_SENDER {
		fmt.Printf("⚠️ Intento de acceso no autorizado de: %s\n", sender)
		return wppc.reply(ctx, "⛔ Acceso denegado. Este bot es privado.")
	}

	fmt.Printf("📩 Mensaje de Facu: %s\n", msg)

	return wppc.handleCommand(ctx, msg, sender)
}
func (wppc *WhatsappController) handleCommand(ctx *fiber.Ctx, msg, sender string) error {
	msgLower := strings.ToLower(strings.TrimSpace(msg))

	// 1. PRIORIDAD ALTA: Comandos estrictos (Cancelar, Confirmar)
	// Si el usuario dice "cancelar" o "si", no gastamos tokens de IA, ejecutamos directo.
	if handler, exists := wppc.commands[msgLower]; exists {
		return handler(ctx, msg, sender)
	}

	// 2. PRIORIDAD INTELIGENTE: Todo lo demás va a Gemini
	// Aquí llamamos a la función que creamos en el paso anterior (service/ai.go)
	fmt.Println("🧠 Consultando a Gemini...")
	aiDecision, err := service.ProcessIntent(msg)

	if err != nil {
		fmt.Printf("Error IA: %v\n", err)
		return wppc.reply(ctx, "⚠️ La IA tuvo un error. Intenta de nuevo.")
	}

	// 3. Ejecutar según lo que decidió la IA
	switch aiDecision.Type {
	case "CHAT":
		return wppc.reply(ctx, aiDecision.Content)

	case "CONFIRM_SEND":
		// Si la IA detecta que el usuario quiere enviar lo pendiente (ej: "Dale mandalo")
		return wppc.handleConfirm(ctx, msg, sender)

	case "EMAIL_DRAFT":
		return wppc.handleAiDraft(ctx, sender, aiDecision)

	default:
		return wppc.reply(ctx, "🤖 No entendí la respuesta de la IA.")
	}
}

// Nueva función para manejar el borrador que generó la IA
func (wppc *WhatsappController) handleAiDraft(ctx *fiber.Ctx, sender string, data *models.AIResponse) error {

	// Validar si Gemini encontró el destinatario
	if data.To == "PENDIENTE" || data.To == "" {
		return wppc.reply(ctx, fmt.Sprintf("✍️ Entendido: '%s'\n\nPero... ¿A quién se lo mando? (Dime el email)", data.Subject))
	}

	// Crear el borrador con los datos enriquecidos por la IA
	draftID, err := service.CreateDraft(data.To, data.Subject, data.Content)
	if err != nil {
		return wppc.reply(ctx, "❌ Error al crear borrador: "+err.Error())
	}

	// Guardar en memoria
	wppc.pendingDrafts[sender] = draftID

	// Mostrar preview elegante
	previewMsg := fmt.Sprintf("📝 *Borrador IA Creado*\n\n*Para:* %s\n*Asunto:* %s\n\n%s\n\n_¿Lo envío? (Responde Sí)_",
		data.To, data.Subject, data.Content)

	return wppc.reply(ctx, previewMsg)
}

func (wppc *WhatsappController) handleCancel(ctx *fiber.Ctx, msg, sender string) error {
	delete(wppc.pendingDrafts, sender)
	return wppc.reply(ctx, "🗑️ Operación cancelada. Memoria limpia.")
}

func (wppc *WhatsappController) handleConfirm(ctx *fiber.Ctx, msg, sender string) error {
	draftID, exists := wppc.pendingDrafts[sender]
	if !exists {
		return wppc.reply(ctx, "🤷‍♂️ No tienes borradores pendientes.")
	}

	if err := service.SendDraft(draftID); err != nil {
		return wppc.reply(ctx, "❌ Error al enviar: "+err.Error())
	}

	delete(wppc.pendingDrafts, sender)
	return wppc.reply(ctx, "🚀 ¡Correo enviado exitosamente!")
}

func (wppc *WhatsappController) reply(ctx *fiber.Ctx, message string) error {
	ctx.Set("Content-Type", "text/xml")
	return ctx.SendString(fmt.Sprintf(models.ReplyMessage, message))
}
