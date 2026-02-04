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

	if strings.HasPrefix(msgLower, constants.COMMAND_MAIL_PREFIX) {
		return wppc.handleMail(ctx, msg, sender)
	}

	if handler, exists := wppc.commands[msgLower]; exists {
		return handler(ctx, msg, sender)
	}

	return wppc.reply(ctx, "🤖 No entendí. Comandos:\n- Mail a [email]: [mensaje]\n- Si (para confirmar)")
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

func (wppc *WhatsappController) handleMail(ctx *fiber.Ctx, msg, sender string) error {
	cleanMsg := strings.TrimSpace(msg[constants.MAIL_PREFIX_LENGTH:])
	parts := strings.SplitN(cleanMsg, ":", 2)

	if len(parts) < 2 {
		return wppc.reply(ctx, "⚠️ Formato incorrecto.\nUsa: Mail a correo@test.com: Hola mensaje")
	}

	toEmail := strings.TrimSpace(parts[0])
	bodyContent := strings.TrimSpace(parts[1])

	// Remover comillas simples o dobles del inicio y final si están presentes
	bodyContent = strings.Trim(bodyContent, `"'`)

	draftID, err := service.CreateDraft(toEmail, constants.DEFAULT_SUBJECT, bodyContent)
	if err != nil {
		return wppc.reply(ctx, "❌ Error al crear el borrador: "+err.Error())
	}

	wppc.pendingDrafts[sender] = draftID
	return wppc.reply(ctx, fmt.Sprintf("✅ Borrador creado para %s.\nDic: '%s'\n\nResponde SI para enviar.", toEmail, bodyContent))
}

func (wppc *WhatsappController) reply(ctx *fiber.Ctx, message string) error {
	ctx.Set("Content-Type", "text/xml")
	return ctx.SendString(fmt.Sprintf(models.ReplyMessage, message))
}
