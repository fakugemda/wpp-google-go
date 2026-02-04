package routes

import (
	"whatsapp-gmail-bot/controller"

	"github.com/gofiber/fiber/v2"
)

var WppController = controller.NewWhatsappController()

func RegisterWhatsappRoutes(app *fiber.App) {

	app.Post("/whatsapp", WppController.SendWhatsappController)
}
