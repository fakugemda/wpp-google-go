package models

// WebhookPayload representa el payload completo recibido de WhatsApp
type WebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

// Entry representa una entrada en el webhook
type Entry struct {
	ID      string   `json:"id"`
	Changes []Change `json:"changes"`
}

// Change representa un cambio en el webhook
type Change struct {
	Value Value  `json:"value"`
	Field string `json:"field"`
}

// Value contiene los valores del cambio
type Value struct {
	MessagingProduct string    `json:"messaging_product"`
	Metadata         Metadata  `json:"metadata"`
	Contacts         []Contact `json:"contacts,omitempty"`
	Messages         []Message `json:"messages,omitempty"`
	Statuses         []Status  `json:"statuses,omitempty"`
}

// Metadata contiene metadata del webhook
type Metadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

// Contact representa información de contacto del usuario
type Contact struct {
	Profile Profile `json:"profile"`
	WaID    string  `json:"wa_id"`
}

// Profile contiene el perfil del usuario
type Profile struct {
	Name string `json:"name"`
}

// Message representa un mensaje recibido
type Message struct {
	From      string      `json:"from"`
	ID        string      `json:"id"`
	Timestamp string      `json:"timestamp"`
	Type      string      `json:"type"`
	Text      TextContent `json:"text,omitempty"`
}

// TextContent contiene el contenido de texto de un mensaje
type TextContent struct {
	Body string `json:"body"`
}

// Status representa un estado de mensaje
type Status struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Timestamp    string `json:"timestamp"`
	RecipientID  string `json:"recipient_id"`
	Conversation struct {
		ID     string `json:"id"`
		Origin struct {
			Type string `json:"type"`
		} `json:"origin"`
	} `json:"conversation,omitempty"`
}
