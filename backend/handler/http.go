package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdp/qrterminal"
	"github.com/raufhm/whops/domain"
	"github.com/raufhm/whops/internal/storage"
	"github.com/raufhm/whops/whatsapp"
	"go.mau.fi/whatsmeow"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Server struct {
	Manager     *whatsapp.WhatsAppManager
	Platform    domain.PlatformRepository
	MediaStore  storage.MediaStore
	S3ObjectURL string
	Auth        sessionAuth
}

type sessionAuth interface {
	GetSessionByID(id uuid.UUID) (domain.Session, error)
	GetOperatorByID(tenantID, operatorID uuid.UUID) (domain.Operator, error)
}

type apiError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type accountHostResolver interface {
	AccountHost(uuid.UUID, string) (string, error)
}

type contactConversationTimelineReader interface {
	GetContactConversationTimeline(tenantID, contactID uuid.UUID, limit, offset int) ([]domain.ConversationMessage, error)
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, apiError{Error: message, Code: code})
}

func (s *Server) tenant(r *http.Request) (uuid.UUID, bool) {
	key := r.Header.Get("X-API-Key")
	if key == "" {
		key = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}
	if key == "" {
		key = r.URL.Query().Get("token")
	}
	if key == "" {
		key = r.URL.Query().Get("api_key")
	}
	if key != "" {
		if s.Platform == nil {
			return uuid.Nil, false
		}
		id, err := s.Platform.AuthenticateAPIKey(key)
		return id, err == nil && id != uuid.Nil
	}
	if s.Auth != nil {
		if c, err := r.Cookie(sessionCookieName); err == nil && c.Value != "" {
			if sid, err := uuid.Parse(c.Value); err == nil && sid != uuid.Nil {
				if sess, err := s.Auth.GetSessionByID(sid); err == nil && sess.TenantID != uuid.Nil {
					return sess.TenantID, true
				}
			}
		}
	}
	return uuid.Nil, false
}
func page(r *http.Request) (int, int, bool) {
	limit, offset := 50, 0
	var err error
	if v := r.URL.Query().Get("limit"); v != "" {
		limit, err = strconv.Atoi(v)
	}
	if err != nil || limit < 1 || limit > 1000 {
		return 0, 0, false
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		offset, err = strconv.Atoi(v)
	}
	return limit, offset, err == nil && offset >= 0
}

