package config

import (
	"context"
	"fmt"
	"whatsapp-gmail-bot/store"
)

var Contacts = make(map[string]string)

func LoadContacts() {
	ctx, cancel := context.WithTimeout(context.Background(), store.DefaultTimeout)
	defer cancel()

	loaded, err := store.GetContacts(ctx)
	if err != nil {
		fmt.Printf("No se pudieron cargar contactos desde Redis: %v. El bot funcionará sin agenda.\n", err)
		return
	}

	Contacts = loaded
	fmt.Printf("Agenda cargada desde Redis: %d contactos disponibles.\n", len(Contacts))
}

// GetContactsPrompt - Convierte el mapa a texto para que la IA lo lea
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
