package service

import (
	"fmt"
	"io"
	"net/http"
	"time"
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
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error http (%s): %s", resp.Status, string(bodyBytes))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo bytes: %v", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("el archivo descargado está vacío")
	}

	return data, nil
}
