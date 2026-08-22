package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

// contactDTO is the JSON shape contract, contact directory, and CRM record
// consumers rely on. The domain Contact struct stores transport-oriented
// fields (normalized/provider address) plus free-form metadata, so the
// display-facing fields name/number/email/tags are derived here.
type contactDTO struct {
	ID           uuid.UUID      `json:"id"`
	TenantID     uuid.UUID      `json:"tenant_id"`
	Name         string         `json:"name"`
	Number       string         `json:"number"`
	Email        string         `json:"email"`
	Tags         []string       `json:"tags"`
	IsGroup      bool           `json:"is_group"`
	DealStageKey string         `json:"deal_stage_key,omitempty"`
	DealStageID  *uuid.UUID     `json:"deal_stage_id,omitempty"`
	DealStage    *dealStageDTO  `json:"deal_stage,omitempty"`
	CustomValues map[string]any `json:"custom_values"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type dealStageDTO struct {
	ID     uuid.UUID `json:"id,omitempty"`
	Key    string    `json:"key"`
	Label  string    `json:"label"`
	Color  string    `json:"color"`
	Icon   string    `json:"icon"`
	IsWon  bool      `json:"is_won"`
	IsLost bool      `json:"is_lost"`
}

var defaultDealStages = map[string]dealStageDTO{
	"NEW_LEAD":              {Key: "NEW_LEAD", Label: "New Lead", Color: "#94a3b8", Icon: "user-plus"},
	"APPOINTMENT_SCHEDULED": {Key: "APPOINTMENT_SCHEDULED", Label: "Appointment Scheduled", Color: "#60a5fa", Icon: "calendar"},
	"HOT_LEAD":              {Key: "HOT_LEAD", Label: "Hot Lead", Color: "#f97316", Icon: "flame"},
	"COLD_LEAD":             {Key: "COLD_LEAD", Label: "Cold Lead", Color: "#64748b", Icon: "snowflake"},
	"IN_PROGRESS":           {Key: "IN_PROGRESS", Label: "In Progress", Color: "#a78bfa", Icon: "spinner"},
	"CLOSED_WON":            {Key: "CLOSED_WON", Label: "Closed Won", Color: "#22c55e", Icon: "trophy", IsWon: true},
	"CLOSED_LOST":           {Key: "CLOSED_LOST", Label: "Closed Lost", Color: "#ef4444", Icon: "x-circle", IsLost: true},
}

func toContactDTO(c domain.Contact) contactDTO {
	dto := contactDTO{
		ID:           c.ID,
		TenantID:     c.TenantID,
		Name:         c.DisplayName,
		Number:       c.ProviderAddress,
		IsGroup:      c.IsGroup,
		DealStageKey: c.DealStageKey,
		DealStageID:  c.DealStageID,
		CustomValues: map[string]any{},
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
	if dto.Name == "" {
		dto.Name = c.ProviderAddress
	}
	if c.DealStageKey != "" || c.DealStageID != nil {
		if stage, ok := defaultDealStages[c.DealStageKey]; ok {
			s := stage
			if c.DealStageID != nil {
				s.ID = *c.DealStageID
			}
			dto.DealStage = &s
		} else {
			s := dealStageDTO{Key: c.DealStageKey, Label: c.DealStageKey}
			if c.DealStageID != nil {
				s.ID = *c.DealStageID
			}
			dto.DealStage = &s
		}
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
		for k, v := range c.Metadata {
			if k != "email" && k != "tags" {
				dto.CustomValues[k] = v
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
