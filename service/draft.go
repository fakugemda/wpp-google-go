package service

import (
	"fmt"
	"strings"
	"whatsapp-gmail-bot/models"
	"whatsapp-gmail-bot/utils"
)

const (
	DraftActionCreateDraft       = "create_draft"
	DraftActionAskEmail          = "ask_email"
	DraftActionShowContactChoice = "show_contact_choice"
)

const maxContactChoiceCandidates = 10
const listTitleMaxLen = 24

var utilsDraft = utils.GetUtils()

// DraftRecipientResult es el resultado de procesar el destinatario de un borrador; el controller solo ejecuta la acción.
type DraftRecipientResult struct {
	Action          string
	Data            *models.AIResponse
	Message         string
	LogMessage      string
	BodyText        string
	Candidates      []models.Contact
	Buttons         []InteractiveButton
	Rows            []ListRow
	UseButtons      bool
	TruncateMessage string // si hay muchos candidatos, mensaje a enviar antes de las opciones
}

// ProcessDraftRecipient resuelve el destinatario y decide si crear borrador, pedir email o mostrar opciones de contacto.
func ProcessDraftRecipient(data *models.AIResponse, contacts []models.Contact) DraftRecipientResult {
	data.To = strings.TrimSpace(data.To)

	if data.To == "PENDIENTE" || data.To == "" {
		return DraftRecipientResult{
			Action:     DraftActionAskEmail,
			Message:    fmt.Sprintf("Entendido: '%s'\n\nPero... ¿A quién se lo mando? (Dime el email)", data.Subject),
			LogMessage: "To no resuelto.",
		}
	}

	if utilsDraft.IsValidEmail(data.To) {
		contact := utilsDraft.GetContactByEmail(contacts, data.To)
		if contact != nil {
			sameName := utilsDraft.ContactsWithSameName(contacts, contact)
			if len(sameName) > 1 {
				candidates, truncateMsg := limitCandidatesWithMessage(sameName)
				bodyText, buttons, rows, useButtons := buildContactChoiceContent(candidates)
				return DraftRecipientResult{
					Action:          DraftActionShowContactChoice,
					Data:            data,
					BodyText:        bodyText,
					Candidates:      candidates,
					Buttons:         buttons,
					Rows:            rows,
					UseButtons:      useButtons,
					TruncateMessage: truncateMsg,
					LogMessage:      "Varios contactos con ese nombre; mostrando opciones (esperando elección del usuario).",
				}
			}
		}
		return DraftRecipientResult{
			Action:     DraftActionCreateDraft,
			Data:       data,
			LogMessage: "To=" + data.To,
		}
	}

	matches := utilsDraft.FindContactsByName(contacts, data.To)
	switch len(matches) {
	case 0:
		return DraftRecipientResult{
			Action:     DraftActionAskEmail,
			Message:    fmt.Sprintf("Entendido: '%s'\n\nPero... ¿A quién se lo mando? (Dime el email)", data.Subject),
			LogMessage: "To no resuelto.",
		}
	case 1:
		data.To = matches[0].Email
		return DraftRecipientResult{
			Action:     DraftActionCreateDraft,
			Data:       data,
			LogMessage: "To=" + data.To,
		}
	default:
		candidates, truncateMsg := limitCandidatesWithMessage(matches)
		bodyText, buttons, rows, useButtons := buildContactChoiceContent(candidates)
		return DraftRecipientResult{
			Action:          DraftActionShowContactChoice,
			Data:            data,
			BodyText:        bodyText,
			Candidates:      candidates,
			Buttons:         buttons,
			Rows:            rows,
			UseButtons:      useButtons,
			TruncateMessage: truncateMsg,
			LogMessage:      "Varios contactos con ese nombre; mostrando opciones (esperando elección del usuario).",
		}
	}
}

func limitCandidatesWithMessage(c []models.Contact) ([]models.Contact, string) {
	if len(c) <= maxContactChoiceCandidates {
		return c, ""
	}
	return c[:maxContactChoiceCandidates], "Hay muchos contactos que coinciden. Mostrando los primeros 10; especificá más si no está quien buscás."
}

// buildContactChoiceContent arma siempre una lista (no botones) para que la descripción muestre el email completo (límite lista: título 24, descripción 72 chars).
func buildContactChoiceContent(candidates []models.Contact) (bodyText string, buttons []InteractiveButton, rows []ListRow, useButtons bool) {
	bodyText = "Encontré más de un contacto. Elegí a quién enviar el correo:"
	rows = make([]ListRow, 0, len(candidates))
	for _, c := range candidates {
		rows = append(rows, ListRow{
			ID:          c.Email,
			Title:       utilsDraft.FormatContactOption(c, listTitleMaxLen),
			Description: c.Email,
		})
	}
	return bodyText, nil, rows, false
}

// BuildDraftPreviewMessage arma el texto del preview del borrador para WhatsApp.
func BuildDraftPreviewMessage(data *models.AIResponse) string {
	return fmt.Sprintf("*Borrador IA Creado* ✉️\n\n*Para:* %s\n*Asunto:* %s\n\n%s",
		data.To, data.Subject, data.Content)
}
