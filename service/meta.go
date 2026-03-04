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

// InteractiveButton - Representa un botón de respuesta rápida
type InteractiveButton struct {
	ID    string
	Title string
}

type ListRow struct {
	ID          string
	Title       string
	Description string
}

// SendInteractiveButtons - Envía un mensaje interactivo con botones de respuesta rápida
func SendInteractiveButtons(to, bodyText string, buttons []InteractiveButton) error {
	token := os.Getenv("META_TOKEN")
	phoneID := os.Getenv("META_PHONE_ID")

	to = utilsMeta.NormalizeArgPhoneNumber(to)

	if token == "" {
		return fmt.Errorf("META_TOKEN no está configurada")
	}
	if phoneID == "" {
		return fmt.Errorf("META_PHONE_ID no está configurada")
	}

	btnList := make([]map[string]interface{}, 0, len(buttons))
	for _, b := range buttons {
		btnList = append(btnList, map[string]interface{}{
			"type": "reply",
			"reply": map[string]string{
				"id":    b.ID,
				"title": b.Title,
			},
		})
	}

	url := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/messages", phoneID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "interactive",
		"interactive": map[string]interface{}{
			"type": "button",
			"body": map[string]string{"text": bodyText},
			"action": map[string]interface{}{
				"buttons": btnList,
			},
		},
	}

	jsonBody, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando a meta: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return fmt.Errorf("error de Facebook (%s): %s", resp.Status, buf.String())
	}

	return nil
}

// SendInteractiveList - Envía un mensaje interactivo tipo lista (hasta 10 filas en una sección)
func SendInteractiveList(to, bodyText, buttonLabel string, rows []ListRow) error {
	token := os.Getenv("META_TOKEN")
	phoneID := os.Getenv("META_PHONE_ID")

	to = utilsMeta.NormalizeArgPhoneNumber(to)

	if token == "" {
		return fmt.Errorf("META_TOKEN no está configurada")
	}
	if phoneID == "" {
		return fmt.Errorf("META_PHONE_ID no está configurada")
	}

	rowList := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		row := map[string]interface{}{
			"id":    r.ID,
			"title": truncate(r.Title, 24),
		}
		if r.Description != "" {
			row["description"] = truncate(r.Description, 72)
		}
		rowList = append(rowList, row)
	}

	url := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/messages", phoneID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              "interactive",
		"interactive": map[string]interface{}{
			"type": "list",
			"body": map[string]string{"text": bodyText},
			"action": map[string]interface{}{
				"button":   buttonLabel,
				"sections": []map[string]interface{}{{"rows": rowList}},
			},
		},
	}

	jsonBody, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando a meta: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return fmt.Errorf("error de Facebook (%s): %s", resp.Status, buf.String())
	}

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
