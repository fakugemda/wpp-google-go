package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const whatsappAPIURL = "https://graph.facebook.com/v22.0"

// SendMessageRequest estructura para enviar mensajes a WhatsApp
type SendMessageRequest struct {
	MessagingProduct string      `json:"messaging_product"`
	RecipientType    string      `json:"recipient_type"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Text             TextMessage `json:"text"`
}

// TextMessage contenido del mensaje de texto
type TextMessage struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}

// normalizePhoneNumber ajusta el número de teléfono al formato correcto para la API
func normalizePhoneNumber(phone string) string {
	// Eliminar espacios en blanco
	phone = strings.TrimSpace(phone)

	// Eliminar el prefijo '+' si existe
	phone = strings.TrimPrefix(phone, "+")

	// TODO
	// Recibido: 549345xxxxxxx
	// Necesario para enviar: 54345xxxxxxx
	if strings.HasPrefix(phone, "549") && len(phone) >= 13 {
		// Remover el "9" extra: 549 -> 54
		phone = "54" + phone[3:]
		log.Printf("📱 Número argentino normalizado: %s", phone)
	}

	return phone
}

// SendWhatsAppMessage envía un mensaje de texto a través de la API de WhatsApp
func SendWhatsAppMessage(to, message string) error {
	// Normalizar el número de teléfono al formato internacional
	to = normalizePhoneNumber(to)
	log.Printf("📞 Enviando mensaje a: %s", to)
	accessToken := os.Getenv("WHATSAPP_ACCESS_TOKEN")
	phoneNumberID := os.Getenv("PHONE_NUMBER_ID")

	if accessToken == "" {
		return fmt.Errorf("WHATSAPP_ACCESS_TOKEN no está configurado")
	}
	if phoneNumberID == "" {
		return fmt.Errorf("PHONE_NUMBER_ID no está configurado")
	}

	// Construir el payload
	payload := SendMessageRequest{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
		Text: TextMessage{
			PreviewURL: false,
			Body:       message,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error al serializar payload: %v", err)
	}

	// Construir la URL
	url := fmt.Sprintf("%s/%s/messages", whatsappAPIURL, phoneNumberID)

	// Crear la petición HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Enviar la petición
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando mensaje: %v", err)
	}
	defer resp.Body.Close()

	// Verificar respuesta
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("error de WhatsApp API (status %d): %v", resp.StatusCode, errorResp)
	}

	log.Printf("✅ Mensaje enviado exitosamente a %s", to)
	return nil
}

// Language estructura para el idioma del template
type Language struct {
	Code string `json:"code"`
}

// Template estructura para el template de WhatsApp
type Template struct {
	Name     string   `json:"name"`
	Language Language `json:"language"`
}

// SendTemplateRequest estructura para enviar templates a WhatsApp
type SendTemplateRequest struct {
	MessagingProduct string   `json:"messaging_product"`
	To               string   `json:"to"`
	Type             string   `json:"type"`
	Template         Template `json:"template"`
}

// SendTemplateMessage envía un mensaje de template a través de la API de WhatsApp
func SendTemplateMessage(to, templateName, languageCode string) error {
	// Normalizar el número de teléfono al formato internacional
	to = normalizePhoneNumber(to)
	log.Printf("📞 Enviando template '%s' a: %s", templateName, to)

	accessToken := os.Getenv("WHATSAPP_ACCESS_TOKEN")
	phoneNumberID := os.Getenv("PHONE_NUMBER_ID")

	if accessToken == "" {
		return fmt.Errorf("WHATSAPP_ACCESS_TOKEN no está configurado")
	}
	if phoneNumberID == "" {
		return fmt.Errorf("PHONE_NUMBER_ID no está configurado")
	}

	// Construir el payload del template
	payload := SendTemplateRequest{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "template",
		Template: Template{
			Name: templateName,
			Language: Language{
				Code: languageCode,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error al serializar payload: %v", err)
	}

	// Construir la URL
	url := fmt.Sprintf("%s/%s/messages", whatsappAPIURL, phoneNumberID)

	// Crear la petición HTTP
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creando request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Enviar la petición
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error enviando template: %v", err)
	}
	defer resp.Body.Close()

	// Verificar respuesta
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return fmt.Errorf("error de WhatsApp API (status %d): %v", resp.StatusCode, errorResp)
	}

	log.Printf("✅ Template enviado exitosamente a %s", to)
	return nil
}
