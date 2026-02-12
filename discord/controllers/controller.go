package controllers

import (
	"fmt"
	"whatsapp-gmail-bot/discord/services"
	"whatsapp-gmail-bot/models"

	"github.com/bwmarrin/discordgo"
)

type DiscordController struct {
	DiscordService *services.DiscordService
	ContactsList   string
}

func NewDiscordController(contactsList string) *DiscordController {
	return &DiscordController{
		DiscordService: services.NewDiscordService(),
		ContactsList:   contactsList,
	}
}

func (dc *DiscordController) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if dc.shouldIgnoreMessage(s, m) {
		return
	}

	promptText := m.Content

	if dc.hasAttachments(m) {
		dc.handleAttachment(s, m)
		if promptText == "" {
			return
		}
	}

	dc.processMessage(s, m, promptText)
}

func (dc *DiscordController) shouldIgnoreMessage(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if m.Author.ID == s.State.User.ID {
		return true
	}

	if m.GuildID != "" {
		return true
	}

	return false
}

func (dc *DiscordController) hasAttachments(m *discordgo.MessageCreate) bool {
	return len(m.Attachments) > 0
}

func (dc *DiscordController) handleAttachment(s *discordgo.Session, m *discordgo.MessageCreate) {
	attachment := m.Attachments[0]

	dc.sendMessage(s, m.ChannelID, buildDownloadingMessage(attachment.Filename))

	err := dc.DiscordService.AttachmentService.ProcessAttachment(attachment, m.Author.ID, m.Author.Username)
	if err != nil {
		dc.sendMessage(s, m.ChannelID, buildAttachmentErrorMessage(err))
		return
	}

	dc.sendMessage(s, m.ChannelID, buildAttachmentSavedMessage(attachment.Filename))
}

func (dc *DiscordController) processMessage(s *discordgo.Session, m *discordgo.MessageCreate, promptText string) {
	dc.logIncomingMessage(m.Author.Username, promptText)

	aiResp, err := dc.DiscordService.ProcessIntent(promptText, dc.ContactsList)
	if err != nil {
		dc.sendMessage(s, m.ChannelID, buildProcessingErrorMessage(err))
		return
	}

	dc.dispatchResponse(s, m, aiResp)
}

func (dc *DiscordController) logIncomingMessage(username, message string) {
	fmt.Printf("💬 [Discord] %s: %s\n", username, message)
}

func (dc *DiscordController) dispatchResponse(s *discordgo.Session, m *discordgo.MessageCreate, aiResp *models.AIResponse) {
	switch aiResp.Type {
	case "CHAT":
		dc.handleChatResponse(s, m, aiResp)

	case "EMAIL_DRAFT":
		dc.handleEmailDraftResponse(s, m, aiResp)

	case "CONFIRM_SEND":
		dc.handleConfirmSendResponse(s, m)
	}
}

func (dc *DiscordController) handleChatResponse(s *discordgo.Session, m *discordgo.MessageCreate, aiResp *models.AIResponse) {
	dc.sendMessage(s, m.ChannelID, aiResp.Content)
}

func (dc *DiscordController) handleEmailDraftResponse(s *discordgo.Session, m *discordgo.MessageCreate, aiResp *models.AIResponse) {
	dc.DiscordService.StoreDraft(m.Author.ID, aiResp)

	hasAttachment := dc.DiscordService.AttachmentService.HasAttachmentInCache(m.Author.ID)
	if hasAttachment {
		dc.DiscordService.AttachmentService.LogAttachmentInDraft(m.Author.ID)
	}

	message := buildDraftPreviewMessage(aiResp, hasAttachment)
	dc.sendMessage(s, m.ChannelID, message)
}

func (dc *DiscordController) handleConfirmSendResponse(s *discordgo.Session, m *discordgo.MessageCreate) {
	draft, exists := dc.DiscordService.RetrieveDraft(m.Author.ID)
	if !exists {
		dc.sendMessage(s, m.ChannelID, buildNoDraftMessage())
		return
	}

	dc.sendMessage(s, m.ChannelID, buildSendingEmailMessage())

	attachmentData, filename, _ := dc.DiscordService.AttachmentService.RetrieveFromCache(m.Author.ID)

	err := dc.DiscordService.SendEmail(draft.To, draft.Subject, draft.Content, attachmentData, filename)
	if err != nil {
		dc.sendMessage(s, m.ChannelID, buildEmailErrorMessage(err))
		return
	}

	dc.sendMessage(s, m.ChannelID, buildEmailSentMessage())
	dc.DiscordService.DeleteDraft(m.Author.ID)
	dc.DiscordService.AttachmentService.DeleteFromCache(m.Author.ID)
}

func (dc *DiscordController) sendMessage(s *discordgo.Session, channelID, message string) {
	s.ChannelMessageSend(channelID, message)
}
