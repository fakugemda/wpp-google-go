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
	// Cargar .env si existe (local); en Koyeb/cloud las variables vienen inyectadas en el entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró .env, usando variables de entorno del sistema")
	}

	// Cargar y validar todas las variables de entorno necesarias
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY no está configurada")
	}
	service.GeminiAPIKey = geminiAPIKey

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
