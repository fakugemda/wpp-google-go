package services

import (
	"fmt"
	"whatsapp-gmail-bot/discord/repository"
	"whatsapp-gmail-bot/models"
	"whatsapp-gmail-bot/service"
)

type DiscordService struct {
	DraftRepository   *repository.DraftRepository
	AttachmentService *AttachmentService
}

func NewDiscordService() *DiscordService {
	return &DiscordService{
		DraftRepository:   repository.GetDraftRepository(),
		AttachmentService: &AttachmentService{},
	}
}

func (ds *DiscordService) ProcessIntent(userMessage, contactsList string) (*models.AIResponse, error) {
	return service.ProcessIntent(userMessage, contactsList)
}

func (ds *DiscordService) StoreDraft(userID string, draft *models.AIResponse) {
	ds.DraftRepository.Store(userID, draft)
}

func (ds *DiscordService) RetrieveDraft(userID string) (*models.AIResponse, bool) {
	return ds.DraftRepository.Retrieve(userID)
}

func (ds *DiscordService) DeleteDraft(userID string) {
	ds.DraftRepository.Delete(userID)
}

func (ds *DiscordService) SendEmail(to, subject, body string, attachmentData []byte, filename string) error {
	draftID, err := service.CreateDraft(to, subject, body, attachmentData, filename)
	if err != nil {
		return fmt.Errorf("error creando borrador: %s", err.Error())
	}

	err = service.SendDraft(draftID)
	if err != nil {
		return fmt.Errorf("error enviando borrador: %s", err.Error())
	}

	return nil
}
