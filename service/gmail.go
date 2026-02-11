package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"path/filepath"
	"strings"
	"whatsapp-gmail-bot/auth"
	"whatsapp-gmail-bot/utils"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// encodeSubject codifica el asunto en formato RFC 2047 para soportar UTF-8

func encodeSubject(subject string) string {
	if subject == "" {
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(subject))
	return fmt.Sprintf("=?utf-8?B?%s?=", encoded)
}

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
// Blindado contra problemas UTF-8 (tildes, eñes, emojis)
func CreateDraft(to, subject, body string, attachmentData []byte, filename string) (string, error) {
	to = strings.TrimSpace(to)
	if !utilsGmail.IsValidEmail(to) {
		return "", fmt.Errorf("email inválido: %s", to)
	}

	// Codificar el asunto para soportar UTF-8 (tildes, eñes, emojis)
	encodedSubject := encodeSubject(subject)

	var msg []byte

	// Si NO hay adjunto, mandamos el mail simple de siempre
	if len(attachmentData) == 0 {
		contentType := "text/plain; charset=UTF-8"
		if utilsGmail.ContainsHTML(body) {
			contentType = "text/html; charset=UTF-8"
		}

		// Construir mensaje con asunto codificado
		messageString := fmt.Sprintf("To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: %s\r\n"+
			"\r\n"+
			"%s", to, encodedSubject, contentType, body)
		msg = []byte(messageString)
	} else {
		// Si HAY adjunto, construimos el MIME Multipart
		boundary := "----=_Part_0_" + fmt.Sprintf("%d", len(attachmentData))

		// Detectar tipo de archivo (jpeg, png, etc)
		mimeType := mime.TypeByExtension(filepath.Ext(filename))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		// Codificar archivo a Base64
		fileBase64 := base64.StdEncoding.EncodeToString(attachmentData)

		// Determinar content type del cuerpo
		bodyContentType := "text/plain; charset=UTF-8"
		if utilsGmail.ContainsHTML(body) {
			bodyContentType = "text/html; charset=UTF-8"
		}

		// Codificar el cuerpo en Base64 para blindaje UTF-8
		bodyBase64 := base64.StdEncoding.EncodeToString([]byte(body))

		// Construcción manual del Email Multipart con blindaje UTF-8
		msgParts := []string{
			fmt.Sprintf("To: %s", to),
			fmt.Sprintf("Subject: %s", encodedSubject), // Asunto blindado
			"MIME-Version: 1.0",
			fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"", boundary),
			"",
			fmt.Sprintf("--%s", boundary),
			fmt.Sprintf("Content-Type: %s", bodyContentType),
			"Content-Transfer-Encoding: base64", // Codificación Base64 para el cuerpo
			"",
			bodyBase64, // Cuerpo codificado en Base64
			"",
			fmt.Sprintf("--%s", boundary),
			fmt.Sprintf("Content-Type: %s; name=\"%s\"", mimeType, filename),
			"Content-Transfer-Encoding: base64",
			fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"", filename),
			"",
			fileBase64,
			fmt.Sprintf("--%s--", boundary),
		}

		fullMsg := strings.Join(msgParts, "\r\n")
		msg = []byte(fullMsg)
	}

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
