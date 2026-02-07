package controller

import (
	"fmt"
	"os"
	"strings"
	"whatsapp-gmail-bot/constants"
	"whatsapp-gmail-bot/models"
	"whatsapp-gmail-bot/service"

	"github.com/gofiber/fiber/v2"
)

// Estructuras para leer el JSON de Meta/Facebook
type MetaWebhook struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Messages []struct {
					From string `json:"from"`
					Text struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

type CommandHandler func(sender, msg string) // Quitamos ctx, ya no lo necesitamos para responder

type WhatsappController struct {
	pendingDrafts map[string]string
	commands      map[string]CommandHandler
	contacts      map[string]string
	contactsList  string
}

func NewWhatsappController(contacts map[string]string, contactsList string) *WhatsappController {
	wppc := &WhatsappController{
		pendingDrafts: make(map[string]string),
		commands:      make(map[string]CommandHandler),
		contacts:      contacts,
		contactsList:  contactsList,
	}
	wppc.registerCommands()
	return wppc
}

func (wppc *WhatsappController) registerCommands() {
	wppc.commands[constants.COMMAND_CANCEL] = wppc.handleCancel
	wppc.commands[constants.COMMAND_CONFIRM] = wppc.handleConfirm
}

func (wppc *WhatsappController) VerifyWebhook(ctx *fiber.Ctx) error {
	verifyToken := os.Getenv("META_VERIFY_TOKEN")
	mode := ctx.Query("hub.mode")
	token := ctx.Query("hub.verify_token")
	challenge := ctx.Query("hub.challenge")

	if mode == "subscribe" && token == verifyToken {
		return ctx.SendString(challenge)
	}
	return ctx.SendStatus(403)
}

func (wppc *WhatsappController) ProcessWebhook(ctx *fiber.Ctx) error {
	var payload MetaWebhook
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.SendStatus(400)
	}

	if len(payload.Entry) > 0 && len(payload.Entry[0].Changes) > 0 {
		change := payload.Entry[0].Changes[0].Value
		if len(change.Messages) > 0 {
			msgObj := change.Messages[0]
			sender := msgObj.From
			text := msgObj.Text.Body

			fmt.Printf("📩 Mensaje de %s: %s\n", sender, text)

			go wppc.handleCommand(sender, text)
		}
	}

	return ctx.SendStatus(200)
}

func (wppc *WhatsappController) handleCommand(sender, msg string) {
	msgLower := strings.ToLower(strings.TrimSpace(msg))

	// 1. Verificar comandos directos
	if handler, exists := wppc.commands[msgLower]; exists {
		handler(sender, msg)
		return
	}

	// 2. Consultar a Gemini
	fmt.Println("🧠 Consultando a Gemini...")
	aiDecision, err := service.ProcessIntent(msg, wppc.contactsList) // Asegúrate que ProcessIntent acepte estos params

	if err != nil {
		fmt.Printf("Error IA: %s\n", err.Error())
		wppc.reply(sender, "⚠️ La IA tuvo un error. Intenta de nuevo.")
		return
	}

	switch aiDecision.Type {
	case "CHAT":
		wppc.reply(sender, aiDecision.Content)

	case "CONFIRM_SEND":
		wppc.handleConfirm(sender, msg)

	case "EMAIL_DRAFT":
		wppc.handleAiDraft(sender, aiDecision)

	default:
		wppc.reply(sender, "🤖 No entendí la respuesta de la IA.")
	}
}

// handleAiDraft - Adaptado para no usar ctx
func (wppc *WhatsappController) handleAiDraft(sender string, data *models.AIResponse) {
	fmt.Printf("IA generó: To=%s, Subject=%s\n", data.To, data.Subject)

	if !wppc.validateAndResolveRecipient(data) {
		wppc.reply(sender, fmt.Sprintf("Entendido: '%s'\n\nPero... ¿A quién se lo mando? (Dime el email)", data.Subject))
		return
	}

	draftID, err := wppc.createDraft(data)
	if err != nil {
		fmt.Printf("Error creando draft: %s\n", err.Error())
		wppc.reply(sender, fmt.Sprintf("Error al crear borrador: %s", err.Error()))
		return
	}

	wppc.pendingDrafts[sender] = draftID
	wppc.reply(sender, wppc.buildPreviewMessage(data))
}

func (wppc *WhatsappController) handleCancel(sender, msg string) {
	delete(wppc.pendingDrafts, sender)
	wppc.reply(sender, "🗑️ Operación cancelada. Memoria limpia.")
}

func (wppc *WhatsappController) handleConfirm(sender, msg string) {
	draftID, exists := wppc.pendingDrafts[sender]
	if !exists {
		wppc.reply(sender, "🤷‍♂️ No tienes borradores pendientes.")
		return
	}

	if err := service.SendDraft(draftID); err != nil {
		wppc.reply(sender, fmt.Sprintf("❌ Error al enviar: %s", err.Error()))
		return
	}

	delete(wppc.pendingDrafts, sender)
	wppc.reply(sender, "🚀 Correo enviado exitosamente!")
}

// ---------------------------------------------------------
// FUNCION REPLY NUEVA (La clave del cambio)
// ---------------------------------------------------------
func (wppc *WhatsappController) reply(to, message string) {
	// Ya no escribimos en ctx.
	// Llamamos al servicio que dispara el mensaje a la API de Meta.
	err := service.SendWhatsAppMessage(to, message)
	if err != nil {
		fmt.Printf("❌ Error enviando WhatsApp a %s: %v\n", to, err)
	}
}

// --- Helpers (Sin cambios de lógica, solo firmas si es necesario) ---

func (wppc *WhatsappController) validateAndResolveRecipient(data *models.AIResponse) bool {
	if data.To == "PENDIENTE" || data.To == "" {
		return false
	}
	data.To = strings.TrimSpace(data.To)
	if !wppc.isValidEmailFormat(data.To) {
		if email, found := wppc.resolveContactEmail(data.To); found {
			data.To = email
		} else {
			return false
		}
	}
	return true
}

func (wppc *WhatsappController) createDraft(data *models.AIResponse) (string, error) {
	return service.CreateDraft(data.To, data.Subject, data.Content)
}

func (wppc *WhatsappController) buildPreviewMessage(data *models.AIResponse) string {
	return fmt.Sprintf("*Borrador IA Creado*\n\n*Para:* %s\n*Asunto:* %s\n\n%s\n\n_¿Lo envío? (Responde Sí)_",
		data.To, data.Subject, data.Content)
}

func (wppc *WhatsappController) isValidEmailFormat(text string) bool {
	return strings.Contains(text, "@") && strings.Contains(text, ".")
}

func (wppc *WhatsappController) resolveContactEmail(name string) (string, bool) {
	nameLower := strings.ToLower(strings.TrimSpace(name))
	for contactName, email := range wppc.contacts {
		if strings.ToLower(contactName) == nameLower {
			return email, true
		}
	}
	for contactName, email := range wppc.contacts {
		if strings.Contains(strings.ToLower(contactName), nameLower) || strings.Contains(nameLower, strings.ToLower(contactName)) {
			return email, true
		}
	}
	return "", false
}

func (wppc *WhatsappController) StatusController(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"status": "running"})
}
