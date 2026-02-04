package config

import (
	"log"
	"os"
	"whatsapp-gmail-bot/routes"
	"whatsapp-gmail-bot/service"

	"github.com/gofiber/fiber/v2"
)

// InitApp: Inicializa toda la aplicación (Gmail, Fiber, Rutas y Servidor)
func InitApp() {
	service.InitGmail()
	app := fiber.New()
	routes.RegisterWhatsappRoutes(app)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("🚀 Servidor iniciado en puerto %s", port)
	log.Fatal(app.Listen(":" + port))
}
