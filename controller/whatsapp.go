package controller

import (
	"fmt"
	"os"
	"strings"
	"time"
	"whatsapp-gmail-bot/constants"
	"whatsapp-gmail-bot/models"
	"whatsapp-gmail-bot/service"
	"whatsapp-gmail-bot/utils"

	"github.com/gofiber/fiber/v2"
)

type CommandHandler func(sender, msg string)

var utilsController = utils.GetUtils()

type WhatsappController struct {
	pendingDrafts map[string]string
	commands      map[string]CommandHandler
	contacts      map[string]string
	contactsList  string
	metaWebhook   models.MetaWebhook
}

func NewWhatsappController(contacts map[string]string, contactsList string) *WhatsappController {
	wppc := &WhatsappController{
		pendingDrafts: make(map[string]string),
		commands:      make(map[string]CommandHandler),
		contacts:      contacts,
		contactsList:  contactsList,
	}
	wppc.registerCommands()
	service.StartImageCacheCleanup()
	return wppc
}

func (wppc *WhatsappController) registerCommands() {
	wppc.commands[constants.COMMAND_CANCEL] = wppc.handleCancel
	wppc.commands[constants.COMMAND_CONFIRM] = wppc.handleConfirm
}

func (wppc *WhatsappController) StatusController(ctx *fiber.Ctx) error {
	fmt.Println("🟢 Funcionando correctamente... 🟢")
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"status": "running"})
}

func (wppc *WhatsappController) VerifyWebhook(ctx *fiber.Ctx) error {
	verifyToken := os.Getenv("META_VERIFY_TOKEN")
	mode := ctx.Query("hub.mode")
	token := ctx.Query("hub.verify_token")
	challenge := ctx.Query("hub.challenge")

	if mode == "subscribe" && token == verifyToken {
		return ctx.SendString(challenge)
	}
	return ctx.SendStatus(403)
}

func (wppc *WhatsappController) ProcessWebhook(ctx *fiber.Ctx) error {
	var payload models.MetaWebhook
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.SendStatus(400)
	}

	if len(payload.Entry) > 0 && len(payload.Entry[0].Changes) > 0 {
		change := payload.Entry[0].Changes[0].Value
		if len(change.Messages) > 0 {
			msgObj := change.Messages[0]
			sender := msgObj.From

			if msgObj.Type == "image" {
				fmt.Printf("📷 Imagen recibida de %s, ID: %s\n", sender, msgObj.Image.ID)

				imgBytes, mimeType, err := service.DownloadMedia(msgObj.Image.ID)
				if err != nil {
					fmt.Printf("❌ Error descargando imagen: %s\n", err.Error())
					wppc.reply(sender, "⚠️ Error al procesar la imagen. Intenta de nuevo.")
					return ctx.SendStatus(200)
				}

				// Guardar en caché global con bytes, MIME type y timestamp
				service.SetImageWhatsApp(sender, &service.CachedImage{
					Bytes:     imgBytes,
					MimeType:  mimeType,
					Filename:  "",
					CreatedAt: time.Now(),
				})
				fmt.Printf("✅ Imagen descargada y guardada en caché (tipo: %s, tamaño: %d bytes)\n", mimeType, len(imgBytes))

				// Si tiene caption, lo usamos como prompt
				if msgObj.Image.Caption != "" {
					fmt.Printf("📝 Caption recibido: %s\n", msgObj.Image.Caption)
					go wppc.handleCommand(sender, msgObj.Image.Caption)
				} else {
					// Si no tiene caption, pedimos instrucciones
					wppc.reply(sender, "📷 Foto recibida. ¿Qué quieres que haga con ella? (Ej: 'Mándasela a mamá')")
				}
			} else if msgObj.Type == "document" {
				fmt.Printf("📄 Documento recibido de %s, ID: %s, Nombre: %s\n", sender, msgObj.Document.ID, msgObj.Document.Filename)

				// Descargar el documento
				docBytes, mimeType, err := service.DownloadMedia(msgObj.Document.ID)
				if err != nil {
					fmt.Printf("❌ Error descargando documento: %s\n", err.Error())
					wppc.reply(sender, "⚠️ Error al procesar el documento. Intenta de nuevo.")
					return ctx.SendStatus(200)
				}

				// Usar MIME type del documento si está disponible, sino el detectado
				if msgObj.Document.MimeType != "" {
					mimeType = service.NormalizeMimeType(msgObj.Document.MimeType)
				}

				// Guardar en caché global con nombre de archivo real
				service.SetImageWhatsApp(sender, &service.CachedImage{
					Bytes:     docBytes,
					MimeType:  mimeType,
					Filename:  msgObj.Document.Filename,
					CreatedAt: time.Now(),
				})
				fmt.Printf("✅ Documento descargado y guardado en caché (tipo: %s, tamaño: %d bytes)\n", mimeType, len(docBytes))

				// Si tiene caption, lo usamos como prompt
				if msgObj.Document.Caption != "" {
					fmt.Printf("📝 Caption recibido: %s\n", msgObj.Document.Caption)
					go wppc.handleCommand(sender, msgObj.Document.Caption)
				} else {
					// Si no tiene caption, pedimos instrucciones
					filename := msgObj.Document.Filename
					if filename == "" {
						filename = "el archivo"
					}
					wppc.reply(sender, fmt.Sprintf("📄 Recibí el archivo '%s'. ¿Qué quieres que haga con él? (Ej: 'Mándaselo a Juan')", filename))
				}
			} else if msgObj.Type == "text" {
				// CASO 2: ES UN MENSAJE DE TEXTO
				text := msgObj.Text.Body
				fmt.Printf("📩 Mensaje de %s: %s\n", sender, text)
				go wppc.handleCommand(sender, text)
			} else {
				fmt.Printf("⚠️ Tipo de mensaje no soportado: %s\n", msgObj.Type)
			}
		}
	}

	return ctx.SendStatus(200)
}

