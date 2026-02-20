package config

import (
	"log"
	"os"
	"whatsapp-gmail-bot/discord"
	"whatsapp-gmail-bot/routes"
	"whatsapp-gmail-bot/service"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// InitApp - Inicializa toda la aplicación
func InitApp() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error al cargar variables de entorno")
	}
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
