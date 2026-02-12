package service

import (
	"fmt"
	"sync"
	"time"
)

const (
	ImageCacheTTL             = 30 * time.Minute
	ImageCacheCleanupInterval = 2 * time.Minute
)

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
	cleanupStarted  bool
	cleanupOnce     sync.Once
)

func StartImageCacheCleanup() {
	cleanupOnce.Do(func() {
		cleanupStopChan = make(chan struct{})
		cleanupStarted = true
		go startImageCacheCleanup()
	})
}

func StopImageCacheCleanup() {
	if cleanupStarted && cleanupStopChan != nil {
		close(cleanupStopChan)
		cleanupStarted = false
	}
}

func GetImage(userID string) (*CachedImage, bool) {
	ImageCacheMutex.RLock()
	defer ImageCacheMutex.RUnlock()
	img, exists := ImageCache[userID]
	return img, exists
}

func SetImage(userID string, img *CachedImage) {
	ImageCacheMutex.Lock()
	defer ImageCacheMutex.Unlock()
	ImageCache[userID] = img
}

func DeleteImage(userID string) {
	ImageCacheMutex.Lock()
	defer ImageCacheMutex.Unlock()
	delete(ImageCache, userID)
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
