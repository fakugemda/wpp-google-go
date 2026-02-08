package config

import (
	"log"
	"net/http"
	"os"
	"wpp-quippos/routes"

	"github.com/joho/godotenv"
)

const (
	defaultPort        = "8080"
	defaultVerifyToken = "mi_token_secreto_123"
)

// InitApp inicializa toda la aplicación (configuración, rutas y servidor)
func InitApp() {
	// Cargar variables de entorno desde .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No se encontró archivo .env, usando valores por defecto")
	}

	// Obtener configuración desde variables de entorno
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	verifyToken := os.Getenv("VERIFY_TOKEN")
	if verifyToken == "" {
		verifyToken = defaultVerifyToken
		log.Printf("⚠️  ADVERTENCIA: Usando token de verificación por defecto. Define VERIFY_TOKEN en variables de entorno.")
	}

	// Crear multiplexor HTTP
	mux := http.NewServeMux()

	// Registrar rutas del webhook
	routes.RegisterWebhookRoutes(mux, verifyToken)

	// Logs de inicio
	log.Printf("🚀 Servidor iniciado en puerto %s", port)
	log.Printf("📡 Webhook endpoint: http://localhost:%s/webhook", port)
	log.Printf("🔑 Token de verificación: %s", verifyToken)
	log.Println("💡 Tip: Ejecuta 'ngrok http " + port + "' para exponer el servidor públicamente")
	log.Println("---------------------------------------------------")

	// Iniciar servidor
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("❌ Error al iniciar servidor: %v", err)
	}
}
