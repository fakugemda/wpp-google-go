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

type contactChoicePending struct {
	Data       *models.AIResponse
	Candidates []models.Contact
}

type WhatsappController struct {
	pendingDrafts        map[string]string
	pendingDraftData     map[string]*models.AIResponse
	pendingContactChoice map[string]contactChoicePending
	commands             map[string]CommandHandler
	contacts             []models.Contact
	contactsList         string
	metaWebhook          models.MetaWebhook
}

func NewWhatsappController(contacts []models.Contact, contactsList string) *WhatsappController {
	wppc := &WhatsappController{
		pendingDrafts:        make(map[string]string),
		pendingDraftData:     make(map[string]*models.AIResponse),
		pendingContactChoice: make(map[string]contactChoicePending),
		commands:             make(map[string]CommandHandler),
		contacts:             contacts,
		contactsList:         contactsList,
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
			} else if msgObj.Type == "interactive" && msgObj.Interactive.Type == "button_reply" {
				buttonID := msgObj.Interactive.ButtonReply.ID
				fmt.Printf("🔘 Botón presionado por %s: %s\n", sender, buttonID)
				if pending, ok := wppc.pendingContactChoice[sender]; ok {
					selectedEmail := buttonID
					dataCopy := *pending.Data
					dataCopy.To = selectedEmail
					delete(wppc.pendingContactChoice, sender)
					go wppc.createDraftAndSendConfirm(sender, &dataCopy)
				} else if buttonID == constants.BUTTON_CONFIRM_ID {
					go wppc.handleConfirm(sender, "")
				}
			} else if msgObj.Type == "interactive" && msgObj.Interactive.Type == "list_reply" {
				rowID := msgObj.Interactive.ListReply.ID
				fmt.Printf("📋 Lista elegida por %s: %s\n", sender, rowID)
				if pending, ok := wppc.pendingContactChoice[sender]; ok {
					selectedEmail := rowID
					dataCopy := *pending.Data
					dataCopy.To = selectedEmail
					delete(wppc.pendingContactChoice, sender)
					go wppc.createDraftAndSendConfirm(sender, &dataCopy)
				}
			} else if msgObj.Type == "text" {
				// CASO: ES UN MENSAJE DE TEXTO
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

	// Si hay un borrador pendiente, tratar el mensaje como corrección
	if existingDraft, hasDraft := wppc.pendingDraftData[sender]; hasDraft {
		fmt.Printf("✏️ Corrección de borrador de %s: %s\n", sender, msg)
		wppc.handleCorrection(sender, msg, existingDraft)
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

// handleCorrection - Corrige el borrador pendiente con la instrucción del usuario
func (wppc *WhatsappController) handleCorrection(sender, correctionMsg string, existingDraft *models.AIResponse) {
	wppc.reply(sender, "✏️ Corrigiendo el borrador...")

	correctedDraft, err := service.ProcessCorrection(existingDraft, correctionMsg)
	if err != nil {
		fmt.Printf("Error corrigiendo draft: %s\n", err.Error())
		wppc.reply(sender, "⚠️ No pude aplicar la corrección. Intenta de nuevo.")
		return
	}

	delete(wppc.pendingDrafts, sender)
	delete(wppc.pendingDraftData, sender)

	// Crear y mostrar el nuevo borrador corregido
	wppc.handleAiDraft(sender, correctedDraft)
}

// createDraftAndSendConfirm crea el borrador, lo guarda en pending y envía el preview con botón de confirmar.
func (wppc *WhatsappController) createDraftAndSendConfirm(sender string, data *models.AIResponse) {
	var attachmentData []byte
	var filename string
	if cachedImg, ok := service.GetImageWhatsApp(sender); ok {
		attachmentData = cachedImg.Bytes
		if cachedImg.Filename != "" {
			filename = cachedImg.Filename
		} else {
			filename = utilsController.GetFilenameFromMimeType(cachedImg.MimeType)
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
	wppc.pendingDraftData[sender] = data

	previewBody := service.BuildDraftPreviewMessage(data)
	confirmBtn := service.InteractiveButton{
		ID:    constants.BUTTON_CONFIRM_ID,
		Title: "✅ Sí, enviarlo",
	}
	if err := service.SendInteractiveButtons(sender, previewBody, []service.InteractiveButton{confirmBtn}); err != nil {
		fmt.Printf("❌ Error enviando botones: %v\n", err)
		wppc.reply(sender, previewBody+"\n\n_¿Lo envío? (Responde Sí)_")
	}
}

// handleAiDraft delega la resolución del destinatario al service y solo ejecuta la acción resultante.
func (wppc *WhatsappController) handleAiDraft(sender string, data *models.AIResponse) {
	fmt.Printf("IA generó: Subject=%s\n", data.Subject)

	result := service.ProcessDraftRecipient(data, wppc.contacts)
	if result.LogMessage != "" {
		fmt.Printf("  %s\n", result.LogMessage)
	}

	switch result.Action {
	case service.DraftActionCreateDraft:
		wppc.createDraftAndSendConfirm(sender, result.Data)
	case service.DraftActionAskEmail:
		wppc.reply(sender, result.Message)
	case service.DraftActionShowContactChoice:
		if result.TruncateMessage != "" {
			wppc.reply(sender, result.TruncateMessage)
		}
		wppc.pendingContactChoice[sender] = contactChoicePending{Data: result.Data, Candidates: result.Candidates}
		if result.UseButtons {
			if err := service.SendInteractiveButtons(sender, result.BodyText, result.Buttons); err != nil {
				fmt.Printf("❌ Error enviando botones de contacto: %v\n", err)
				wppc.reply(sender, result.BodyText+"\nEscribí el email de la persona.")
			}
		} else {
			if err := service.SendInteractiveList(sender, result.BodyText, "Ver contactos", result.Rows); err != nil {
				fmt.Printf("❌ Error enviando lista de contactos: %v\n", err)
				wppc.reply(sender, result.BodyText+"\nEscribí el email de la persona.")
			}
		}
	}
}

func (wppc *WhatsappController) handleCancel(sender, msg string) {
	delete(wppc.pendingDrafts, sender)
	delete(wppc.pendingDraftData, sender)
	delete(wppc.pendingContactChoice, sender)
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
	delete(wppc.pendingDraftData, sender)
	service.DeleteImageWhatsApp(sender)
	wppc.reply(sender, "🚀 Correo enviado exitosamente!")
}

func (wppc *WhatsappController) reply(to, message string) {
	err := service.SendWhatsAppMessage(to, message)
	if err != nil {
		fmt.Printf("❌ Error enviando WhatsApp a %s: %v\n", to, err)
	}
}

func (wppc *WhatsappController) createDraft(data *models.AIResponse) (string, error) {
	return service.CreateDraft(data.To, data.Subject, data.Content, nil, "")
}

func (wppc *WhatsappController) createDraftWithAttachment(data *models.AIResponse, attachmentData []byte, filename string) (string, error) {
	return service.CreateDraft(data.To, data.Subject, data.Content, attachmentData, filename)
}
