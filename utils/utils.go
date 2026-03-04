package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"whatsapp-gmail-bot/models"
)

// Utils estructura singleton con funciones de utilidad genérica
type Utils struct{}

var (
	utilsInstance *Utils
	once          sync.Once
)

// GetUtils devuelve la instancia singleton de Utils
func GetUtils() *Utils {
	once.Do(func() {
		utilsInstance = &Utils{}
	})
	return utilsInstance
}

// CleanJSONString elimina los bloques de markdown ```json y ``` y extrae solo el JSON válido
func (u *Utils) CleanJSONString(s string) string {
	s = strings.TrimSpace(s)

	// Eliminar bloques de markdown
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	// Encontrar el inicio del JSON (primer '{' o '[')
	startIdx := strings.IndexAny(s, "{[")
	if startIdx == -1 {
		return s
	}

	// Encontrar el final del JSON válido
	// Para objetos: buscar el '}' que cierra
	// Para arrays: buscar el ']' que cierra
	isObject := s[startIdx] == '{'

	if isObject {
		// Contar llaves para encontrar el cierre correcto
		braceCount := 0
		for i := startIdx; i < len(s); i++ {
			switch s[i] {
			case '{':
				braceCount++
			case '}':
				braceCount--
				if braceCount == 0 {
					return strings.TrimSpace(s[startIdx : i+1])
				}
			}
		}
	} else {
		// Contar corchetes para arrays
		bracketCount := 0
		for i := startIdx; i < len(s); i++ {
			switch s[i] {
			case '[':
				bracketCount++
			case ']':
				bracketCount--
				if bracketCount == 0 {
					return strings.TrimSpace(s[startIdx : i+1])
				}
			}
		}
	}

	return strings.TrimSpace(s[startIdx:])
}

// IsValidEmail valida el formato completo de un email
func (u *Utils) IsValidEmail(email string) bool {
	if email == "" || email == "PENDIENTE" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

// NormalizeArgPhoneNumber normaliza números de teléfono al formato internacional
// Maneja específicamente el caso argentino: 549345xxxxxxx -> 54345xxxxxxx
func (u *Utils) NormalizeArgPhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.TrimPrefix(phone, "+")

	if strings.HasPrefix(phone, "549") && len(phone) >= 13 {
		phone = "54" + phone[3:]
	}

	return phone
}

// ContainsHTML detecta si un texto contiene etiquetas HTML
func (u *Utils) ContainsHTML(text string) bool {
	htmlTags := []string{"<img", "<html", "<body", "<div", "<p>", "<br", "<a ", "<span", "<h1", "<h2", "<h3", "<table", "<tr", "<td", "<th", "<ul", "<ol", "<li"}
	textLower := strings.ToLower(text)
	for _, tag := range htmlTags {
		if strings.Contains(textLower, tag) {
			return true
		}
	}
	return false
}

// LoadResource carga un recurso desde variable de entorno o archivo como fallback
func (u *Utils) LoadResource(envName, fileName string) ([]byte, error) {
	envContent := os.Getenv(envName)
	if envContent != "" {
		fmt.Printf("Cargando %s desde Variable de Entorno\n", envName)
		return []byte(envContent), nil
	}

	fmt.Printf("Cargando %s desde Archivo Local: %s\n", envName, fileName)
	return os.ReadFile(fileName)
}

// LoadJSONFromEnvOrFile carga y deserializa JSON desde variable de entorno o archivo
// Nota: Esta función no puede ser método porque Go no soporta métodos con type parameters
func LoadJSONFromEnvOrFile[T any](envName, fileName string) (T, error) {
	var result T
	envContent := os.Getenv(envName)

	if envContent != "" {
		fmt.Printf("Cargando %s desde Variable de Entorno\n", envName)
		err := json.Unmarshal([]byte(envContent), &result)
		return result, err
	}

	fmt.Printf("Cargando %s desde Archivo Local: %s\n", envName, fileName)
	f, err := os.Open(fileName)
	if err != nil {
		return result, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&result)
	return result, err
}

