package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"wpp-quippos/models"
)

// ProcessWebhookPayload procesa el payload del webhook y extrae información relevante
func ProcessWebhookPayload(payload map[string]interface{}) (*models.WebhookPayload, error) {
	// Convertir el map a JSON y luego a estructura
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error al serializar payload: %v", err)
	}

	var webhookPayload models.WebhookPayload
	if err := json.Unmarshal(jsonData, &webhookPayload); err != nil {
		return nil, fmt.Errorf("error al deserializar payload: %v", err)
	}

	return &webhookPayload, nil
}

// LogWebhookEvent registra el evento del webhook de forma formateada
func LogWebhookEvent(payload map[string]interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Println("===================================================")
	log.Printf("📨 POST /webhook - Evento recibido [%s]", timestamp)
	log.Println("===================================================")
}

// ExtractMessageInfo extrae y registra información relevante del payload
func ExtractMessageInfo(payload *models.WebhookPayload) {
	if payload.Object != "" {
		log.Printf("📌 Tipo de objeto: %s", payload.Object)
	}

	if len(payload.Entry) > 0 {
		log.Printf("📌 Número de entradas: %d", len(payload.Entry))

		for _, entry := range payload.Entry {
			for _, change := range entry.Changes {
				// Mensajes recibidos
				if len(change.Value.Messages) > 0 {
					log.Printf("💬 Se recibieron %d mensaje(s)", len(change.Value.Messages))
					for _, msg := range change.Value.Messages {
						if msg.Type == "text" && msg.Text.Body != "" {
							log.Printf("   - De: %s", msg.From)
							log.Printf("   - Texto: %s", msg.Text.Body)
						}
					}
				}

				// Cambios de estado
				if len(change.Value.Statuses) > 0 {
					log.Printf("📊 Se recibieron %d actualización(es) de estado", len(change.Value.Statuses))
				}
			}
		}
	}

	log.Println("---------------------------------------------------")
}
