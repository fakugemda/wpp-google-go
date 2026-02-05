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

func ProcessIntent(userMessage string) (*models.AIResponse, error) {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creando cliente: %s", err.Error())
	}
	defer client.Close()

	model := client.GenerativeModel(constants.GEMINI_MODEL)
	model.ResponseMIMEType = "application/json"

	// El Prompt del Sistema es la clave. Aquí le das la personalidad.
	prompt := models.BuildPrompt(userMessage)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("respuesta vacía del modelo")
	}

	jsonRaw := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	var aiResponse models.AIResponse
	if err := json.Unmarshal([]byte(jsonRaw), &aiResponse); err != nil {
		return nil, fmt.Errorf("error parseando respuesta JSON: %s", err.Error())
	}

	return &aiResponse, nil

}
