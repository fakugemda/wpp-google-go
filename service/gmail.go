package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"whatsapp-gmail-bot/auth"
	"whatsapp-gmail-bot/utils"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var GmailService *gmail.Service

var utilsGmail = utils.GetUtils()

// InitGmail - Configura la conexión (Compatible con Nube y Local)
func InitGmail() {
	ctx := context.Background()

	credentials, err := utilsGmail.LoadResource("GOOGLE_CREDENTIALS", "resources/google_credentials.json")
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
	if !utilsGmail.IsValidEmail(to) {
		return "", fmt.Errorf("email inválido: %s", to)
	}

	contentType := "text/plain; charset=UTF-8"
	if utilsGmail.ContainsHTML(body) {
		contentType = "text/html; charset=UTF-8"
	}

	messageString := utilsGmail.BuildEmailMessage(to, subject, body, contentType)
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