// APIHandler is the versioned, authenticated application API. Legacy handlers
// remain registered separately for backwards compatibility.
func (s *Server) APIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost && r.Method != http.MethodPatch && r.Method != http.MethodPut && r.Method != http.MethodDelete {
		writeAPIError(w, 405, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	tenant, ok := s.tenant(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "valid API key required")
		return
	}

	if s.Platform != nil {
		quota, err := s.Platform.GetQuota(tenant)
		if err == nil && quota.CurrentUsage >= quota.MonthlyLimit {
			writeAPIError(w, 429, "QUOTA_EXCEEDED", "quota exceeded")
			return
		}
	}

	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/"), "/")
	parts := strings.Split(path, "/")
	if path == "" {
		writeAPIError(w, 404, "NOT_FOUND", "resource not found")
		return
	}
	limit, offset, valid := page(r)
	if !valid {
		writeAPIError(w, 400, "INVALID_PAGINATION", "limit must be between 1 and 100 and offset non-negative")
		return
	}
	switch parts[0] {
	case "accounts":
		if len(parts) == 1 {
			if r.Method == http.MethodGet {
				v, err := s.Platform.ListAccounts(tenant)
				if err != nil {
					writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
					return
				}
				WriteJSON(w, 200, v)
				return
			}
			writeAPIError(w, 405, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}
		if len(parts) == 3 && parts[2] == "messages" && r.Method == http.MethodPost {
			host := parts[1]
			if resolver, ok := s.Platform.(accountHostResolver); ok {
				var err error
				host, err = resolver.AccountHost(tenant, parts[1])
				if err != nil {
					writeAPIError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "account not found")
					return
				}
			}
			s.sendAPI(w, r, host)
			return
		}
	case "pipelines":
		if len(parts) == 1 && r.Method == http.MethodGet {
			var isActive *bool
			if actStr := r.URL.Query().Get("is_active"); actStr != "" {
				actVal := actStr == "true" || actStr == "1"
				isActive = &actVal
			}
			pipelines, err := s.Platform.ListPipelines(tenant, isActive)
			if err != nil {
				writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
				return
			}
			WriteJSON(w, 200, pipelines)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodPost {
			var req struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				IsDefault   bool   `json:"is_default"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeAPIError(w, 400, "INVALID_BODY", "invalid request body")
				return
			}
			if strings.TrimSpace(req.Name) == "" {
				writeAPIError(w, 400, "INVALID_BODY", "pipeline name is required")
				return
			}
			pipeline, err := s.Platform.CreatePipeline(tenant, req.Name, req.Description, req.IsDefault)
			if err != nil {
				if errors.Is(err, domain.ErrPipelineNameExists) {
					writeAPIError(w, 409, "DUPLICATE_NAME", "pipeline with this name already exists")
					return
				}
				writeAPIError(w, 500, "INTERNAL_ERROR", "failed to create pipeline")
				return
			}
			WriteJSON(w, 201, pipeline)
			return
		}
		if len(parts) == 2 {
			id, err := uuid.Parse(parts[1])
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid pipeline id")
				return
			}
			if r.Method == http.MethodGet {
				pipeline, err := s.Platform.GetPipeline(tenant, id)
				if err != nil {
					if errors.Is(err, domain.ErrPipelineNotFound) {
						writeAPIError(w, 404, "NOT_FOUND", "pipeline not found")
						return
					}
					writeAPIError(w, 500, "INTERNAL_ERROR", "failed to get pipeline")
					return
				}
				WriteJSON(w, 200, pipeline)
				return
			}
			if r.Method == http.MethodPatch || r.Method == http.MethodPut {
				var req struct {
					Name        *string `json:"name"`
					Description *string `json:"description"`
					IsDefault   *bool   `json:"is_default"`
					IsActive    *bool   `json:"is_active"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeAPIError(w, 400, "INVALID_BODY", "invalid request body")
					return
				}
				pipeline, err := s.Platform.UpdatePipeline(tenant, id, req.Name, req.Description, req.IsDefault, req.IsActive)
				if err != nil {
					if errors.Is(err, domain.ErrPipelineNotFound) {
						writeAPIError(w, 404, "NOT_FOUND", "pipeline not found")
						return
					}
					if errors.Is(err, domain.ErrPipelineNameExists) {
						writeAPIError(w, 409, "DUPLICATE_NAME", "pipeline with this name already exists")
						return
					}
					if errors.Is(err, domain.ErrDefaultPipelineCannotBeInactive) || errors.Is(err, domain.ErrCannotDeleteDefaultPipeline) {
						writeAPIError(w, 400, "INVALID_REQUEST", err.Error())
						return
					}
					writeAPIError(w, 500, "INTERNAL_ERROR", "failed to update pipeline")
					return
				}
				WriteJSON(w, 200, pipeline)
				return
			}
			if r.Method == http.MethodDelete {
				if err := s.Platform.DeletePipeline(tenant, id); err != nil {
					if errors.Is(err, domain.ErrPipelineNotFound) {
						writeAPIError(w, 404, "NOT_FOUND", "pipeline not found")
						return
					}
					if errors.Is(err, domain.ErrCannotDeleteDefaultPipeline) {
						writeAPIError(w, 400, "CANNOT_DELETE_DEFAULT", "cannot delete default pipeline")
						return
					}
					if errors.Is(err, domain.ErrPipelineContainsStages) {
						writeAPIError(w, 409, "PIPELINE_NOT_EMPTY", "cannot delete pipeline containing stages")
						return
					}
					writeAPIError(w, 500, "INTERNAL_ERROR", "failed to delete pipeline")
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
	case "deal-stages":
		if len(parts) == 1 && r.Method == http.MethodGet {
			var pipelineID *uuid.UUID
			if plStr := r.URL.Query().Get("pipeline_id"); plStr != "" {
				if plUUID, err := uuid.Parse(plStr); err == nil {
					pipelineID = &plUUID
				}
			}
			var isActive *bool
			if actStr := r.URL.Query().Get("is_active"); actStr != "" {
				actVal := actStr == "true" || actStr == "1"
				isActive = &actVal
			}
			stages, err := s.Platform.ListDealStages(tenant, pipelineID, isActive)
			if err != nil {
				writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
				return
			}
			WriteJSON(w, 200, stages)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodPost {
			var req struct {
				PipelineID *string `json:"pipeline_id,omitempty"`
				Key        string  `json:"key"`
				Label      string  `json:"label"`
				Color      string  `json:"color"`
				Icon       string  `json:"icon"`
				SortOrder  int     `json:"sort_order"`
				IsWon      bool    `json:"is_won"`
				IsLost     bool    `json:"is_lost"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeAPIError(w, 400, "INVALID_BODY", "invalid request body")
				return
			}
			if strings.TrimSpace(req.Key) == "" || strings.TrimSpace(req.Label) == "" {
				writeAPIError(w, 400, "INVALID_BODY", "key and label are required")
				return
			}
			var plUUID *uuid.UUID
			if req.PipelineID != nil && *req.PipelineID != "" {
				parsed, err := uuid.Parse(*req.PipelineID)
				if err != nil {
					writeAPIError(w, 400, "INVALID_ID", "invalid pipeline_id")
					return
				}
				plUUID = &parsed
			}
			stage, err := s.Platform.CreateDealStage(tenant, plUUID, req.Key, req.Label, req.Color, req.Icon, req.SortOrder, req.IsWon, req.IsLost)
			if err != nil {
				if errors.Is(err, domain.ErrPipelineNotFound) {
					writeAPIError(w, 404, "NOT_FOUND", "pipeline not found")
					return
				}
				writeAPIError(w, 500, "INTERNAL_ERROR", "failed to create deal stage")
				return
			}
			WriteJSON(w, 201, stage)
			return
		}
		if len(parts) == 2 {
			id, err := uuid.Parse(parts[1])
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid deal stage id")
				return
			}
			if r.Method == http.MethodGet {
				stage, err := s.Platform.GetDealStage(tenant, id)
				if err != nil {
					if errors.Is(err, domain.ErrStageNotFound) {
						writeAPIError(w, 404, "NOT_FOUND", "deal stage not found")
						return
					}
					writeAPIError(w, 500, "INTERNAL_ERROR", "failed to get deal stage")
					return
				}
				WriteJSON(w, 200, stage)
				return
			}
			if r.Method == http.MethodPatch || r.Method == http.MethodPut {
				var req struct {
					PipelineID *string `json:"pipeline_id"`
					Label      *string `json:"label"`
					Color      *string `json:"color"`
					Icon       *string `json:"icon"`
					SortOrder  *int    `json:"sort_order"`
					IsActive   *bool   `json:"is_active"`
					IsWon      *bool   `json:"is_won"`
					IsLost     *bool   `json:"is_lost"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeAPIError(w, 400, "INVALID_BODY", "invalid request body")
					return
				}
				var plUUID *uuid.UUID
				if req.PipelineID != nil && *req.PipelineID != "" {
					parsed, err := uuid.Parse(*req.PipelineID)
					if err != nil {
						writeAPIError(w, 400, "INVALID_ID", "invalid pipeline_id")
						return
					}
					plUUID = &parsed
				}
				stage, err := s.Platform.UpdateDealStage(tenant, id, plUUID, req.Label, req.Color, req.Icon, req.SortOrder, req.IsActive, req.IsWon, req.IsLost)
				if err != nil {
					if errors.Is(err, domain.ErrStageNotFound) {
						writeAPIError(w, 404, "NOT_FOUND", "deal stage not found")
						return
					}
					if errors.Is(err, domain.ErrPipelineNotFound) {
						writeAPIError(w, 404, "NOT_FOUND", "pipeline not found")
						return
					}
					writeAPIError(w, 500, "INTERNAL_ERROR", "failed to update deal stage")
					return
				}
				WriteJSON(w, 200, stage)
				return
			}
			if r.Method == http.MethodDelete {
				if err := s.Platform.DeleteDealStage(tenant, id); err != nil {
					if errors.Is(err, domain.ErrStageNotFound) {
						writeAPIError(w, 404, "NOT_FOUND", "deal stage not found")
						return
					}
					if errors.Is(err, domain.ErrStageAssignedToContacts) {
						writeAPIError(w, 409, "STAGE_ASSIGNED_TO_CONTACTS", "stage is currently assigned to contacts")
						return
					}
					writeAPIError(w, 500, "INTERNAL_ERROR", "failed to delete deal stage")
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
	case "contacts":
		if len(parts) == 1 && r.Method == http.MethodGet {
			search := r.URL.Query().Get("q")
			v, total, err := s.Platform.ListContacts(tenant, limit, offset, search)
			if err != nil {
				writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
				return
			}
			WriteJSON(w, 200, map[string]any{
				"items":  toContactListDTO(v),
				"total":  total,
				"limit":  limit,
				"offset": offset,
			})
			return
		}
		if len(parts) == 2 && r.Method == http.MethodGet {
			id, err := uuid.Parse(parts[1])
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid contact id")
				return
			}
			v, err := s.Platform.GetContact(tenant, id)
			if err != nil {
				writeAPIError(w, 404, "NOT_FOUND", "contact not found")
				return
			}
			WriteJSON(w, 200, toContactDTO(v))
			return
		}
		if len(parts) == 2 && r.Method == http.MethodPatch {
			id, err := uuid.Parse(parts[1])
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid contact id")
				return
			}
			if _, err := s.Platform.GetContact(tenant, id); err != nil {
				writeAPIError(w, 404, "NOT_FOUND", "contact not found")
				return
			}
			var req struct {
				Name         string         `json:"name"`
				Email        string         `json:"email"`
				Tags         []string       `json:"tags"`
				CustomValues map[string]any `json:"custom_values"`
				DealStageKey *string        `json:"deal_stage_key,omitempty"`
				DealStageID  *string        `json:"deal_stage_id,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeAPIError(w, 400, "INVALID_BODY", "invalid request body")
				return
			}
			var dealStageKey string
			var dealStageID *uuid.UUID
			var clearDealStage bool
			if req.DealStageID != nil {
				if *req.DealStageID == "" {
					clearDealStage = true
				} else {
					parsed, err := uuid.Parse(*req.DealStageID)
					if err != nil {
						writeAPIError(w, 400, "INVALID_ID", "invalid deal_stage_id")
						return
					}
					dealStageID = &parsed
				}
			} else if req.DealStageKey != nil {
				if *req.DealStageKey == "" {
					clearDealStage = true
				} else {
					dealStageKey = *req.DealStageKey
				}
			}
			v, err := s.Platform.UpdateContact(tenant, id, domain.ContactUpdateInput{
				DisplayName:    req.Name,
				Email:          req.Email,
				Tags:           req.Tags,
				CustomValues:   req.CustomValues,
				DealStageKey:   dealStageKey,
				DealStageID:    dealStageID,
				ClearDealStage: clearDealStage,
			})
			if err != nil {
				writeAPIError(w, 500, "INTERNAL_ERROR", "failed to update contact")
				return
			}
			WriteJSON(w, 200, toContactDTO(v))
			return
		}
		if len(parts) == 3 && parts[2] == "activities" {
			id, err := uuid.Parse(parts[1])
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid contact id")
				return
			}
			if _, err := s.Platform.GetContact(tenant, id); err != nil {
				writeAPIError(w, 404, "NOT_FOUND", "contact not found")
				return
			}
			if r.Method == http.MethodGet {
				v, err := s.Platform.ListContactActivities(tenant, id, limit, offset)
				if err != nil {
					writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
					return
				}
				WriteJSON(w, 200, v)
				return
			}
			if r.Method == http.MethodPost {
				s.createContactActivity(w, r, tenant, id)
				return
			}
		}
		if len(parts) == 3 && parts[2] == "conversations" && r.Method == http.MethodGet {
			id, err := uuid.Parse(parts[1])
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid contact id")
				return
			}
			if _, err := s.Platform.GetContact(tenant, id); err != nil {
				writeAPIError(w, 404, "NOT_FOUND", "contact not found")
				return
			}
			v, err := s.Platform.ListContactConversations(tenant, id, limit, offset)
			if err != nil {
				writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
				return
			}
			WriteJSON(w, 200, v)
			return
		}
		if len(parts) == 3 && parts[2] == "deal-history" {
			id, err := uuid.Parse(parts[1])
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid contact id")
				return
			}
			if _, err := s.Platform.GetContact(tenant, id); err != nil {
				writeAPIError(w, 404, "NOT_FOUND", "contact not found")
				return
			}
			dhLimit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
			if dhLimit <= 0 {
				dhLimit = 50
			}
			dhOffset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
			history, err := s.Platform.ListDealStageHistory(tenant, id, dhLimit, dhOffset)
			if err != nil {
				writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
				return
			}
			WriteJSON(w, 200, history)
			return
		}
		if len(parts) == 3 && (parts[2] == "deal-stage" || parts[2] == "move-stage" || parts[2] == "move_stage") && r.Method == http.MethodPost {
			id, err := uuid.Parse(parts[1])
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid contact id")
				return
			}
			if _, err := s.Platform.GetContact(tenant, id); err != nil {
				writeAPIError(w, 404, "NOT_FOUND", "contact not found")
				return
			}
			var req struct {
				StageKey    string  `json:"stage_key"`
				StageID     *string `json:"stage_id,omitempty"`
				DealStageID *string `json:"deal_stage_id,omitempty"`
				Note        string  `json:"note,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeAPIError(w, 400, "INVALID_BODY", "invalid request body")
				return
			}
			var stageUUID *uuid.UUID
			if req.DealStageID != nil && *req.DealStageID != "" {
				parsed, err := uuid.Parse(*req.DealStageID)
				if err != nil {
					writeAPIError(w, 400, "INVALID_ID", "invalid deal_stage_id")
					return
				}
				stageUUID = &parsed
			} else if req.StageID != nil && *req.StageID != "" {
				parsed, err := uuid.Parse(*req.StageID)
				if err != nil {
					writeAPIError(w, 400, "INVALID_ID", "invalid stage_id")
					return
				}
				stageUUID = &parsed
			}
			opID := uuid.Nil
			if opStr := s.operatorID(r); opStr != "" {
				if parsed, err := uuid.Parse(opStr); err == nil {
					opID = parsed
				}
			}
			transition, err := s.Platform.MoveContactToStage(tenant, id, req.StageKey, stageUUID, req.Note, opID)
			if err != nil {
				if errors.Is(err, domain.ErrStageNotFound) {
					writeAPIError(w, 404, "NOT_FOUND", "deal stage not found")
					return
				}
				writeAPIError(w, 500, "INTERNAL_ERROR", "failed to move deal stage")
				return
			}
			WriteJSON(w, 200, transition)
			return
		}
	case "contact-field-definitions":
		if len(parts) == 1 {
			if r.Method == http.MethodGet {
				v, err := s.Platform.ListContactFieldDefinitions(tenant)
				if err != nil {
					writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
					return
				}
				WriteJSON(w, 200, v)
				return
			}
			if r.Method == http.MethodPost {
				var req struct {
					Key        string   `json:"key"`
					Label      string   `json:"label"`
					FieldType  string   `json:"field_type"`
					Options    []string `json:"options"`
					IsRequired bool     `json:"is_required"`
					SortOrder  int      `json:"sort_order"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeAPIError(w, 400, "INVALID_BODY", "invalid request body")
					return
				}
				v, err := s.Platform.CreateContactFieldDefinition(tenant, req.Key, req.Label, req.FieldType, req.Options, req.IsRequired, req.SortOrder)
				if err != nil {
					writeAPIError(w, 500, "INTERNAL_ERROR", "failed to create field")
					return
				}
				WriteJSON(w, 201, v)
				return
			}
		}
		if len(parts) == 2 {
			id, err := uuid.Parse(parts[1])
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid field id")
				return
			}
			if r.Method == http.MethodGet {
				v, err := s.Platform.GetContactFieldDefinition(tenant, id)
				if err != nil {
					writeAPIError(w, 404, "NOT_FOUND", "field not found")
					return
				}
				WriteJSON(w, 200, v)
				return
			}
			if r.Method == http.MethodPatch || r.Method == http.MethodPut {
				var req struct {
					Label      string   `json:"label"`
					FieldType  string   `json:"field_type"`
					Options    []string `json:"options"`
					IsRequired bool     `json:"is_required"`
					SortOrder  int      `json:"sort_order"`
					IsActive   bool     `json:"is_active"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					writeAPIError(w, 400, "INVALID_BODY", "invalid request body")
					return
				}
				v, err := s.Platform.UpdateContactFieldDefinition(tenant, id, req.Label, req.FieldType, req.Options, req.IsRequired, req.SortOrder, req.IsActive)
				if err != nil {
					writeAPIError(w, 500, "INTERNAL_ERROR", "failed to update field")
					return
				}
				WriteJSON(w, 200, v)
				return
			}
			if r.Method == http.MethodDelete {
				if err := s.Platform.DeleteContactFieldDefinition(tenant, id); err != nil {
					writeAPIError(w, 500, "INTERNAL_ERROR", "failed to delete field")
					return
				}
				WriteJSON(w, 204, nil)
				return
			}
		}
	case "conversations", "tickets":
		if r.Method == http.MethodGet {
			status := r.URL.Query().Get("status")
			v, err := s.Platform.ListConversationSummaries(tenant, status, limit, offset)
			if err != nil {
				log.Printf("API conversations list error tenant=%s status=%q limit=%d offset=%d: %v", tenant, status, limit, offset, err)
				writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
				return
			}
			if len(parts) == 1 {
				WriteJSON(w, 200, v)
				return
			}
			if len(parts) == 2 || (len(parts) == 3 && parts[2] == "messages") {
				for _, c := range v {
					if strconv.FormatInt(c.TicketNumber, 10) == parts[1] || c.ID.String() == parts[1] {
						var msgs []domain.ConversationMessage
						var e error
						if history, ok := s.Platform.(contactConversationTimelineReader); ok {
							msgs, e = history.GetContactConversationTimeline(tenant, c.ContactID, limit, offset)
						} else {
							msgs, e = s.Platform.GetConversationTimeline(tenant, c.ID, limit, offset)
						}
						if e != nil {
							log.Printf("API conversation timeline error tenant=%s conversation=%s limit=%d offset=%d: %v", tenant, c.ID, limit, offset, e)
							writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
							return
						}
						WriteJSON(w, 200, map[string]any{"conversation": c, "messages": msgs})
						return
					}
				}
				if convID, err := uuid.Parse(parts[1]); err == nil {
					if conv, err := s.Platform.GetConversation(tenant, convID); err == nil {
						var msgs []domain.ConversationMessage
						var e error
						if history, ok := s.Platform.(contactConversationTimelineReader); ok {
							msgs, e = history.GetContactConversationTimeline(tenant, conv.ContactID, limit, offset)
						} else {
							msgs, e = s.Platform.GetConversationTimeline(tenant, conv.ID, limit, offset)
						}
						if e != nil {
							log.Printf("API conversation timeline error tenant=%s conversation=%s limit=%d offset=%d: %v", tenant, conv.ID, limit, offset, e)
							writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
							return
						}
						summary := domain.ConversationSummary{
							Conversation: conv,
						}
						if contact, err := s.Platform.GetContact(tenant, conv.ContactID); err == nil {
							summary.ContactName = contact.DisplayName
							summary.ContactNumber = contact.ProviderAddress
							summary.IsGroup = contact.IsGroup
						}
						WriteJSON(w, 200, map[string]any{"conversation": summary, "messages": msgs})
						return
					}
				}
				writeAPIError(w, 404, "NOT_FOUND", "conversation not found")
				return
			}
		}
	case "activities":
		if len(parts) == 1 && r.Method == http.MethodGet {
			v, err := s.Platform.ListActivities(tenant, r.URL.Query().Get("status"), limit, offset)
			if err != nil {
				writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
				return
			}
			WriteJSON(w, 200, v)
			return
		}
		if len(parts) == 3 && parts[2] == "acknowledge" && r.Method == http.MethodPost {
			id, err := uuid.Parse(parts[1])
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid activity id")
				return
			}
			v, err := s.Platform.AcknowledgeActivity(tenant, id, s.operatorID(r), time.Now().UTC())
			if err != nil {
				writeAPIError(w, 404, "NOT_FOUND", "activity not found")
				return
			}
			WriteJSON(w, 200, v)
			return
		}
	case "bot-rules":
		// GET /api/v1/bot-rules - list all ruleset versions
		// POST /api/v1/bot-rules - create a new ruleset version
		// GET /api/v1/bot-rules/active - get the active ruleset
		// POST /api/v1/bot-rules/activate?version=N - activate a version
		if len(parts) == 1 && r.Method == http.MethodGet {
			s.ListBotRuleSetsHandler(w, r)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodPost {
			s.CreateBotRuleSetHandler(w, r)
			return
		}
		if len(parts) == 2 && parts[1] == "active" && r.Method == http.MethodGet {
			s.GetActiveBotRuleSetHandler(w, r)
			return
		}
		if len(parts) == 2 && parts[1] == "activate" && r.Method == http.MethodPost {
			s.ActivateBotRuleSetHandler(w, r)
			return
		}
	case "operator":
		// POST /api/v1/operator/assign?id=<conv> - assign operator
		// POST /api/v1/operator/handoff?id=<conv> - handoff
		// POST /api/v1/operator/close?id=<conv> - close with reason
		// POST /api/v1/operator/reopen?id=<conv> - reopen
		// DELETE /api/v1/operator/delete?id=<conv> - permanently delete
		if len(parts) == 2 && ((r.Method == http.MethodPost) || (r.Method == http.MethodDelete && parts[1] == "delete")) {
			switch parts[1] {
			case "assign":
				s.AssignConversationHandler(w, r)
				return
			case "handoff":
				s.HandoffConversationHandler(w, r)
				return
			case "close":
				s.CloseConversationHandler(w, r)
				return
			case "reopen":
				s.ReopenConversationHandler(w, r)
				return
			case "delete":
				s.DeleteConversationHandler(w, r)
				return
			}
		}
	case "notes":
		// POST /api/v1/notes?id=<conv> - add internal note to conversation
		if len(parts) == 1 && r.Method == http.MethodPost {
			s.AddInternalNoteHandler(w, r)
			return
		}
	case "merge":
		// POST /api/v1/merge - merge source conversation into target
		if len(parts) == 1 && r.Method == http.MethodPost {
			s.MergeConversationsHandler(w, r)
			return
		}
	case "split":
		// POST /api/v1/split?id=<source> - split conversation by message ids
		if len(parts) == 1 && r.Method == http.MethodPost {
			s.SplitConversationHandler(w, r)
			return
		}
	case "audit-logs":
		// GET /api/v1/audit-logs?limit=&offset= - list operator audit logs
		if len(parts) == 1 && r.Method == http.MethodGet {
			s.ListOperatorAuditLogsHandler(w, r)
			return
		}
	case "media":
		// GET /api/v1/media/{key...}
		if r.Method != http.MethodGet {
			writeAPIError(w, 405, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}
		key := strings.TrimPrefix(path, "media/")
		if key == "" || s.MediaStore == nil {
			writeAPIError(w, 404, "NOT_FOUND", "media not found")
			return
		}
		obj, err := s.Platform.GetMediaObject(r.Context(), tenant, key)
		if err != nil {
			writeAPIError(w, 404, "NOT_FOUND", "media not found")
			return
		}
		rc, err := s.MediaStore.Open(r.Context(), key)
		if err != nil {
			writeAPIError(w, 404, "NOT_FOUND", "media not found")
			return
		}
		defer rc.Close()
		if obj.MimeType != "" {
			w.Header().Set("Content-Type", obj.MimeType)
		}
		if obj.Size > 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(obj.Size, 10))
		}
		w.Header().Set("Cache-Control", "private, max-age=86400")
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, rc)
		return
	}
	writeAPIError(w, 404, "NOT_FOUND", "resource not found")
}

// contactActivityRequest is the JSON body for creating a contact-level follow-up.
type contactActivityRequest struct {
	Type       string `json:"type"`
	Summary    string `json:"summary"`
	NextAction string `json:"next_action"`
	Priority   string `json:"priority"`
	DueAt      string `json:"due_at"`
}

func (s *Server) createContactActivity(w http.ResponseWriter, r *http.Request, tenant, contactID uuid.UUID) {
	var req contactActivityRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}
	if req.Summary == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "summary is required")
		return
	}
	if req.Type == "" {
		req.Type = "FOLLOW_UP"
	}
	priority := req.Priority
	if priority == "" {
		priority = "NORMAL"
	}
	var due *time.Time
	if req.DueAt != "" {
		t, err := time.Parse(time.RFC3339, req.DueAt)
		if err != nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "due_at must be an RFC3339 timestamp")
			return
		}
		due = &t
	}
	activity, err := s.Platform.CreateContactActivity(tenant, contactID, domain.ContactActivityInput{
		Type:       req.Type,
		Summary:    req.Summary,
		NextAction: req.NextAction,
		Priority:   priority,
		DueAt:      due,
	})
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
		return
	}
	WriteJSON(w, 201, activity)
}