func (u *Utils) FindContactsByName(contacts []models.Contact, name string) []models.Contact {
	nameLower := strings.ToLower(strings.TrimSpace(name))
	fullName := func(c models.Contact) string { return strings.ToLower(c.Name + " " + c.Lastname) }
	lastnameFirst := func(c models.Contact) string { return strings.ToLower(c.Lastname + " " + c.Name) }
	seen := make(map[string]bool)
	var out []models.Contact
	add := func(c models.Contact) {
		if !seen[c.Email] {
			seen[c.Email] = true
			out = append(out, c)
		}
	}

	for _, c := range contacts {
		if fullName(c) == nameLower || lastnameFirst(c) == nameLower {
			add(c)
		}
	}
	for _, c := range contacts {
		if strings.ToLower(strings.TrimSpace(c.Name)) == nameLower || strings.ToLower(strings.TrimSpace(c.Lastname)) == nameLower {
			add(c)
		}
	}
	for _, c := range contacts {
		full := fullName(c)
		if strings.Contains(full, nameLower) || strings.Contains(nameLower, full) {
			add(c)
		}
	}

	if strings.Contains(nameLower, " ") {
		words := strings.Fields(nameLower)
		if len(words) == 0 {
			return out
		}
		firstWord := words[0]
		var byFirstName []models.Contact
		for _, c := range contacts {
			if strings.ToLower(strings.TrimSpace(c.Name)) == firstWord {
				byFirstName = append(byFirstName, c)
			}
		}
		if len(byFirstName) > len(out) {
			return byFirstName
		}
	}

	return out
}

// GetContactByEmail returns the contact with the given email, or nil if not found.
func (u *Utils) GetContactByEmail(contacts []models.Contact, email string) *models.Contact {
	email = strings.TrimSpace(strings.ToLower(email))
	for i := range contacts {
		if strings.ToLower(strings.TrimSpace(contacts[i].Email)) == email {
			return &contacts[i]
		}
	}
	return nil
}

// ContactsWithSameName returns all contacts that have the same Name (first name) as the given contact.
func (u *Utils) ContactsWithSameName(contacts []models.Contact, c *models.Contact) []models.Contact {
	if c == nil {
		return nil
	}
	nameLower := strings.ToLower(strings.TrimSpace(c.Name))
	var out []models.Contact
	for _, contact := range contacts {
		if strings.ToLower(strings.TrimSpace(contact.Name)) == nameLower {
			out = append(out, contact)
		}
	}
	return out
}

// ResolveContactByName returns a single email if exactly one contact matches; otherwise ("", false).
func (u *Utils) ResolveContactByName(contacts []models.Contact, name string) (string, bool) {
	matches := u.FindContactsByName(contacts, name)
	if len(matches) != 1 {
		return "", false
	}
	return matches[0].Email, true
}

// FormatContactOption returns "Name Lastname (email)" truncated to maxLen (e.g. for WhatsApp button/list limits).
func (u *Utils) FormatContactOption(c models.Contact, maxLen int) string {
	parts := strings.SplitN(c.Email, "@", 2)
	emailUser := c.Email
	if len(parts) > 0 && parts[0] != "" {
		emailUser = parts[0]
	}
	s := c.Name + " " + c.Lastname + " (" + emailUser + ")"
	if maxLen > 0 && len([]rune(s)) > maxLen {
		runes := []rune(s)
		if maxLen > 3 {
			s = string(runes[:maxLen-3]) + "..."
		} else {
			s = string(runes[:maxLen])
		}
	}
	return s
}

// GetFilenameFromMimeType returns a suggested filename extension for the given MIME type.
func (u *Utils) GetFilenameFromMimeType(mimeType string) string {
	mimeToExt := map[string]string{
		"image/jpeg":      ".jpg",
		"image/jpg":       ".jpg",
		"image/png":       ".png",
		"image/gif":       ".gif",
		"image/webp":      ".webp",
		"image/bmp":       ".bmp",
		"image/tiff":      ".tiff",
		"image/svg+xml":   ".svg",
		"application/pdf": ".pdf",
	}
	if ext, ok := mimeToExt[strings.ToLower(mimeType)]; ok {
		return "foto_whatsapp" + ext
	}
	return "foto_whatsapp.jpg"
}

// BuildEmailMessage construye el string de un mensaje de email con headers MIME
func (u *Utils) BuildEmailMessage(to, subject, body, contentType string) string {
	return fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: %s\r\n"+
		"\r\n"+
		"%s", to, subject, contentType, body)
}
