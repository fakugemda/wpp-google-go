package controller

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
	"whatsapp-gmail-bot/constants"
	"whatsapp-gmail-bot/models"
	"whatsapp-gmail-bot/service"
	"whatsapp-gmail-bot/store"
	"whatsapp-gmail-bot/utils"

	"github.com/gofiber/fiber/v2"
)

type CommandHandler func(sender, msg string)

var utilsController = utils.GetUtils()

type NewContactState struct {
	Stage    string
	TempName string
}

type WhatsappController struct {
	pendingDrafts       map[string]string
	pendingDraftData    map[string]*models.AIResponse // borrador actual para correcciones
	pendingNewRecipient map[string]*models.AIResponse // borrador IA sin destinatario resuelto
	newContacts         map[string]*NewContactState   // flujo de alta de nuevo contacto
	commands            map[string]CommandHandler
	contacts            map[string]string
	contactsList        string
	metaWebhook         models.MetaWebhook
}

func NewWhatsappController(contacts map[string]string, contactsList string) *WhatsappController {
	wppc := &WhatsappController{
		pendingDrafts:       make(map[string]string),
		pendingDraftData:    make(map[string]*models.AIResponse),
		pendingNewRecipient: make(map[string]*models.AIResponse),
		newContacts:         make(map[string]*NewContactState),
		commands:            make(map[string]CommandHandler),
		contacts:            contacts,
		contactsList:        contactsList,
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
				switch buttonID {
				case constants.BUTTON_CONFIRM_ID:
					go wppc.handleConfirm(sender, "")
				case constants.BUTTON_ADD_CONTACT_YES_ID:
					wppc.startAddContactName(sender)
				case constants.BUTTON_ADD_CONTACT_NO_ID:
					wppc.cancelAddContactFlow(sender)
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

	// Flujo de alta de nuevo contacto (tiene prioridad sobre comandos normales)
	if state, exists := wppc.newContacts[sender]; exists {
		switch state.Stage {
		case "awaiting_name":
			wppc.handleNewContactName(sender, msg)
			return
		case "awaiting_email":
			wppc.handleNewContactEmail(sender, msg, state)
			return
		}
	}

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

// handleAiDraft
func (wppc *WhatsappController) handleAiDraft(sender string, data *models.AIResponse) {
	fmt.Printf("IA generó: To=%s, Subject=%s\n", data.To, data.Subject)

	ok, unresolved := wppc.validateAndResolveRecipient(data)
	if !ok {
		// Si la IA devolvió un nombre que no existe en contactos, ofrecer guardarlo
		if strings.TrimSpace(unresolved) != "" {
			wppc.pendingNewRecipient[sender] = data
			wppc.newContacts[sender] = &NewContactState{
				Stage:    "awaiting_confirm",
				TempName: unresolved,
			}
			body := fmt.Sprintf("No tengo a \"%s\" en tu agenda.\n\n¿Quieres guardarlo como contacto para futuros mails?", unresolved)
			yesBtn := service.InteractiveButton{
				ID:    constants.BUTTON_ADD_CONTACT_YES_ID,
				Title: "Sí, agregar contacto", // máx 20 caracteres (WhatsApp)
			}
			noBtn := service.InteractiveButton{
				ID:    constants.BUTTON_ADD_CONTACT_NO_ID,
				Title: "No, solo esta vez",
			}
			if err := service.SendInteractiveButtons(sender, body, []service.InteractiveButton{yesBtn, noBtn}); err != nil {
				fmt.Printf("❌ Error enviando botones alta contacto: %v\n", err)
				wppc.reply(sender, body+"\n\nResponde con el email para enviar solo esta vez.")
			}
			return
		}

		// Caso original: To vacío o pendiente
		wppc.reply(sender, fmt.Sprintf("Entendido: '%s'\n\nPero... ¿A quién se lo mando? (Dime el email)", data.Subject))
		return
	}

	wppc.finalizeDraftWithRecipient(sender, data)
}

// finalizeDraftWithRecipient asume que data.To ya es un email válido
func (wppc *WhatsappController) finalizeDraftWithRecipient(sender string, data *models.AIResponse) {
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
	wppc.pendingDraftData[sender] = data // guardar para posibles correcciones

	// Enviar preview con botón interactivo de confirmación
	previewBody := wppc.buildPreviewMessage(data)
	confirmBtn := service.InteractiveButton{
		ID:    constants.BUTTON_CONFIRM_ID,
		Title: "✅ Sí, enviarlo",
	}
	if err := service.SendInteractiveButtons(sender, previewBody, []service.InteractiveButton{confirmBtn}); err != nil {
		fmt.Printf("❌ Error enviando botones: %v\n", err)
		// Fallback: texto plano con instrucciones
		wppc.reply(sender, previewBody+"\n\n_¿Lo envío? (Responde Sí)_")
	}
}

func (wppc *WhatsappController) handleCancel(sender, msg string) {
	delete(wppc.pendingDrafts, sender)
	delete(wppc.pendingDraftData, sender)
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

func (wppc *WhatsappController) validateAndResolveRecipient(data *models.AIResponse) (bool, string) {
	if data.To == "PENDIENTE" || data.To == "" {
		return false, ""
	}
	data.To = strings.TrimSpace(data.To)
	resolved, unresolved, ok := wppc.resolveRecipientsList(data.To)
	if !ok {
		return false, unresolved
	}

	data.To = strings.Join(resolved, ", ")
	return true, ""
}

// resolveRecipientsList se encarga de procesar una cadena de destinatarios
// potencialmente múltiples (separados por comas / "y" / "e")
func (wppc *WhatsappController) resolveRecipientsList(raw string) ([]string, string, bool) {
	parts := utilsController.SplitRecipients(raw)
	if len(parts) == 0 {
		return nil, "", false
	}

	var resolved []string

	for _, p := range parts {
		if utilsController.IsValidEmail(p) {
			resolved = append(resolved, p)
			continue
		}

		// Intentar resolver contra la agenda por nombre/apodo
		if email, found := utilsController.ResolveContactByName(wppc.contacts, p); found {
			resolved = append(resolved, email)
			continue
		}

		return nil, p, false
	}

	if len(resolved) == 0 {
		return nil, "", false
	}

	return resolved, "", true
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
	return fmt.Sprintf("*Borrador IA Creado* ✉️\n\n*Para:* %s\n*Asunto:* %s\n\n%s",
		data.To, data.Subject, data.Content)
}

// --- Flujo de alta de nuevo contacto ---

func (wppc *WhatsappController) startAddContactName(sender string) {
	state, ok := wppc.newContacts[sender]
	if !ok {
		state = &NewContactState{}
		wppc.newContacts[sender] = state
	}
	state.Stage = "awaiting_name"
	wppc.reply(sender, "Perfecto, dime cómo quieres llamar a este contacto (ej: Mamá, Juan, Contador).")
}

func (wppc *WhatsappController) cancelAddContactFlow(sender string) {
	delete(wppc.newContacts, sender)
	delete(wppc.pendingNewRecipient, sender)
	wppc.reply(sender, "OK, no lo guardo como contacto. Dime el email al que quieres enviar y lo uso solo esta vez.")
}

func (wppc *WhatsappController) handleNewContactName(sender, msg string) {
	name := strings.TrimSpace(msg)
	if name == "" {
		wppc.reply(sender, "Necesito un nombre para el contacto. Intenta de nuevo.")
		return
	}
	state := wppc.newContacts[sender]
	state.TempName = name
	state.Stage = "awaiting_email"
	wppc.reply(sender, fmt.Sprintf("Genial, '%s'. Ahora dime el email de este contacto.", name))
}

func (wppc *WhatsappController) handleNewContactEmail(sender, msg string, state *NewContactState) {
	email := strings.TrimSpace(msg)
	if !utilsController.IsValidEmail(email) {
		wppc.reply(sender, "Ese no parece un email válido. Intenta de nuevo (ejemplo: persona@dominio.com).")
		return
	}

	// Verificar duplicado por email (el nombre sí puede repetirse)
	for _, existingEmail := range wppc.contacts {
		if strings.EqualFold(existingEmail, email) {
			wppc.reply(sender, "Ya tengo un contacto con ese email. No lo vuelvo a guardar.")
			delete(wppc.newContacts, sender)
			delete(wppc.pendingNewRecipient, sender)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeout)
	defer cancel()

	if err := store.SetContact(ctx, state.TempName, email); err != nil {
		fmt.Printf("❌ Error guardando contacto en Redis: %v\n", err)
		wppc.reply(sender, "⚠️ No pude guardar el contacto. Intenta de nuevo más tarde.")
		delete(wppc.newContacts, sender)
		delete(wppc.pendingNewRecipient, sender)
		return
	}

	// Actualizar mapa en memoria
	wppc.contacts[state.TempName] = email
	wppc.reply(sender, fmt.Sprintf("Contacto guardado: %s -> %s ✅", state.TempName, email))

	// Si había un borrador IA pendiente sin destinatario resuelto, completarlo ahora
	if pending, ok := wppc.pendingNewRecipient[sender]; ok {
		ok, unresolved := wppc.validateAndResolveRecipient(pending)
		if !ok {
			if strings.TrimSpace(unresolved) != "" {
				wppc.reply(sender, fmt.Sprintf("Aún no tengo todos los destinatarios resueltos (me falta \"%s\").", unresolved))
				return
			}
			wppc.reply(sender, "Todavía me falta saber a quién enviar el mail. Intenta indicarlo de nuevo.")
			return
		}

		delete(wppc.pendingNewRecipient, sender)
		delete(wppc.newContacts, sender)
		wppc.finalizeDraftWithRecipient(sender, pending)
		return
	}

	delete(wppc.newContacts, sender)
}
