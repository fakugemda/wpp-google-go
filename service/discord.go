package service

import (
	"fmt"
	"os"
	"whatsapp-gmail-bot/models"

	"github.com/bwmarrin/discordgo"
)

var (
	discordDrafts   = make(map[string]*models.AIResponse)
	discordContacts string
)

// StartDiscordService inicializa la conexión y se queda escuchando
func StartDiscordService(contactsList string) {
	discordContacts = contactsList
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("⚠️ DISCORD_TOKEN no configurado. Saltando inicio de Discord.")
		return
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("❌ Error creando sesión Discord:", err)
		return
	}

	dg.AddHandler(discordMessageHandler)
	dg.Identify.Intents = discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

	// Abrir conexión
	err = dg.Open()
	if err != nil {
		fmt.Println("❌ Error conectando a Discord WebSocket:", err)
		return
	}

	fmt.Println("🤖 Servicio Discord ONLINE (Escuchando DMs)")
}

// discordMessageHandler maneja cada mensaje que entra
func discordMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.GuildID != "" {
		return
	}

	fmt.Printf("💬 [Discord] %s: %s\n", m.Author.Username, m.Content)

	aiResp, err := ProcessIntent(m.Content, discordContacts)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "⚠️ Error procesando solicitud: "+err.Error())
		return
	}

	switch aiResp.Type {
	case "CHAT":
		s.ChannelMessageSend(m.ChannelID, aiResp.Content)

	case "EMAIL_DRAFT":
		discordDrafts[m.Author.ID] = aiResp

		msg := fmt.Sprintf("**📝 Borrador Generado**\n\n**Para:** `%s`\n**Asunto:** `%s`\n\n%s\n\n_Escribe 'sí' o 'envíalo' para confirmar._",
			aiResp.To, aiResp.Subject, aiResp.Content)
		s.ChannelMessageSend(m.ChannelID, msg)

	case "CONFIRM_SEND":
		draft, exists := discordDrafts[m.Author.ID]
		if !exists {
			s.ChannelMessageSend(m.ChannelID, "🤷‍♂️ No tengo ningún mail pendiente. Pídeme redactar uno primero.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "🚀 Enviando correo...")

		err := SendEmail(draft.To, draft.Subject, draft.Content)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Error enviando: "+err.Error())
		} else {
			s.ChannelMessageSend(m.ChannelID, "✅ ¡Correo enviado exitosamente!")
			delete(discordDrafts, m.Author.ID) // Limpiar memoria
		}
	}
}

// SendEmail crea un borrador y lo envía inmediatamente
func SendEmail(to, subject, body string) error {
	draftID, err := CreateDraft(to, subject, body, nil, "")
	if err != nil {
		return fmt.Errorf("error creando borrador: %s", err.Error())
	}

	err = SendDraft(draftID)
	if err != nil {
		return fmt.Errorf("error enviando borrador: %s", err.Error())
	}

	return nil
}
