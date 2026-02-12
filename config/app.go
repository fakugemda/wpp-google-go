package config

import (
	"log"
	"os"
	"whatsapp-gmail-bot/discord"
	"whatsapp-gmail-bot/routes"
	"whatsapp-gmail-bot/service"

	"github.com/gofiber/fiber/v2"
)

// InitApp - Inicializa toda la aplicación
func InitApp() {
	service.InitGmail()
	LoadContacts()
	go discord.StartService(GetContactsPrompt())
	app := fiber.New()
	routes.RegisterWhatsappRoutes(app, Contacts, GetContactsPrompt())
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf(" Servidor iniciado en puerto %s", port)
	log.Fatal(app.Listen(":" + port))
}
