package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"whatsapp-gmail-bot/constants"
	"whatsapp-gmail-bot/models"
	"whatsapp-gmail-bot/utils"

	// Asegúrate de que esta ruta coincida con donde guardaste el archivo template.go
	"whatsapp-gmail-bot/models/template"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var utilsAI = utils.GetUtils()

// ProcessIntent - Procesa el mensaje usando Gemini AI
func ProcessIntent(userMessage, contactsList string) (*models.AIResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY no está configurada")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creando cliente: %s", err.Error())
	}
	defer client.Close()

	model := client.GenerativeModel(constants.GEMINI_MODEL)
	model.ResponseMIMEType = "application/json"
	model.SetTemperature(0.8)

	prompt := template.BuildPrompt(userMessage, contactsList)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("error generando contenido: %s", err.Error())
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("respuesta vacía del modelo")
	}

	part := resp.Candidates[0].Content.Parts[0]

	var rawJSON string
	if txt, ok := part.(genai.Text); ok {
		rawJSON = string(txt)
	} else {
		return nil, fmt.Errorf("respuesta no es texto")
	}

	cleanJSON := utilsAI.CleanJSONString(rawJSON)

	var aiResponse models.AIResponse
	if err := json.Unmarshal([]byte(cleanJSON), &aiResponse); err != nil {
		fmt.Printf("❌ Error parseando JSON sucio: %s\n", rawJSON)
		return nil, fmt.Errorf("error parseando JSON: %s", err.Error())
	}

	return &aiResponse, nil
}
