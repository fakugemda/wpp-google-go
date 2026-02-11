package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
	"whatsapp-gmail-bot/utils"
)

var utilsMeta = utils.GetUtils()

// SendWhatsAppMessage - Envía texto a un usuario usando la Cloud API de Meta
func SendWhatsAppMessage(to string, message string) error {
	token := os.Getenv("META_TOKEN")
	phoneID := os.Getenv("META_PHONE_ID")

	to = utilsMeta.NormalizeArgPhoneNumber(to)

	if token == "" {
		return fmt.Errorf("META_TOKEN no está configurada")
	}
	if phoneID == "" {
		return fmt.Errorf("META_PHONE_ID no está configurada")
	}

	url := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/messages", phoneID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text": map[string]string{
			"body": message,
		},
	}

	jsonBody, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando a meta: %v", err)
	}
	defer resp.Body.Close()

	// Verificar respuesta de Facebook
	if resp.StatusCode >= 400 {
		// Leemos el cuerpo del error para saber qué pasó
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return fmt.Errorf("error de Facebook (%s): %s", resp.Status, buf.String())
	}

	return nil
}
