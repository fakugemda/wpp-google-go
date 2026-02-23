package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"whatsapp-gmail-bot/constants"
	"whatsapp-gmail-bot/models"
	"whatsapp-gmail-bot/utils"

	// Asegúrate de que esta ruta coincida con donde guardaste el archivo template.go
	"whatsapp-gmail-bot/models/template"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiAPIKey se inicializa en config/app.go al arrancar el microservicio
var GeminiAPIKey string

var utilsAI = utils.GetUtils()

// ProcessIntent - Procesa el mensaje usando Gemini AI
func ProcessIntent(userMessage, contactsList string) (*models.AIResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(GeminiAPIKey))
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
		fmt.Printf("❌ Error parseando JSON\n")
		fmt.Printf("   Raw: %s\n", rawJSON)
		fmt.Printf("   Clean: %s\n", cleanJSON)
		fmt.Printf("   Error: %s\n", err.Error())
		return nil, fmt.Errorf("error parseando JSON: %s", err.Error())
	}

	return &aiResponse, nil
}

// ProcessCorrection - Corrige un borrador existente con la instrucción del usuario
func ProcessCorrection(existingDraft *models.AIResponse, correction string) (*models.AIResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(GeminiAPIKey))
	if err != nil {
		return nil, fmt.Errorf("error creando cliente: %s", err.Error())
	}
	defer client.Close()

	model := client.GenerativeModel(constants.GEMINI_MODEL)
	model.ResponseMIMEType = "application/json"

	prompt := template.BuildCorrectionPrompt(existingDraft.To, existingDraft.Subject, existingDraft.Content, correction)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("error generando corrección: %s", err.Error())
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
		return nil, fmt.Errorf("error parseando corrección JSON: %s", err.Error())
	}

	return &aiResponse, nil
}
