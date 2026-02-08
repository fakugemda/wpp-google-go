package routes

import (
	"net/http"
	"wpp-quippos/controller"
)

var webhookController *controller.WebhookController

// RegisterWebhookRoutes registra todas las rutas del webhook
func RegisterWebhookRoutes(mux *http.ServeMux, verifyToken string) {
	// Inicializar controller con token de verificación
	webhookController = controller.NewWebhookController(verifyToken)

	// Registrar rutas
	mux.HandleFunc("/webhook", handleWebhook)
	mux.HandleFunc("/", webhookController.HomeHandler)
}

// handleWebhook enruta las peticiones GET y POST al controller apropiado
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Endpoint de verificación del webhook (requerido por Meta)
		webhookController.VerifyWebhook(w, r)
	case http.MethodPost:
		// Endpoint para recibir mensajes y eventos
		webhookController.ReceiveWebhook(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}