func (s *Server) sendAPI(w http.ResponseWriter, r *http.Request, host string) {
	var req SendRequest
	if DecodeJSONBody(r, &req) != nil || strings.TrimSpace(req.Recipient) == "" || strings.TrimSpace(req.Message) == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "recipient and message are required")
		return
	}
	if req.Type == "" {
		req.Type = domain.Text
	}
	isGroup := req.IsGroup || strings.HasSuffix(req.Recipient, "@g.us") || strings.Contains(req.Recipient, "-")

	// Resolve operator attribution
	tenantID, hasTenant := s.tenant(r)
	var actualOperatorName string
	var actualOperatorID *uuid.UUID
	var operatorID *uuid.UUID
	var operatorName string
	var onBehalfAuditNote string
	if opIDStr := s.operatorID(r); opIDStr != "" && opIDStr != "api" {
		if opUUID, err := uuid.Parse(opIDStr); err == nil {
			actualOperatorID = &opUUID
			operatorID = &opUUID
			if hasTenant && s.Auth != nil {
				if op, err := s.Auth.GetOperatorByID(tenantID, opUUID); err == nil {
					actualOperatorName = op.Name
					operatorName = op.Name
				}
			}
		}
	}
	if req.OnBehalfOperatorID != "" {
		targetID, err := uuid.Parse(req.OnBehalfOperatorID)
		if err != nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "valid on_behalf_operator_id is required")
			return
		}
		if !hasTenant || s.Auth == nil {
			writeAPIError(w, 403, "FORBIDDEN", "operator attribution unavailable")
			return
		}
		target, err := s.Auth.GetOperatorByID(tenantID, targetID)
		if err != nil || !target.IsActive {
			writeAPIError(w, 404, "OPERATOR_NOT_FOUND", "operator not found")
			return
		}
		operatorID = &targetID
		operatorName = target.Name
		if req.ConversationID != uuid.Nil && actualOperatorID != nil && actualOperatorID.String() != targetID.String() {
			auditor := actualOperatorName
			if auditor == "" {
				auditor = "Operator"
			}
			onBehalfAuditNote = "Sent on behalf: " + auditor + " replied as " + target.Name
		}
	}

	if err := s.Manager.SendMessageRequest(host, domain.MessageRequest{
		Recipient:      req.Recipient,
		Message:        req.Message,
		IsGroup:        isGroup,
		Type:           req.Type,
		MediaPath:      req.MediaPath,
		MediaKey:       req.MediaKey,
		ReactionTarget: req.ReactionTarget,
		Actor:          domain.ActorOperator,
		OperatorID:     operatorID,
		OperatorName:   operatorName,
	}); err != nil {
		writeAPIError(w, 404, "ACCOUNT_NOT_FOUND", err.Error())
		return
	}
	if onBehalfAuditNote != "" && actualOperatorID != nil {
		_, _ = s.Platform.AddInternalNote(tenantID, req.ConversationID, domain.ActorSystem, actualOperatorID.String(), onBehalfAuditNote)
	}
	WriteJSON(w, 202, map[string]any{"status": "queued", "account": host, "recipient": req.Recipient})
}

