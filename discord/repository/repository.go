package repository

import (
	"sync"
	"whatsapp-gmail-bot/models"
)

type DraftRepository struct {
	drafts map[string]*models.AIResponse
	mu     sync.RWMutex
}

var (
	draftInstance *DraftRepository
	once          sync.Once
)

func GetDraftRepository() *DraftRepository {
	once.Do(func() {
		draftInstance = &DraftRepository{
			drafts: make(map[string]*models.AIResponse),
		}
	})
	return draftInstance
}

func (dr *DraftRepository) Store(userID string, draft *models.AIResponse) {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	dr.drafts[userID] = draft
}

func (dr *DraftRepository) Retrieve(userID string) (*models.AIResponse, bool) {
	dr.mu.RLock()
	defer dr.mu.RUnlock()
	draft, exists := dr.drafts[userID]
	return draft, exists
}

func (dr *DraftRepository) Delete(userID string) {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	delete(dr.drafts, userID)
}
