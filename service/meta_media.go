package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"whatsapp-gmail-bot/models"
)

// DownloadMedia obtiene la URL de la imagen desde la API de Meta
func DownloadMedia(mediaID string) ([]byte, string, error) {
	token := os.Getenv("META_TOKEN")
	if token == "" {
		return nil, "", fmt.Errorf("META_TOKEN no está configurada")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	urlRequest := fmt.Sprintf("https://graph.facebook.com/v22.0/%s", mediaID)

	req, err := http.NewRequest("GET", urlRequest, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error creando request para obtener URL de media: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error obteniendo URL de media: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("Meta API error (%s): %s", resp.Status, string(bodyBytes))
	}

	var mediaResp models.MediaURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&mediaResp); err != nil {
		return nil, "", fmt.Errorf("error decodificando JSON de media: %v", err)
	}

	if mediaResp.URL == "" {
		return nil, "", fmt.Errorf("la respuesta de Meta no contiene URL de descarga")
	}

	downReq, err := http.NewRequest("GET", mediaResp.URL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error creando request para descargar binario: %v", err)
	}

	downReq.Header.Set("Authorization", "Bearer "+token)

	downResp, err := client.Do(downReq)
	if err != nil {
		return nil, "", fmt.Errorf("error descargando binario: %v", err)
	}
	defer downResp.Body.Close()

	if downResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(downResp.Body)
		return nil, "", fmt.Errorf("error descargando archivo (%s): %s", downResp.Status, string(bodyBytes))
	}

	fileBytes, err := io.ReadAll(downResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("error leyendo bytes del archivo: %v", err)
	}

	if len(fileBytes) == 0 {
		return nil, "", fmt.Errorf("el archivo descargado está vacío")
	}

	return fileBytes, mediaResp.MimeType, nil
}
