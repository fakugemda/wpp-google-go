package service

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	MaxFileSize = 10 * 1024 * 1024 // 10 MB
)

// DownloadFile descarga un archivo desde una URL pública (como las de Discord)
func DownloadFile(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error descargando archivo: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		limitedBody := io.LimitReader(resp.Body, 1024)
		bodyBytes, _ := io.ReadAll(limitedBody)
		return nil, fmt.Errorf("error http (%s): %s", resp.Status, string(bodyBytes))
	}

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		size, err := strconv.ParseInt(contentLength, 10, 64)
		if err == nil && size > MaxFileSize {
			return nil, fmt.Errorf("archivo demasiado grande: %d bytes (máximo permitido: %d bytes)", size, MaxFileSize)
		}
	}

	limitedReader := io.LimitReader(resp.Body, MaxFileSize+1)

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("error leyendo bytes: %v", err)
	}

	if len(data) > MaxFileSize {
		return nil, fmt.Errorf("archivo demasiado grande: excede el límite de %d bytes", MaxFileSize)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("el archivo descargado está vacío")
	}

	return data, nil
}
