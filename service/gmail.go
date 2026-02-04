package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"whatsapp-gmail-bot/auth" // Asegúrate de que este path sea correcto en tu go.mod

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var GmailService *gmail.Service

// InitGmail: Configura la conexión (Compatible con Nube y Local)
func InitGmail() {
	ctx := context.Background()

	// 1. Intentamos cargar las credenciales usando nuestra función "anfibia"
	// Busca la variable "GOOGLE_CREDENTIALS". Si no está, busca el archivo.
	credentials, err := loadResource("GOOGLE_CREDENTIALS", "resources/google_credentials.json")
	if err != nil {
		log.Fatalf("❌ No se pudieron cargar las credenciales (ni ENV ni archivo): %v", err)
	}

	config, err := google.ConfigFromJSON(credentials, gmail.GmailComposeScope, gmail.GmailModifyScope)
	if err != nil {
		log.Fatalf("Error config JSON: %v", err)
	}

	// 2. Pasamos el control al paquete Auth para obtener el cliente
	client := auth.GetClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Error creando servicio Gmail: %v", err)
	}

	GmailService = srv
	fmt.Println("✅ Servicio de Gmail conectado y listo.")
}

func CreateDraft(to, subject, body string) (string, error) {
	messageString := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s", to, subject, body)

	msg := []byte(messageString)

	gmailMessage := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString(msg),
	}

	draftEmail := &gmail.Draft{
		Message: gmailMessage,
	}

	createdDraft, err := GmailService.Users.Drafts.Create("me", draftEmail).Do()
	if err != nil {
		return "", fmt.Errorf("error creando borrador: %v", err)
	}

	return createdDraft.Id, nil
}

// SendDraft: Envía el borrador por ID
func SendDraft(draftId string) error {
	draftToSend := &gmail.Draft{
		Id: draftId,
	}

	_, err := GmailService.Users.Drafts.Send("me", draftToSend).Do()
	if err != nil {
		return fmt.Errorf("error enviando email: %v", err)
	}

	return nil
}

// loadResource: Intenta leer una Variable de Entorno. Si está vacía, lee el archivo local.
func loadResource(envName, fileName string) ([]byte, error) {
	envContent := os.Getenv(envName)
	if envContent != "" {
		fmt.Printf("☁️ Cargando %s desde Variable de Entorno\n", envName)
		return []byte(envContent), nil
	}

	fmt.Printf("🏠 Cargando %s desde Archivo Local: %s\n", envName, fileName)
	return os.ReadFile(fileName)
}
