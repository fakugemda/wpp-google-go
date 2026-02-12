package utils

import (
	"path/filepath"
	"strings"
	"whatsapp-gmail-bot/discord/models"
)

func DetectFileType(filename string, width, height int) models.FileInfo {
	isImage := width > 0 || height > 0 || IsImageExtension(filename)

	if isImage {
		return models.FileInfo{
			IsImage:     true,
			Type:        "📸 Imagen",
			DisplayName: "recibida",
		}
	}

	return models.FileInfo{
		IsImage:     false,
		Type:        "📄 Archivo",
		DisplayName: "recibido",
	}
}

func IsImageExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff", ".svg"}

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}

	return false
}

func GenerateFilenameFromMime(mimeType string) string {
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

	ext, exists := mimeToExt[strings.ToLower(mimeType)]
	if !exists {
		return "archivo_discord.bin"
	}

	if strings.HasPrefix(mimeType, "image/") {
		return "foto_discord" + ext
	}

	return "archivo_discord" + ext
}

func DetermineAttachmentType(mimeType string) string {
	if strings.HasPrefix(mimeType, "image/") {
		return "imagen"
	}
	return "archivo"
}
