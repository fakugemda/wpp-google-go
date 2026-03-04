package models

// Estructura para el contacto
type Contact struct {
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Email    string `json:"email"`
}

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
					Type string `json:"type"` // "text", "image" o "document"
					Text struct {
						Body string `json:"body"`
					} `json:"text,omitempty"`
					Image struct {
						ID      string `json:"id"`
						Caption string `json:"caption,omitempty"`
					} `json:"image,omitempty"`
					Document struct {
						ID       string `json:"id"`
						Caption  string `json:"caption,omitempty"`
						Filename string `json:"filename,omitempty"`
						MimeType string `json:"mime_type,omitempty"`
					} `json:"document,omitempty"`
					Interactive struct {
						Type        string `json:"type"`
						ButtonReply struct {
							ID    string `json:"id"`
							Title string `json:"title"`
						} `json:"button_reply"`
						ListReply struct {
							ID    string `json:"id"`
							Title string `json:"title"`
						} `json:"list_reply"`
					} `json:"interactive,omitempty"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

// Estructura para leer la respuesta de la URL de Meta
type MediaURLResponse struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
}
