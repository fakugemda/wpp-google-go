package services

import (
	"fmt"
	"mime"
	"path/filepath"
	"time"
	"whatsapp-gmail-bot/discord/models"
	"whatsapp-gmail-bot/discord/utils"
	"whatsapp-gmail-bot/service"

	"github.com/bwmarrin/discordgo"
)

type AttachmentService struct{}

func (as *AttachmentService) ProcessAttachment(attachment *discordgo.MessageAttachment, userID, username string) error {
	fileInfo := utils.DetectFileType(attachment.Filename, attachment.Width, attachment.Height)
	as.logAttachmentReceived(username, attachment.Filename, fileInfo)

	fileBytes, err := as.downloadFile(attachment.URL)
	if err != nil {
		return err
	}

	mimeType := as.detectMimeType(attachment.Filename)
	as.cacheFile(userID, fileBytes, mimeType, attachment.Filename)
	as.logAttachmentCached(fileInfo.Type, mimeType, len(fileBytes))

	return nil
}

func (as *AttachmentService) downloadFile(url string) ([]byte, error) {
	fileBytes, err := service.DownloadFile(url)
	if err != nil {
		fmt.Printf("❌ Error descargando archivo de Discord: %s\n", err.Error())
		return nil, err
	}
	return fileBytes, nil
}

func (as *AttachmentService) detectMimeType(filename string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType == "" {
		return "application/octet-stream"
	}
	return service.NormalizeMimeType(mimeType)
}

func (as *AttachmentService) cacheFile(userID string, fileBytes []byte, mimeType, filename string) {
	service.SetImageDiscord(userID, &service.CachedImage{
		Bytes:     fileBytes,
		MimeType:  mimeType,
		Filename:  filename,
		CreatedAt: time.Now(),
	})
}

func (as *AttachmentService) logAttachmentReceived(username, filename string, fileInfo models.FileInfo) {
	fmt.Printf("%s [Discord] Adjunto %s de %s: %s\n", fileInfo.Type, fileInfo.DisplayName, username, filename)
}

func (as *AttachmentService) logAttachmentCached(fileType, mimeType string, size int) {
	fmt.Printf("✅ [Discord] %s guardado en caché (tipo: %s, tamaño: %d bytes)\n", fileType, mimeType, size)
}

func (as *AttachmentService) RetrieveFromCache(userID string) (attachmentData []byte, filename string, hasAttachment bool) {
	cachedImg, ok := service.GetImageDiscord(userID)
	if !ok {
		return nil, "", false
	}

	attachmentData = cachedImg.Bytes
	filename = cachedImg.Filename
	if filename == "" {
		filename = utils.GenerateFilenameFromMime(cachedImg.MimeType)
	}

	attachmentType := utils.DetermineAttachmentType(cachedImg.MimeType)
	fmt.Printf("📎 [Discord] Adjuntando %s al correo: %s\n", attachmentType, filename)

	return attachmentData, filename, true
}

func (as *AttachmentService) DeleteFromCache(userID string) {
	service.DeleteImageDiscord(userID)
}

func (as *AttachmentService) HasAttachmentInCache(userID string) bool {
	_, ok := service.GetImageDiscord(userID)
	return ok
}

func (as *AttachmentService) LogAttachmentInDraft(userID string) {
	cachedImg, ok := service.GetImageDiscord(userID)
	if !ok {
		return
	}

	attachmentType := utils.DetermineAttachmentType(cachedImg.MimeType)
	fmt.Printf("📎 [Discord] Incluyendo %s en borrador (tipo: %s, tamaño: %d bytes)\n", attachmentType, cachedImg.MimeType, len(cachedImg.Bytes))
}
