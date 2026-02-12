package service

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
	"whatsapp-gmail-bot/models"

	"github.com/bwmarrin/discordgo"
)

var (
	discordDrafts   = make(map[string]*models.AIResponse)
	discordContacts string
)

// StartDiscordService inicializa la conexión y se queda escuchando
func StartDiscordService(contactsList string) {
	discordContacts = contactsList
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("⚠️ DISCORD_TOKEN no configurado. Saltando inicio de Discord.")
		return
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("❌ Error creando sesión Discord:", err)
		return
	}

	dg.AddHandler(discordMessageHandler)
	dg.Identify.Intents = discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

	StartImageCacheCleanup()

	// Abrir conexión
	err = dg.Open()
	if err != nil {
		fmt.Println("❌ Error conectando a Discord WebSocket:", err)
		return
	}

	fmt.Println("🤖 Servicio Discord ONLINE (Escuchando DMs)")
}

// discordMessageHandler maneja cada mensaje que entra
func discordMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.GuildID != "" {
		return
	}

	var promptText string = m.Content

	if len(m.Attachments) > 0 {
		attachment := m.Attachments[0]

		// Determinar tipo de archivo para mensaje descriptivo
		isImage := attachment.Width > 0 || attachment.Height > 0 || isImageExtension(attachment.Filename)
		fileType := "📄 Archivo"
		receivedText := "recibido"
		if isImage {
			fileType = "📸 Imagen"
			receivedText = "recibida"
		}

		fmt.Printf("%s [Discord] Adjunto %s de %s: %s\n", fileType, receivedText, m.Author.Username, attachment.Filename)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⬇️ Descargando %s...", attachment.Filename))

		// 1. Descargar archivo (funciona para cualquier tipo)
		fileBytes, err := DownloadFile(attachment.URL)
		if err != nil {
			fmt.Printf("❌ Error descargando archivo de Discord: %s\n", err.Error())
			s.ChannelMessageSend(m.ChannelID, "❌ Error descargando archivo: "+err.Error())
			return
		}

		// 2. Detectar MIME type desde extensión
		mimeType := mime.TypeByExtension(filepath.Ext(attachment.Filename))
		if mimeType == "" {
			mimeType = "application/octet-stream" // Tipo genérico por defecto
		} else {
			mimeType = NormalizeMimeType(mimeType)
		}

		// 3. Guardar en caché global
		SetImageDiscord(m.Author.ID, &CachedImage{
			Bytes:     fileBytes,
			MimeType:  mimeType,
			Filename:  attachment.Filename,
			CreatedAt: time.Now(),
		})

		fmt.Printf("✅ [Discord] %s guardado en caché (tipo: %s, tamaño: %d bytes)\n", fileType, mimeType, len(fileBytes))
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ %s guardado temporalmente. ¿Qué hago con él?", attachment.Filename))

		if promptText == "" {
			return
		}
	}

	fmt.Printf("💬 [Discord] %s: %s\n", m.Author.Username, promptText)

	aiResp, err := ProcessIntent(promptText, discordContacts)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "⚠️ Error procesando solicitud: "+err.Error())
		return
	}

	switch aiResp.Type {
	case "CHAT":
		s.ChannelMessageSend(m.ChannelID, aiResp.Content)

	case "EMAIL_DRAFT":
		discordDrafts[m.Author.ID] = aiResp

		// Verificar si hay adjunto en caché para este usuario
		var hasAttachment bool
		if cachedImg, ok := GetImageDiscord(m.Author.ID); ok {
			hasAttachment = true
			attachmentType := "archivo"
			if strings.HasPrefix(cachedImg.MimeType, "image/") {
				attachmentType = "imagen"
			}
			fmt.Printf("📎 [Discord] Incluyendo %s en borrador (tipo: %s, tamaño: %d bytes)\n", attachmentType, cachedImg.MimeType, len(cachedImg.Bytes))
		}

		attachmentNote := ""
		if hasAttachment {
			attachmentNote = "\n📎 _Incluye adjunto_"
		}

		msg := fmt.Sprintf("**📝 Borrador Generado**\n\n**Para:** `%s`\n**Asunto:** `%s`\n\n%s%s\n\n_Escribe 'sí' o 'envíalo' para confirmar._",
			aiResp.To, aiResp.Subject, aiResp.Content, attachmentNote)
		s.ChannelMessageSend(m.ChannelID, msg)

	case "CONFIRM_SEND":
		draft, exists := discordDrafts[m.Author.ID]
		if !exists {
			s.ChannelMessageSend(m.ChannelID, "🤷‍♂️ No tengo ningún mail pendiente. Pídeme redactar uno primero.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "🚀 Enviando correo...")

		// Buscar adjunto en caché
		var attachmentData []byte
		var filename string
		if cachedImg, ok := GetImageDiscord(m.Author.ID); ok {
			attachmentData = cachedImg.Bytes
			if cachedImg.Filename != "" {
				filename = cachedImg.Filename
			} else {
				filename = getFilenameFromMimeType(cachedImg.MimeType)
			}
			attachmentType := "archivo"
			if strings.HasPrefix(cachedImg.MimeType, "image/") {
				attachmentType = "imagen"
			}
			fmt.Printf("📎 [Discord] Adjuntando %s al correo: %s\n", attachmentType, filename)
		}

		err := SendEmail(draft.To, draft.Subject, draft.Content, attachmentData, filename)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Error enviando: "+err.Error())
		} else {
			s.ChannelMessageSend(m.ChannelID, "✅ ¡Correo enviado exitosamente!")
			delete(discordDrafts, m.Author.ID)
			DeleteImageDiscord(m.Author.ID)
		}
	}
}

// isImageExtension verifica si un archivo es una imagen basándose en su extensión
func isImageExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff", ".svg"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// getFilenameFromMimeType convierte un MIME type a un nombre de archivo con extensión apropiada
func getFilenameFromMimeType(mimeType string) string {
	mimeToExt := map[string]string{
		"image/jpeg":         ".jpg",
		"image/jpg":          ".jpg",
		"image/png":          ".png",
		"image/gif":          ".gif",
		"image/webp":         ".webp",
		"image/bmp":          ".bmp",
		"image/tiff":         ".tiff",
		"image/svg+xml":      ".svg",
		"application/pdf":    ".pdf",
		"application/msword": ".doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
		"application/vnd.ms-excel": ".xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
		"text/plain": ".txt",
		"text/csv":   ".csv",
	}

	if ext, ok := mimeToExt[strings.ToLower(mimeType)]; ok {
		// Determinar prefijo según tipo
		if strings.HasPrefix(mimeType, "image/") {
			return "foto_discord" + ext
		}
		return "archivo_discord" + ext
	}

	// Fallback genérico (no siempre .jpg)
	return "archivo_discord.bin"
}

// SendEmail crea un borrador y lo envía inmediatamente
// Si attachmentData está vacío, envía un email simple
// Si attachmentData tiene contenido, envía un email con adjunto
func SendEmail(to, subject, body string, attachmentData []byte, filename string) error {
	draftID, err := CreateDraft(to, subject, body, attachmentData, filename)
	if err != nil {
		return fmt.Errorf("error creando borrador: %s", err.Error())
	}

	err = SendDraft(draftID)
	if err != nil {
		return fmt.Errorf("error enviando borrador: %s", err.Error())
	}

	return nil
}
