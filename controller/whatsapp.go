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

func (wppc *WhatsappController) StatusController(ctx *fiber.Ctx) error {
	fmt.Println("Llamada a Status exitosa!")
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "running",
	})
}

func (wppc *WhatsappController) SendWhatsappController(ctx *fiber.Ctx) error {
	msg := ctx.FormValue("Body")
	sender := ctx.FormValue("From")

	if sender != constants.ONLY_SENDER {
		fmt.Printf("Intento de acceso no autorizado de: %s\n", sender)
		return wppc.reply(ctx, "Acceso denegado. Este bot es privado.")
	}

	fmt.Printf("Mensaje de Facu: %s\n", msg)

	return wppc.handleCommand(ctx, msg, sender)
}
func (wppc *WhatsappController) handleCommand(ctx *fiber.Ctx, msg, sender string) error {
	msgLower := strings.ToLower(strings.TrimSpace(msg))

	if handler, exists := wppc.commands[msgLower]; exists {
		return handler(ctx, msg, sender)
	}

	fmt.Println("Consultando a Gemini...")
	aiDecision, err := service.ProcessIntent(msg, wppc.contactsList)

	if err != nil {
		fmt.Printf("Error IA: %s\n", err.Error())
		return wppc.reply(ctx, "La IA tuvo un error. Intenta de nuevo.")
	}

	switch aiDecision.Type {
	case "CHAT":
		return wppc.reply(ctx, aiDecision.Content)

	case "CONFIRM_SEND":
		return wppc.handleConfirm(ctx, msg, sender)

	case "EMAIL_DRAFT":
		return wppc.handleAiDraft(ctx, sender, aiDecision)

	default:
		return wppc.reply(ctx, "No entendí la respuesta de la IA.")
	}
}

// handleAiDraft - Maneja el borrador generado por la IA
func (wppc *WhatsappController) handleAiDraft(ctx *fiber.Ctx, sender string, data *models.AIResponse) error {
	fmt.Printf("IA generó: To=%s, Subject=%s, Content length=%d\n", data.To, data.Subject, len(data.Content))

	if !wppc.validateAndResolveRecipient(data) {
		return wppc.reply(ctx, fmt.Sprintf("Entendido: '%s'\n\nPero... ¿A quién se lo mando? (Dime el email)", data.Subject))
	}

	draftID, err := wppc.createDraft(data)
	if err != nil {
		fmt.Printf("Error creando draft: %s\n", err.Error())
		return wppc.reply(ctx, fmt.Sprintf("Error al crear borrador: %s", err.Error()))
	}

	wppc.pendingDrafts[sender] = draftID
	return wppc.reply(ctx, wppc.buildPreviewMessage(data))
}

// validateAndResolveRecipient - Valida y resuelve el destinatario del email
func (wppc *WhatsappController) validateAndResolveRecipient(data *models.AIResponse) bool {
	if data.To == "PENDIENTE" || data.To == "" {
		return false
	}

	data.To = strings.TrimSpace(data.To)

	if !wppc.isValidEmailFormat(data.To) {
		if email, found := wppc.resolveContactEmail(data.To); found {
			fmt.Printf("Contacto resuelto: '%s' -> '%s'\n", data.To, email)
			data.To = email
		} else {
			return false
		}
	}

	return true
}

// createDraft - Crea un borrador de email
func (wppc *WhatsappController) createDraft(data *models.AIResponse) (string, error) {
	return service.CreateDraft(data.To, data.Subject, data.Content)
}

// buildPreviewMessage - Construye el mensaje de preview del borrador
func (wppc *WhatsappController) buildPreviewMessage(data *models.AIResponse) string {
	return fmt.Sprintf("*Borrador IA Creado*\n\n*Para:* %s\n*Asunto:* %s\n\n%s\n\n_¿Lo envío? (Responde Sí)_",
		data.To, data.Subject, data.Content)
}

func (wppc *WhatsappController) handleCancel(ctx *fiber.Ctx, msg, sender string) error {
	delete(wppc.pendingDrafts, sender)
	return wppc.reply(ctx, "Operación cancelada. Memoria limpia.")
}

func (wppc *WhatsappController) handleConfirm(ctx *fiber.Ctx, msg, sender string) error {
	draftID, exists := wppc.pendingDrafts[sender]
	if !exists {
		return wppc.reply(ctx, "No tienes borradores pendientes.")
	}

	if err := service.SendDraft(draftID); err != nil {
		return wppc.reply(ctx, fmt.Sprintf("Error al enviar: %s", err.Error()))
	}

	delete(wppc.pendingDrafts, sender)
	return wppc.reply(ctx, "Correo enviado exitosamente!")
}

func (wppc *WhatsappController) reply(ctx *fiber.Ctx, message string) error {
	ctx.Set("Content-Type", "text/xml")
	return ctx.SendString(fmt.Sprintf(models.ReplyMessage, message))
}

// isValidEmailFormat - Valida si es un formato de email válido
func (wppc *WhatsappController) isValidEmailFormat(text string) bool {
	return strings.Contains(text, "@") && strings.Contains(text, ".")
}

// resolveContactEmail - Busca un nombre/apodo en la lista de contactos y devuelve su email
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
