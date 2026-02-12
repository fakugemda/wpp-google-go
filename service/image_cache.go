package service

import (
	"fmt"
	"mime"
	"sync"
	"time"
)

// NormalizeMimeType normaliza un MIME type eliminando parámetros (p. ej. "; charset=utf-8")
func NormalizeMimeType(mimeType string) string {
	if mimeType == "" {
		return ""
	}

	// Parsear el MIME type para obtener solo el tipo base
	if base, _, err := mime.ParseMediaType(mimeType); err == nil && base != "" {
		return base
	}

	// Si falla el parseo, retornar el original
	return mimeType
}

const (
	ImageCacheTTL             = 30 * time.Minute
	ImageCacheCleanupInterval = 2 * time.Minute

	// Namespaces para evitar colisiones entre plataformas
	NamespaceWhatsApp = "whatsapp"
	NamespaceDiscord  = "discord"
)

// buildCacheKey construye una clave namespaced para el caché
func buildCacheKey(namespace, userID string) string {
	return fmt.Sprintf("%s:%s", namespace, userID)
}

type CachedImage struct {
	Bytes     []byte
	MimeType  string
	Filename  string
	CreatedAt time.Time
}

var (
	ImageCache      = make(map[string]*CachedImage)
	ImageCacheMutex sync.RWMutex
	cleanupStopChan chan struct{}
	cleanupOnce     sync.Once
	cleanupStopOnce sync.Once
)

func StartImageCacheCleanup() {
	cleanupOnce.Do(func() {
		cleanupStopChan = make(chan struct{})
		go startImageCacheCleanup()
	})
}

func StopImageCacheCleanup() {
	cleanupStopOnce.Do(func() {
		if cleanupStopChan != nil {
			close(cleanupStopChan)
		}
	})
}

// GetImage obtiene una imagen del caché usando namespace y userID
func GetImage(namespace, userID string) (*CachedImage, bool) {
	key := buildCacheKey(namespace, userID)
	ImageCacheMutex.RLock()
	defer ImageCacheMutex.RUnlock()
	img, exists := ImageCache[key]
	return img, exists
}

// SetImage guarda una imagen en el caché usando namespace y userID
func SetImage(namespace, userID string, img *CachedImage) {
	key := buildCacheKey(namespace, userID)
	ImageCacheMutex.Lock()
	defer ImageCacheMutex.Unlock()
	ImageCache[key] = img
}

// DeleteImage elimina una imagen del caché usando namespace y userID
func DeleteImage(namespace, userID string) {
	key := buildCacheKey(namespace, userID)
	ImageCacheMutex.Lock()
	defer ImageCacheMutex.Unlock()
	delete(ImageCache, key)
}

// GetImageWhatsApp obtiene una imagen de WhatsApp del caché
func GetImageWhatsApp(userID string) (*CachedImage, bool) {
	return GetImage(NamespaceWhatsApp, userID)
}

// SetImageWhatsApp guarda una imagen de WhatsApp en el caché
func SetImageWhatsApp(userID string, img *CachedImage) {
	SetImage(NamespaceWhatsApp, userID, img)
}

// DeleteImageWhatsApp elimina una imagen de WhatsApp del caché
func DeleteImageWhatsApp(userID string) {
	DeleteImage(NamespaceWhatsApp, userID)
}

// GetImageDiscord obtiene una imagen de Discord del caché
func GetImageDiscord(userID string) (*CachedImage, bool) {
	return GetImage(NamespaceDiscord, userID)
}

// SetImageDiscord guarda una imagen de Discord en el caché
func SetImageDiscord(userID string, img *CachedImage) {
	SetImage(NamespaceDiscord, userID, img)
}

// DeleteImageDiscord elimina una imagen de Discord del caché
func DeleteImageDiscord(userID string) {
	DeleteImage(NamespaceDiscord, userID)
}

func startImageCacheCleanup() {
	ticker := time.NewTicker(ImageCacheCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cleanExpiredImages()
		case <-cleanupStopChan:
			fmt.Println("🛑 Deteniendo limpieza de caché de imágenes...")
			return
		}
	}
}

func cleanExpiredImages() {
	ImageCacheMutex.Lock()
	defer ImageCacheMutex.Unlock()

	now := time.Now()
	var expiredKeys []string

	for key, cachedImg := range ImageCache {
		if now.Sub(cachedImg.CreatedAt) > ImageCacheTTL {
			expiredKeys = append(expiredKeys, key)
		}
	}

	if len(expiredKeys) > 0 {
		for _, key := range expiredKeys {
			delete(ImageCache, key)
		}
		fmt.Printf("🧹 Limpieza de caché: eliminadas %d imagen(es) expirada(s)\n", len(expiredKeys))
	}
}
