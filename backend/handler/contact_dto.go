package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
)

// contactDTO is the JSON shape contract, contact directory, and CRM record
// consumers rely on. The domain Contact struct stores transport-oriented
// fields (normalized/provider address) plus free-form metadata, so the
// display-facing fields name/number/email/tags are derived here.
type contactDTO struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Name      string    `json:"name"`
	Number    string    `json:"number"`
	Email     string    `json:"email"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toContactDTO(c domain.Contact) contactDTO {
	dto := contactDTO{
		ID:        c.ID,
		TenantID:  c.TenantID,
		Name:      c.DisplayName,
		Number:    c.ProviderAddress,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	if c.Metadata != nil {
		if email, ok := c.Metadata["email"].(string); ok {
			dto.Email = email
		}
		if rawTags, ok := c.Metadata["tags"].([]any); ok {
			for _, t := range rawTags {
				if tag, ok := t.(string); ok && tag != "" {
					dto.Tags = append(dto.Tags, tag)
				}
			}
		}
	}
	return dto
}

func toContactListDTO(contacts []domain.Contact) []contactDTO {
	result := make([]contactDTO, 0, len(contacts))
	for _, c := range contacts {
		result = append(result, toContactDTO(c))
	}
	return result
}