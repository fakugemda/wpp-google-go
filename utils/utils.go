package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
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

// CleanJSONString elimina los bloques de markdown ```json y ```
func (u *Utils) CleanJSONString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
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

// ResolveContactByName busca un email en un mapa de contactos
// Primero intenta coincidencia exacta, luego coincidencia parcial (case-insensitive)
func (u *Utils) ResolveContactByName(contacts map[string]string, name string) (string, bool) {
	nameLower := strings.ToLower(strings.TrimSpace(name))

	for contactName, email := range contacts {
		if strings.ToLower(contactName) == nameLower {
			return email, true
		}
	}

	for contactName, email := range contacts {
		if strings.Contains(strings.ToLower(contactName), nameLower) || strings.Contains(nameLower, strings.ToLower(contactName)) {
			return email, true
		}
	}

	return "", false
}

// BuildEmailMessage construye el string de un mensaje de email con headers MIME
func (u *Utils) BuildEmailMessage(to, subject, body, contentType string) string {
	return fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: %s\r\n"+
		"\r\n"+
		"%s", to, subject, contentType, body)
}
