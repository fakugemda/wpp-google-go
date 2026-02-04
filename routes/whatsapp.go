package routes

import (
	"fmt"
	"strings"
	"whatsapp-gmail-bot/service"

	"github.com/gofiber/fiber/v2"
)

const ONLY_SENDER = "whatsapp:+5493454127524"

var pendingDrafts = make(map[string]string)

func CreateFiberWhatsappRoutes(app *fiber.App) {
	app.Post("/whatsapp", WhatsappHandler)
}

func WhatsappHandler(ctx *fiber.Ctx) error {
	msg := ctx.FormValue("Body")
	sender := ctx.FormValue("From")

	if sender != ONLY_SENDER {
		fmt.Printf("⚠️ Intento de acceso no autorizado de: %s\n", sender)
		return reply(ctx, "⛔ Acceso denegado. Este bot es privado.")
	}

	fmt.Printf("📩 Mensaje de Facu: %s\n", msg)

	if strings.EqualFold(msg, "cancelar") {
		return handleCancelCommand(ctx, sender)
	}

	if strings.HasPrefix(strings.ToLower(msg), "mail a") {
		return handleMailCommand(ctx, msg, sender)
	}

	if strings.EqualFold(strings.TrimSpace(msg), "si") {
		return handleConfirmCommand(ctx, sender)
	}

	return reply(ctx, "🤖 No entendí. Comandos:\n- Mail a [email]: [mensaje]\n- Si (para confirmar)")
}

func handleCancelCommand(ctx *fiber.Ctx, sender string) error {
	delete(pendingDrafts, sender)
	return reply(ctx, "🗑️ Operación cancelada. Memoria limpia.")
}

func handleMailCommand(ctx *fiber.Ctx, msg, sender string) error {
	cleanMsg := msg[7:]
	parts := strings.SplitN(cleanMsg, ":", 2)

	if len(parts) < 2 {
		return reply(ctx, "⚠️ Formato incorrecto.\nUsa: Mail a correo@test.com: Hola mensaje")
	}

	toEmail := strings.TrimSpace(parts[0])
	bodyContent := strings.TrimSpace(parts[1])
	subject := "Mail enviado desde WhatsApp"

	draftID, err := service.CreateDraft(toEmail, subject, bodyContent)
	if err != nil {
		return reply(ctx, "❌ Error al crear el borrador: "+err.Error())
	}

	pendingDrafts[sender] = draftID
	return reply(ctx, fmt.Sprintf("✅ Borrador creado para %s.\nDic: '%s'\n\nResponde SI para enviar.", toEmail, bodyContent))
}

func handleConfirmCommand(ctx *fiber.Ctx, sender string) error {
	draftID, exists := pendingDrafts[sender]
	if !exists {
		return reply(ctx, "🤷‍♂️ No tienes borradores pendientes.")
	}

	err := service.SendDraft(draftID)
	if err != nil {
		return reply(ctx, "❌ Error al enviar: "+err.Error())
	}

	delete(pendingDrafts, sender)
	return reply(ctx, "🚀 ¡Correo enviado exitosamente!")
}

func reply(ctx *fiber.Ctx, message string) error {
	ctx.Set("Content-Type", "text/xml")
	return ctx.SendString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><Response><Message>%s</Message></Response>`, message))
}
