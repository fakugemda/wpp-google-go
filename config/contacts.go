package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var Contacts = make(map[string]string)

func LoadContacts() {
	var data []byte
	var err error

	// 1. Intentamos leer desde Variable de Entorno (Para Render)
	envData := os.Getenv("CONTACTS_JSON")
	if envData != "" {
		fmt.Println("☁️ Cargando contactos desde Variable de Entorno")
		data = []byte(envData)
	} else {
		// 2. Si no hay variable, leemos archivo local (Para tu PC)
		fmt.Println("🏠 Cargando contactos desde contacts.json")
		data, err = os.ReadFile("resources/contacts.json")
		if err != nil {
			fmt.Println("⚠️ No se encontró lista de contactos. El bot funcionará sin agenda.")
			return
		}
	}

	// 3. Convertimos el JSON al Mapa
	err = json.Unmarshal(data, &Contacts)
	if err != nil {
		fmt.Printf("❌ Error procesando JSON de contactos: %v\n", err)
		return
	}

	fmt.Printf("✅ Agenda cargada: %d contactos disponibles.\n", len(Contacts))
}

// GetContactsPrompt: Convierte el mapa a texto para que la IA lo lea
func GetContactsPrompt() string {
	if len(Contacts) == 0 {
		return ""
	}

	list := "LISTA DE CONTACTOS DE CONFIANZA:\n"
	for name, email := range Contacts {
		list += fmt.Sprintf("- Nombre/Apodo: %s | Email: %s\n", name, email)
	}
	return list
}
