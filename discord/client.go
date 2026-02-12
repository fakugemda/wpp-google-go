package discord

import (
	"fmt"
	"os"
	"whatsapp-gmail-bot/discord/controllers"
	"whatsapp-gmail-bot/service"

	"github.com/bwmarrin/discordgo"
)

func StartService(contactsList string) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("⚠️ DISCORD_TOKEN no configurado. Saltando inicio de Discord.")
		return
	}

	session, err := createDiscordSession(token)
	if err != nil {
		fmt.Println("❌ Error creando sesión Discord:", err)
		return
	}

	controller := controllers.NewDiscordController(contactsList)
	registerHandlers(session, controller)
	configureIntents(session)
	startCacheCleanup()

	err = openConnection(session)
	if err != nil {
		fmt.Println("❌ Error conectando a Discord WebSocket:", err)
		return
	}

	fmt.Println("🤖 Servicio Discord ONLINE (Escuchando DMs)")
}

func createDiscordSession(token string) (*discordgo.Session, error) {
	return discordgo.New("Bot " + token)
}

func registerHandlers(session *discordgo.Session, controller *controllers.DiscordController) {
	session.AddHandler(controller.HandleMessage)
}

func configureIntents(session *discordgo.Session) {
	session.Identify.Intents = discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent
}

func startCacheCleanup() {
	service.StartImageCacheCleanup()
}

func openConnection(session *discordgo.Session) error {
	return session.Open()
}
