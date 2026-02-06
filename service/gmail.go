package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"whatsapp-gmail-bot/auth"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var GmailService *gmail.Service

// InitGmail - Configura la conexión (Compatible con Nube y Local)
func InitGmail() {
	ctx := context.Background()

	credentials, err := loadResource("GOOGLE_CREDENTIALS", "resources/google_credentials.json")
	if err != nil {
		log.Fatalf("No se pudieron cargar las credenciales (ni ENV ni archivo): %s", err.Error())
	}

	config, err := google.ConfigFromJSON(credentials, gmail.GmailComposeScope, gmail.GmailModifyScope)
	if err != nil {
		log.Fatalf("Error config JSON: %s", err.Error())
	}

	client := auth.GetClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Error creando servicio Gmail: %s", err.Error())
	}

	GmailService = srv
	fmt.Println("Servicio de Gmail conectado y listo.")
}

// CreateDraft - Crea un borrador de email en Gmail
func CreateDraft(to, subject, body string) (string, error) {
	to = strings.TrimSpace(to)
	if !isValidEmail(to) {
		return "", fmt.Errorf("email inválido: %s", to)
	}

	contentType := "text/plain; charset=UTF-8"
	if containsHTML(body) {
		contentType = "text/html; charset=UTF-8"
	}

	messageString := buildMessageString(to, subject, body, contentType)
	msg := []byte(messageString)

	draftEmail := &gmail.Draft{
		Message: &gmail.Message{
			Raw: base64.URLEncoding.EncodeToString(msg),
		},
	}

	createdDraft, err := GmailService.Users.Drafts.Create("me", draftEmail).Do()
	if err != nil {
		return "", fmt.Errorf("error creando borrador: %s", err.Error())
	}

	return createdDraft.Id, nil
}

// buildMessageString - Construye el string del mensaje con los headers apropiados
func buildMessageString(to, subject, body, contentType string) string {
	return fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: %s\r\n"+
		"\r\n"+
		"%s", to, subject, contentType, body)
}

// containsHTML - Detecta si el texto contiene etiquetas HTML
func containsHTML(text string) bool {
	htmlTags := []string{"<img", "<html", "<body", "<div", "<p>", "<br", "<a ", "<span", "<h1", "<h2", "<h3", "<table", "<tr", "<td", "<th", "<ul", "<ol", "<li"}
	textLower := strings.ToLower(text)
	for _, tag := range htmlTags {
		if strings.Contains(textLower, tag) {
			return true
		}
	}
	return false
}

// isValidEmail - Valida formato básico de email
func isValidEmail(email string) bool {
	if email == "" || email == "PENDIENTE" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

// SendDraft - Envía el borrador por ID
func SendDraft(draftId string) error {
	draftToSend := &gmail.Draft{
		Id: draftId,
	}

	_, err := GmailService.Users.Drafts.Send("me", draftToSend).Do()
	if err != nil {
		return fmt.Errorf("error enviando email: %s", err.Error())
	}

	return nil
}

// loadResource - Intenta leer una Variable de Entorno. Si está vacía, lee el archivo local.
func loadResource(envName, fileName string) ([]byte, error) {
	envContent := os.Getenv(envName)
	if envContent != "" {
		fmt.Printf("Cargando %s desde Variable de Entorno\n", envName)
		return []byte(envContent), nil
	}

	fmt.Printf("Cargando %s desde Archivo Local: %s\n", envName, fileName)
	return os.ReadFile(fileName)
}
