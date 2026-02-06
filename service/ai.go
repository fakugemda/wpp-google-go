package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"whatsapp-gmail-bot/constants"
	"whatsapp-gmail-bot/models"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// ProcessIntent - Procesa el mensaje del usuario usando Gemini AI y devuelve la respuesta estructurada
func ProcessIntent(userMessage, contactsList string) (*models.AIResponse, error) {
	ctx := context.Background()
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

	prompt := models.BuildPrompt(userMessage, contactsList)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("error generando contenido: %s", err.Error())
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("respuesta vacía del modelo")
	}

	part := resp.Candidates[0].Content.Parts[0]
	textPart, ok := part.(genai.Text)
	if !ok {
		return nil, fmt.Errorf("respuesta no es texto: %T", part)
	}

	var aiResponse models.AIResponse
	if err := json.Unmarshal([]byte(string(textPart)), &aiResponse); err != nil {
		return nil, fmt.Errorf("error parseando respuesta JSON: %s", err.Error())
	}

	return &aiResponse, nil
}