type OnboardRequest struct {
	Email string `json:"email"`
}

type SendRequest struct {
	HostNumber         string             `json:"host_number"`
	ConversationID     uuid.UUID          `json:"conversation_id,omitempty"`
	Recipient          string             `json:"recipient"`
	Message            string             `json:"message"`
	IsGroup            bool               `json:"is_group"`
	Type               domain.MessageType `json:"type"`
	MediaPath          string             `json:"media_path,omitempty"`
	MediaKey           string             `json:"media_key,omitempty"`
	ReactionTarget     string             `json:"reaction_target,omitempty"`
	OnBehalfOperatorID string             `json:"on_behalf_operator_id,omitempty"`
}

func RequireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return false
}

func DecodeJSONBody(r *http.Request, target any) error {
	return json.NewDecoder(r.Body).Decode(target)
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) OnboardHandler(w http.ResponseWriter, r *http.Request) {
	if !RequireMethod(w, r, http.MethodPost) {
		return
	}
	var req OnboardRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	device := s.Manager.Container.NewDevice()
	client := whatsmeow.NewClient(device, waLog.Stdout("Onboard", "WARN", true))
	qrChan, _ := client.GetQRChannel(context.Background())
	if err := client.Connect(); err != nil {
		http.Error(w, "Failed to connect", http.StatusInternalServerError)
		return
	}
	go func() {
		for evt := range qrChan {
			if evt.Event == "code" {
				log.Printf("Subsystem: New QR Code for %s", req.Email)
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, log.Writer())

				pngDataURL, err := whatsapp.EncodeQRDataURL(evt.Code)
				if err == nil {
					log.Printf("Base64 QR (%s): %s", req.Email, pngDataURL)
				}
			} else if evt.Event == "success" {
				log.Printf("Subsystem: Onboarding success for %s. Disconnecting QR client.", req.Email)
				var hostPhone string
				if client.Store != nil && client.Store.ID != nil {
					hostPhone = whatsapp.ResolveJIDPhone(*client.Store.ID)
				} else if device.ID != nil {
					hostPhone = whatsapp.ResolveJIDPhone(*device.ID)
				}
				if s.Platform != nil && s.Manager != nil && s.Manager.ResolveTenant != nil && hostPhone != "" {
					tenantID := s.Manager.ResolveTenant(hostPhone)
					if tenantID != uuid.Nil {
						_, _ = s.Platform.RegisterAccount(tenantID, hostPhone, req.Email, "whatsmeow")
					}
				}
				client.Disconnect()

				// Small delay to let WhatsApp state settle.
				time.Sleep(2 * time.Second)

				s.Manager.SpawnInstance(device)
			}
		}
	}()
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "Onboarding initiated.")
}

