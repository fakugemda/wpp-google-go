package controllers

import (
	"fmt"
	"whatsapp-gmail-bot/models"
)

func buildDownloadingMessage(filename string) string {
	return fmt.Sprintf("⬇️ Descargando %s...", filename)
}

func buildAttachmentSavedMessage(filename string) string {
	return fmt.Sprintf("✅ %s guardado temporalmente. ¿Qué hago con él?", filename)
}

func buildAttachmentErrorMessage(err error) string {
	return "❌ Error descargando archivo: " + err.Error()
}

func buildDraftPreviewMessage(draft *models.AIResponse, hasAttachment bool) string {
	attachmentNote := ""
	if hasAttachment {
		attachmentNote = "\n📎 _Incluye adjunto_"
	}

	return fmt.Sprintf(
		"**📝 Borrador Generado**\n\n**Para:** `%s`\n**Asunto:** `%s`\n\n%s%s\n\n_Escribe 'sí' o 'envíalo' para confirmar._",
		draft.To,
		draft.Subject,
		draft.Content,
		attachmentNote,
	)
}

func buildNoDraftMessage() string {
	return "🤷‍♂️ No tengo ningún mail pendiente. Pídeme redactar uno primero."
}

func buildSendingEmailMessage() string {
	return "🚀 Enviando correo..."
}

func buildEmailSentMessage() string {
	return "✅ ¡Correo enviado exitosamente!"
}

func buildEmailErrorMessage(err error) string {
	return "❌ Error enviando: " + err.Error()
}

func buildProcessingErrorMessage(err error) string {
	return "⚠️ Error procesando solicitud: " + err.Error()
}

