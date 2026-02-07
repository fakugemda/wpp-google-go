package models

import (
	"fmt"
)

var ReplyMessage = `<?xml version="1.0" encoding="UTF-8"?><Response><Message>%s</Message></Response>`

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
   - Identifica el destinatario. Si menciona un nombre/apodo, búscalo en la lista de contactos y usa su email.
   - Si no encuentras el contacto o no hay email explícito, pon "PENDIENTE".
   - SIEMPRE devuelve un email válido en el campo "to", nunca solo un nombre.
   - Redacta un asunto profesional o acorde al tono.
   - Redacta el cuerpo del mensaje mejorando lo que dijo el usuario. Hazlo elegante pero conciso.
   - Devuelve JSON: {"type": "EMAIL_DRAFT", "to": "email@ejemplo.com", "subject": "...", "content": "..."}

2. Si el usuario solo está saludando o charlando:
   - Responde amablemente.
   - Devuelve JSON: {"type": "CHAT", "content": "Tu respuesta aquí"}

3. Si el usuario pide confirmar un envío anterior (palabras como "sí", "envíalo", "dale"):
   - Devuelve JSON: {"type": "CONFIRM_SEND", "content": ""}

4. Si el usuario pide un tono mas informal o violento, responde con un tono mas informal o violento.

IMPORTANTE: El campo "to" DEBE ser siempre un email válido (ejemplo@dominio.com), nunca solo un nombre.

NO uses markdown, solo JSON puro.`

// BuildPrompt: Construye el prompt formateado con el mensaje del usuario y la lista de contactos
func BuildPrompt(userMessage, contactsList string) string {
	prompt := fmt.Sprintf(PromptTemplate, userMessage)
	if contactsList != "" {
		prompt = contactsList + "\n\n" + prompt
	}
	return prompt
}