func (s *Server) SendHandler(w http.ResponseWriter, r *http.Request) {
	if !RequireMethod(w, r, http.MethodPost) {
		return
	}
	var req SendRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = domain.Text
	}

	err := s.Manager.SendMessageRequest(req.HostNumber, domain.MessageRequest{
		Recipient:      req.Recipient,
		Message:        req.Message,
		IsGroup:        req.IsGroup,
		Type:           req.Type,
		MediaPath:      req.MediaPath,
		MediaKey:       req.MediaKey,
		ReactionTarget: req.ReactionTarget,
		Actor:          domain.ActorOperator,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Message queued.")
}

func (s *Server) ListBotsHandler(w http.ResponseWriter, r *http.Request) {
	if !RequireMethod(w, r, http.MethodGet) {
		return
	}
	bots := s.Manager.ListInstances()
	WriteJSON(w, http.StatusOK, bots)
}

func (s *Server) GetBotHandler(w http.ResponseWriter, r *http.Request) {
	if !RequireMethod(w, r, http.MethodGet) {
		return
	}
	host := r.URL.Query().Get("host")
	if host == "" {
		http.Error(w, "Host number required", http.StatusBadRequest)
		return
	}
	bot, err := s.Manager.GetInstance(host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	WriteJSON(w, http.StatusOK, bot)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
