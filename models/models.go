package models

// Definimos la estructura que queremos que Gemini nos devuelva siempre
type AIResponse struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	To      string `json:"to,omitempty"`
	Subject string `json:"subject,omitempty"`
}

// MetaWebhook: Estructura para leer el JSON de Meta/Facebook
type MetaWebhook struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Messages []struct {
					From string `json:"from"`
					Text struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}
