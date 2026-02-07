package routes

import (
	"whatsapp-gmail-bot/controller"

	"github.com/gofiber/fiber/v2"
)

func RegisterWhatsappRoutes(app *fiber.App, contacts map[string]string, contactsList string) {
	// Inicializar controller con contactos inyectados
	wppController := controller.NewWhatsappController(contacts, contactsList)
	app.Get("/status", wppController.StatusController)

	// Meta Routes
	app.Get("/webhook", wppController.VerifyWebhook)
	app.Post("/webhook", wppController.ProcessWebhook)
}
