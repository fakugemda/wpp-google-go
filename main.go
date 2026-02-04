package main

import (
	"whatsapp-gmail-bot/routes"
	"whatsapp-gmail-bot/service"

	"github.com/gofiber/fiber/v2"
)

func main() {
	service.InitGmail()
	app := fiber.New()
	routes.CreateFiberWhatsappRoutes(app)
	app.Listen(":3000")
}
