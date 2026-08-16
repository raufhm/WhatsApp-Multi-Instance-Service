package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdp/qrterminal"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/internal/storage"
	"github.com/raufhm/whatsapp-testing/whatsapp"
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
}

type apiError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type accountHostResolver interface {
	AccountHost(uuid.UUID, string) (string, error)
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
	if err != nil || limit < 1 || limit > 100 {
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
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeAPIError(w, 405, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	tenant, ok := s.tenant(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "valid API key required")
		return
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
		if len(parts) == 1 && r.Method == http.MethodGet {
			v, err := s.Platform.ListAccounts(tenant)
			if err != nil {
				writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
				return
			}
			WriteJSON(w, 200, v)
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
	case "contacts":
		if len(parts) == 1 && r.Method == http.MethodGet {
			v, err := s.Platform.ListContacts(tenant, limit, offset)
			if err != nil {
				writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
				return
			}
			WriteJSON(w, 200, toContactListDTO(v))
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
	case "conversations", "tickets":
		if r.Method == http.MethodGet {
			status := r.URL.Query().Get("status")
			v, err := s.Platform.ListConversations(tenant, status, limit, offset)
			if err != nil {
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
						msgs, e := s.Platform.GetConversationTimeline(tenant, c.ID, limit, offset)
						if e != nil {
							writeAPIError(w, 500, "INTERNAL_ERROR", "request failed")
							return
						}
						WriteJSON(w, 200, map[string]any{"conversation": c, "messages": msgs})
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
		if len(parts) == 2 && r.Method == http.MethodPost {
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
	if err := s.Manager.SendMessageRequest(host, domain.MessageRequest{
		Recipient:      req.Recipient,
		Message:        req.Message,
		IsGroup:        req.IsGroup,
		Type:           req.Type,
		MediaPath:      req.MediaPath,
		MediaKey:       req.MediaKey,
		ReactionTarget: req.ReactionTarget,
		Actor:          domain.ActorOperator,
	}); err != nil {
		writeAPIError(w, 404, "ACCOUNT_NOT_FOUND", err.Error())
		return
	}
	WriteJSON(w, 202, map[string]any{"status": "queued", "account": host, "recipient": req.Recipient})
}

type OnboardRequest struct {
	Email string `json:"email"`
}

type SendRequest struct {
	HostNumber     string             `json:"host_number"`
	Recipient      string             `json:"recipient"`
	Message        string             `json:"message"`
	IsGroup        bool               `json:"is_group"`
	Type           domain.MessageType `json:"type"`
	MediaPath      string             `json:"media_path,omitempty"`
	MediaKey       string             `json:"media_key,omitempty"`
	ReactionTarget string             `json:"reaction_target,omitempty"`
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
