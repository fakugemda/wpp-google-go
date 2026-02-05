package routes

import (
	"whatsapp-gmail-bot/controller"

	"github.com/gofiber/fiber/v2"
)

var WppController *controller.WhatsappController

func RegisterWhatsappRoutes(app *fiber.App, contacts map[string]string, contactsList string) {
	// Inicializar controller con contactos inyectados
	WppController = controller.NewWhatsappController(contacts, contactsList)
	app.Post("/whatsapp", WppController.SendWhatsappController)
}