func (wppc *WhatsappController) handleCommand(sender, msg string) {
	msgLower := strings.ToLower(strings.TrimSpace(msg))

	if handler, exists := wppc.commands[msgLower]; exists {
		handler(sender, msg)
		return
	}

	fmt.Println("🧠 Consultando a Gemini...")
	aiDecision, err := service.ProcessIntent(msg, wppc.contactsList)

	if err != nil {
		fmt.Printf("Error IA: %s\n", err.Error())
		wppc.reply(sender, "⚠️ La IA tuvo un error. Intenta de nuevo.")
		return
	}

	switch aiDecision.Type {
	case "CHAT":
		wppc.reply(sender, aiDecision.Content)
	case "CONFIRM_SEND":
		wppc.handleConfirm(sender, msg)
	case "EMAIL_DRAFT":
		wppc.handleAiDraft(sender, aiDecision)
	default:
		wppc.reply(sender, "🤖 No entendí la respuesta de la IA.")
	}
}

// handleAiDraft
func (wppc *WhatsappController) handleAiDraft(sender string, data *models.AIResponse) {
	fmt.Printf("IA generó: To=%s, Subject=%s\n", data.To, data.Subject)

	if !wppc.validateAndResolveRecipient(data) {
		wppc.reply(sender, fmt.Sprintf("Entendido: '%s'\n\nPero... ¿A quién se lo mando? (Dime el email)", data.Subject))
		return
	}

	// Verificar si hay imagen en caché para este usuario
	var attachmentData []byte
	var filename string
	if cachedImg, ok := service.GetImageWhatsApp(sender); ok {
		attachmentData = cachedImg.Bytes
		if cachedImg.Filename != "" {
			filename = cachedImg.Filename
		} else {
			filename = wppc.getFilenameFromMimeType(cachedImg.MimeType)
		}
		fmt.Printf("📎 Incluyendo imagen en el borrador (tipo: %s, tamaño: %d bytes)\n", cachedImg.MimeType, len(attachmentData))
	}

	draftID, err := wppc.createDraftWithAttachment(data, attachmentData, filename)
	if err != nil {
		fmt.Printf("Error creando draft: %s\n", err.Error())
		wppc.reply(sender, fmt.Sprintf("Error al crear borrador: %s", err.Error()))
		return
	}

	wppc.pendingDrafts[sender] = draftID
	wppc.reply(sender, wppc.buildPreviewMessage(data))
}

func (wppc *WhatsappController) handleCancel(sender, msg string) {
	delete(wppc.pendingDrafts, sender)
	service.DeleteImageWhatsApp(sender)
	wppc.reply(sender, "🗑️ Operación cancelada. Memoria limpia.")
}

func (wppc *WhatsappController) handleConfirm(sender, msg string) {
	draftID, exists := wppc.pendingDrafts[sender]
	if !exists {
		wppc.reply(sender, "🤷‍♂️ No tienes borradores pendientes.")
		return
	}

	if err := service.SendDraft(draftID); err != nil {
		wppc.reply(sender, fmt.Sprintf("❌ Error al enviar: %s", err.Error()))
		return
	}

	delete(wppc.pendingDrafts, sender)
	service.DeleteImageWhatsApp(sender)
	wppc.reply(sender, "🚀 Correo enviado exitosamente!")
}

func (wppc *WhatsappController) reply(to, message string) {
	err := service.SendWhatsAppMessage(to, message)
	if err != nil {
		fmt.Printf("❌ Error enviando WhatsApp a %s: %v\n", to, err)
	}
}

func (wppc *WhatsappController) validateAndResolveRecipient(data *models.AIResponse) bool {
	if data.To == "PENDIENTE" || data.To == "" {
		return false
	}
	data.To = strings.TrimSpace(data.To)
	if !utilsController.IsValidEmail(data.To) {
		if email, found := utilsController.ResolveContactByName(wppc.contacts, data.To); found {
			data.To = email
		} else {
			return false
		}
	}
	return true
}

func (wppc *WhatsappController) createDraft(data *models.AIResponse) (string, error) {
	return service.CreateDraft(data.To, data.Subject, data.Content, nil, "")
}

func (wppc *WhatsappController) createDraftWithAttachment(data *models.AIResponse, attachmentData []byte, filename string) (string, error) {
	return service.CreateDraft(data.To, data.Subject, data.Content, attachmentData, filename)
}

// getFilenameFromMimeType convierte un MIME type a un nombre de archivo con extensión apropiada
func (wppc *WhatsappController) getFilenameFromMimeType(mimeType string) string {
	mimeToExt := map[string]string{
		"image/jpeg":      ".jpg",
		"image/jpg":       ".jpg",
		"image/png":       ".png",
		"image/gif":       ".gif",
		"image/webp":      ".webp",
		"image/bmp":       ".bmp",
		"image/tiff":      ".tiff",
		"image/svg+xml":   ".svg",
		"application/pdf": ".pdf",
	}

	if ext, ok := mimeToExt[strings.ToLower(mimeType)]; ok {
		return "foto_whatsapp" + ext
	}

	return "foto_whatsapp.jpg"
}

func (wppc *WhatsappController) buildPreviewMessage(data *models.AIResponse) string {
	return fmt.Sprintf("*Borrador IA Creado*\n\n*Para:* %s\n*Asunto:* %s\n\n%s\n\n_¿Lo envío? (Responde Sí)_",
		data.To, data.Subject, data.Content)
}
