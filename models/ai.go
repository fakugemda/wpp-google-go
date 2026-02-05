package models

import (
	"fmt"
)

// Definimos la estructura que queremos que Gemini nos devuelva siempre
type AIResponse struct {
	Type    string `json:"type"`    // "EMAIL_DRAFT" o "CHAT"
	Content string `json:"content"` // La respuesta de chat o el cuerpo del mail
	To      string `json:"to,omitempty"`
	Subject string `json:"subject,omitempty"`
}

// PromptTemplate: Plantilla del prompt del sistema para el asistente de IA
const PromptTemplate = `Eres un asistente personal inteligente conectado al Gmail de Facu.
Tu trabajo es analizar el mensaje del usuario: "%s"

Reglas:
1. Si el usuario quiere enviar un correo:
   - Identifica el destinatario (si no hay email explícito, pon "PENDIENTE").
   - Redacta un asunto profesional o acorde al tono.
   - Redacta el cuerpo del mensaje mejorando lo que dijo el usuario. Hazlo elegante pero conciso.
   - Devuelve JSON: {"type": "EMAIL_DRAFT", "to": "...", "subject": "...", "content": "..."}

2. Si el usuario solo está saludando o charlando:
   - Responde amablemente.
   - Devuelve JSON: {"type": "CHAT", "content": "Tu respuesta aquí"}

3. Si el usuario pide confirmar un envío anterior (palabras como "sí", "envíalo", "dale"):
   - Devuelve JSON: {"type": "CONFIRM_SEND", "content": ""}

NO uses markdown, solo JSON puro.`

// BuildPrompt: Construye el prompt formateado con el mensaje del usuario
func BuildPrompt(userMessage string) string {
	return fmt.Sprintf(PromptTemplate, userMessage)
}
