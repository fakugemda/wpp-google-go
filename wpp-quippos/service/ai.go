package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
)

// ProcessMessage procesa un mensaje de WhatsApp con Gemini y genera una respuesta
func ProcessMessage(messageText, senderName, senderPhone string) (string, error) {
	ctx := context.Background()

	// El cliente obtiene la API key de la variable de entorno GEMINI_API_KEY
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("error creando cliente Gemini: %v", err)
	}

	// Construir el prompt con contexto del usuario
	prompt := buildPrompt(messageText, senderName, senderPhone)

	log.Printf("🧠 Consultando a Gemini con mensaje de %s...", senderName)

	// Generar contenido usando el modelo
	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview", // Modelo más reciente
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("error generando contenido: %v", err)
	}

	response := result.Text()
	log.Printf("✅ Gemini respondió: %s", response)

	return response, nil
}

// buildPrompt construye el prompt para Gemini con contexto del usuario
func buildPrompt(messageText, senderName, senderPhone string) string {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("⚠️  ADVERTENCIA: GEMINI_API_KEY no está configurada")
	}

	prompt := fmt.Sprintf(`Eres un asistente de WhatsApp amigable y conversacional.

Usuario: %s (Teléfono: %s)
Mensaje: "%s"

Genera una respuesta natural, amigable y concisa. 
- Usa el nombre del usuario si es apropiado
- Responde de forma conversacional como si fueras un humano
- Mantén la respuesta corta (máximo 2-3 oraciones)
- Si el usuario saluda, saluda de vuelta
- Si pregunta algo, responde de forma útil

Respuesta:`, senderName, senderPhone, messageText)

	return prompt
}
