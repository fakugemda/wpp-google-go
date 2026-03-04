package config

import (
	"encoding/json"
	"fmt"
	"os"
	"whatsapp-gmail-bot/models"
)

var Contacts []models.Contact

func LoadContacts() {
	envData := os.Getenv("CONTACTS_JSON")
	var data []byte
	var err error

	if envData != "" {
		fmt.Println("Cargando contactos desde Variable de Entorno")
		data = []byte(envData)
	} else {
		fmt.Println("Cargando contactos desde contacts.json")
		data, err = os.ReadFile("resources/contacts.json")
		if err != nil {
			fmt.Println("No se encontró lista de contactos. El bot funcionará sin agenda.")
			return
		}
	}

	err = json.Unmarshal(data, &Contacts)
	if err != nil {
		fmt.Printf("Error procesando JSON de contactos: %s\n", err.Error())
		return
	}

	fmt.Printf("Agenda cargada: %d contactos disponibles.\n", len(Contacts))
}

// GetContactsPrompt - Convierte la lista de contactos a texto para que la IA lo lea
func GetContactsPrompt() string {
	if len(Contacts) == 0 {
		return ""
	}

	list := "LISTA DE CONTACTOS DE CONFIANZA:\n"
	for _, c := range Contacts {
		list += fmt.Sprintf("- Nombre: %s %s | Email: %s\n", c.Name, c.Lastname, c.Email)
	}
	return list
}
