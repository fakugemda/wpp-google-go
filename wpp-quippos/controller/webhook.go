package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"wpp-quippos/models"
	"wpp-quippos/service"
)

// WebhookController maneja la lógica de negocio del webhook
type WebhookController struct {
	verifyToken string
}

// NewWebhookController crea una nueva instancia del controller
func NewWebhookController(verifyToken string) *WebhookController {
	return &WebhookController{
		verifyToken: verifyToken,
	}
}

// VerifyWebhook maneja la verificación del webhook por parte de Meta (GET)
func (wc *WebhookController) VerifyWebhook(w http.ResponseWriter, r *http.Request) {
	// Meta envía estos parámetros para verificar el webhook
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	log.Printf("📥 GET /webhook - Solicitud de verificación recibida")
	log.Printf("   Mode: %s", mode)
	log.Printf("   Token recibido: %s", token)
	log.Printf("   Challenge: %s", challenge)

	// Verificar que el modo sea "subscribe" y el token coincida
	if mode == "subscribe" && token == wc.verifyToken {
		log.Println("✅ Verificación exitosa - Webhook confirmado")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, challenge)
		return
	}

	log.Println("❌ Verificación fallida - Token incorrecto o modo inválido")
	http.Error(w, "Verificación fallida", http.StatusForbidden)
}

// ReceiveWebhook maneja los eventos y mensajes recibidos de WhatsApp (POST)
func (wc *WebhookController) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	// Leer el body completo
	var payload map[string]interface{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&payload); err != nil {
		log.Printf("❌ Error al decodificar JSON: %v", err)
		http.Error(w, "Error al leer datos", http.StatusBadRequest)
		return
	}

	// Registrar el evento completo
	service.LogWebhookEvent(payload)

	// Procesar y extraer información
	webhookPayload, err := service.ProcessWebhookPayload(payload)
	if err != nil {
		log.Printf("⚠️  Error al procesar payload: %v", err)
	} else {
		service.ExtractMessageInfo(webhookPayload)

		// Procesar mensajes de texto con IA
		wc.processIncomingMessages(webhookPayload)
	}

	// Responder con 200 OK para confirmar recepción
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// processIncomingMessages procesa mensajes entrantes y genera respuestas con IA
func (wc *WebhookController) processIncomingMessages(payload *models.WebhookPayload) {
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			// Solo procesar si hay mensajes
			if len(change.Value.Messages) == 0 {
				continue
			}

			// Obtener nombre del contacto
			senderName := "Usuario"
			if len(change.Value.Contacts) > 0 {
				senderName = change.Value.Contacts[0].Profile.Name
			}

			// Procesar cada mensaje
			for _, msg := range change.Value.Messages {
				// Solo procesar mensajes de texto
				if msg.Type != "text" {
					continue
				}

				go wc.handleTextMessage(msg, senderName)
			}
		}
	}
}

// handleTextMessage procesa un mensaje de texto individual con IA
func (wc *WebhookController) handleTextMessage(msg models.Message, senderName string) {
	log.Printf("💬 Procesando mensaje de %s: %s", senderName, msg.Text.Body)

	// Generar respuesta con IA
	response, err := service.ProcessMessage(msg.Text.Body, senderName, msg.From)
	if err != nil {
		log.Printf("❌ Error procesando con IA: %v", err)
		return
	}

	// Enviar respuesta
	err = service.SendWhatsAppMessage(msg.From, response)
	if err != nil {
		log.Printf("❌ Error enviando respuesta: %v", err)
		return
	}

	log.Printf("✅ Respuesta enviada a %s", senderName)
}

// HomeHandler maneja la ruta raíz
func (wc *WebhookController) HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "WhatsApp Webhook Server - Use /webhook endpoint")
}
