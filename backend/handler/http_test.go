package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
)

// apiRepoStub satisfies domain.PlatformRepository for handler tests.
type apiRepoStub struct {
	tenant uuid.UUID
	calls  int
}

func (f *apiRepoStub) AuthenticateAPIKey(key string) (uuid.UUID, error) {
	if key != "good" {
		return uuid.Nil, errors.New("denied")
	}
	return f.tenant, nil
}
func (f *apiRepoStub) RegisterAccount(tenantID uuid.UUID, hostID, displayName, provider string) (domain.WhatsAppAccount, error) {
	return domain.WhatsAppAccount{
		ID:          uuid.New(),
		TenantID:    tenantID,
		HostID:      hostID,
		DisplayName: displayName,
		Provider:    provider,
	}, nil
}
func (f *apiRepoStub) ListAccounts(uuid.UUID) ([]domain.WhatsAppAccount, error) {
	f.calls++
	return []domain.WhatsAppAccount{}, nil
}
func (f *apiRepoStub) GetAccount(uuid.UUID, uuid.UUID) (domain.WhatsAppAccount, error) {
	return domain.WhatsAppAccount{}, nil
}
func (f *apiRepoStub) ListContacts(uuid.UUID, int, int, string) ([]domain.Contact, int, error) {
	return []domain.Contact{}, 0, nil
}
func (f *apiRepoStub) GetContact(uuid.UUID, uuid.UUID) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (f *apiRepoStub) ListConversations(uuid.UUID, string, int, int) ([]domain.Conversation, error) {
	return []domain.Conversation{}, nil
}
func (f *apiRepoStub) GetConversation(uuid.UUID, uuid.UUID) (domain.Conversation, error) {
	return domain.Conversation{}, nil
}
func (f *apiRepoStub) ListConversationSummaries(uuid.UUID, string, int, int) ([]domain.ConversationSummary, error) {
	return []domain.ConversationSummary{}, nil
}
func (f *apiRepoStub) GetConversationTimeline(uuid.UUID, uuid.UUID, int, int) ([]domain.ConversationMessage, error) {
	return nil, nil
}
func (f *apiRepoStub) ListActivities(uuid.UUID, string, int, int) ([]domain.Activity, error) {
	return nil, nil
}
func (f *apiRepoStub) AcknowledgeActivity(uuid.UUID, uuid.UUID, string, time.Time) (domain.Activity, error) {
	return domain.Activity{}, nil
}
func (f *apiRepoStub) ListContactActivities(uuid.UUID, uuid.UUID, int, int) ([]domain.Activity, error) {
	return nil, nil
}
func (f *apiRepoStub) CreateContactActivity(uuid.UUID, uuid.UUID, domain.ContactActivityInput) (domain.Activity, error) {
	return domain.Activity{}, nil
}
func (f *apiRepoStub) GetActiveBotRuleSet(uuid.UUID) (domain.BotRuleSet, error) {
	return domain.BotRuleSet{}, nil
}
func (f *apiRepoStub) SaveBotRuleSet(uuid.UUID, []domain.BotRule) (domain.BotRuleSet, error) {
	return domain.BotRuleSet{}, nil
}
func (f *apiRepoStub) ListBotRuleSets(uuid.UUID) ([]domain.BotRuleSet, error) {
	return nil, nil
}
func (f *apiRepoStub) ActivateBotRuleSetVersion(uuid.UUID, int) (domain.BotRuleSet, error) {
	return domain.BotRuleSet{}, nil
}
func (f *apiRepoStub) AssignConversation(uuid.UUID, uuid.UUID, string, string) (domain.Conversation, error) {
	return domain.Conversation{}, nil
}
func (f *apiRepoStub) HandoffConversation(uuid.UUID, uuid.UUID, string) (domain.Conversation, error) {
	return domain.Conversation{}, nil
}
func (f *apiRepoStub) CloseConversationWithReason(uuid.UUID, uuid.UUID, string, string) (domain.Conversation, error) {
	return domain.Conversation{}, nil
}
func (f *apiRepoStub) ReopenConversation(uuid.UUID, uuid.UUID, string) (domain.Conversation, error) {
	return domain.Conversation{}, nil
}
func (f *apiRepoStub) AddInternalNote(uuid.UUID, uuid.UUID, domain.Actor, string, string) (domain.ConversationMessage, error) {
	return domain.ConversationMessage{}, nil
}
func (f *apiRepoStub) MergeConversations(uuid.UUID, uuid.UUID, uuid.UUID, string) (domain.Conversation, error) {
	return domain.Conversation{}, nil
}
func (f *apiRepoStub) SplitConversation(uuid.UUID, uuid.UUID, []uuid.UUID, string) (domain.Conversation, error) {
	return domain.Conversation{}, nil
}
func (f *apiRepoStub) ListOperatorAuditLogs(uuid.UUID, int, int) ([]domain.OperatorAuditLog, error) {
	return nil, nil
}
func (f *apiRepoStub) RecordMediaObject(context.Context, uuid.UUID, string, string, int64) error {
	return nil
}
func (f *apiRepoStub) GetMediaObject(context.Context, uuid.UUID, string) (domain.MediaObject, error) {
	return domain.MediaObject{}, nil
}
func (f *apiRepoStub) UpdateContact(uuid.UUID, uuid.UUID, domain.ContactUpdateInput) (domain.Contact, error) {
	return domain.Contact{}, nil
}
func (f *apiRepoStub) ListContactConversations(uuid.UUID, uuid.UUID, int, int) ([]domain.Conversation, error) {
	return nil, nil
}
func (f *apiRepoStub) ListContactFieldDefinitions(uuid.UUID) ([]domain.ContactFieldDefinition, error) {
	return nil, nil
}
func (f *apiRepoStub) GetContactFieldDefinition(uuid.UUID, uuid.UUID) (domain.ContactFieldDefinition, error) {
	return domain.ContactFieldDefinition{}, nil
}
func (f *apiRepoStub) CreateContactFieldDefinition(uuid.UUID, string, string, string, []string, bool, int) (domain.ContactFieldDefinition, error) {
	return domain.ContactFieldDefinition{}, nil
}
func (f *apiRepoStub) UpdateContactFieldDefinition(uuid.UUID, uuid.UUID, string, string, []string, bool, int, bool) (domain.ContactFieldDefinition, error) {
	return domain.ContactFieldDefinition{}, nil
}
func (f *apiRepoStub) DeleteContactFieldDefinition(uuid.UUID, uuid.UUID) error {
	return nil
}
func (f *apiRepoStub) ListPipelines(uuid.UUID, *bool) ([]domain.Pipeline, error) {
	return nil, nil
}
func (f *apiRepoStub) GetPipeline(uuid.UUID, uuid.UUID) (domain.Pipeline, error) {
	return domain.Pipeline{}, nil
}
func (f *apiRepoStub) GetDefaultPipeline(uuid.UUID) (domain.Pipeline, error) {
	return domain.Pipeline{}, nil
}
func (f *apiRepoStub) CreatePipeline(uuid.UUID, string, string, bool) (domain.Pipeline, error) {
	return domain.Pipeline{}, nil
}
func (f *apiRepoStub) UpdatePipeline(uuid.UUID, uuid.UUID, *string, *string, *bool, *bool) (domain.Pipeline, error) {
	return domain.Pipeline{}, nil
}
func (f *apiRepoStub) DeletePipeline(uuid.UUID, uuid.UUID) error {
	return nil
}
func (f *apiRepoStub) ListDealStages(uuid.UUID, *uuid.UUID, *bool) ([]domain.DealStage, error) {
	return nil, nil
}
func (f *apiRepoStub) GetDealStage(uuid.UUID, uuid.UUID) (domain.DealStage, error) {
	return domain.DealStage{}, nil
}
func (f *apiRepoStub) CreateDealStage(uuid.UUID, *uuid.UUID, string, string, string, string, int, bool, bool) (domain.DealStage, error) {
	return domain.DealStage{}, nil
}
func (f *apiRepoStub) UpdateDealStage(uuid.UUID, uuid.UUID, *uuid.UUID, *string, *string, *string, *int, *bool, *bool, *bool) (domain.DealStage, error) {
	return domain.DealStage{}, nil
}
func (f *apiRepoStub) DeleteDealStage(uuid.UUID, uuid.UUID) error {
	return nil
}
func (f *apiRepoStub) MoveContactToStage(uuid.UUID, uuid.UUID, string, *uuid.UUID, string, uuid.UUID) (domain.DealStageTransition, error) {
	return domain.DealStageTransition{}, nil
}
func (f *apiRepoStub) ListDealStageHistory(uuid.UUID, uuid.UUID, int, int) ([]domain.DealStageTransition, error) {
	return nil, nil
}

func (f *apiRepoStub) GetSubscription(uuid.UUID) (domain.Subscription, error) {
	return domain.Subscription{}, nil
}

func (f *apiRepoStub) GetQuota(uuid.UUID) (domain.Quota, error) {
	return domain.Quota{}, nil
}

func (f *apiRepoStub) IncrementQuota(uuid.UUID) error {
	return nil
}

// acknowledgeStub captures acknowledge call arguments.
type acknowledgeStub struct {
	apiRepoStub
	ackCalls int
	ackActor string
}

func (s *acknowledgeStub) AcknowledgeActivity(_ uuid.UUID, _ uuid.UUID, actor string, _ time.Time) (domain.Activity, error) {
	s.ackCalls++
	s.ackActor = actor
	return domain.Activity{Status: domain.ActivityAcknowledged}, nil
}

// --- Authentication ---

func TestAPIRequiresTenantAuthentication(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/accounts", nil)
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401 without key, got %d", w.Code)
	}
	if f.calls != 0 {
		t.Fatal("unauthenticated request reached repository")
	}
}

func TestAPIBearerTokenAuth(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/accounts", nil)
	r.Header.Set("Authorization", "Bearer good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("Bearer auth failed: status=%d", w.Code)
	}
}

// --- Pagination ---

func TestAPIPaginationValidation(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}

	// valid pagination
	r := httptest.NewRequest("GET", "/api/v1/contacts?limit=10&offset=20", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("valid pagination: status=%d body=%s", w.Code, w.Body.String())
	}

	// limit > 100
	r = httptest.NewRequest("GET", "/api/v1/contacts?limit=101", nil)
	r.Header.Set("X-API-Key", "good")
	w = httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 400 {
		t.Fatalf("limit>100 expected 400, got %d", w.Code)
	}

	// negative offset
	r = httptest.NewRequest("GET", "/api/v1/contacts?offset=-1", nil)
	r.Header.Set("X-API-Key", "good")
	w = httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 400 {
		t.Fatalf("negative offset expected 400, got %d", w.Code)
	}
}

// --- Accounts ---

func TestAPIListAccounts(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/accounts", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
}

// --- Contacts ---

func TestAPIListContacts(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/contacts", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestAPIGetContactByID(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/contacts/"+uuid.New().String(), nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestAPIGetContactBadID(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/contacts/not-a-uuid", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400 for bad UUID, got %d", w.Code)
	}
}

// contactRepoStub returns real contacts/activities so JSON serialization is
// exercised end-to-end through the API handler.
type contactRepoStub struct {
	apiRepoStub
	contact  domain.Contact
	activity domain.Activity
}

func (f *contactRepoStub) ListContacts(uuid.UUID, int, int, string) ([]domain.Contact, int, error) {
	return []domain.Contact{f.contact}, 1, nil
}
func (f *contactRepoStub) GetContact(uuid.UUID, uuid.UUID) (domain.Contact, error) {
	return f.contact, nil
}
func (f *contactRepoStub) ListContactActivities(uuid.UUID, uuid.UUID, int, int) ([]domain.Activity, error) {
	return []domain.Activity{f.activity}, nil
}
func (f *contactRepoStub) CreateContactActivity(uuid.UUID, uuid.UUID, domain.ContactActivityInput) (domain.Activity, error) {
	a := f.activity
	a.Type = "FOLLOW_UP"
	a.Summary = "created"
	return a, nil
}

func contactStub() *contactRepoStub {
	return &contactRepoStub{
		apiRepoStub: apiRepoStub{tenant: uuid.New()},
		contact: domain.Contact{
			ID:                uuid.New(),
			TenantID:          uuid.New(),
			NormalizedAddress: "15551234567@s.whatsapp.net",
			ProviderAddress:   "15551234567",
			DisplayName:       "Acme Customer",
			Metadata: map[string]any{
				"email": "customer@acme.test",
				"tags":  []any{"vip", "support"},
			},
		},
		activity: domain.Activity{
			ID:         uuid.New(),
			TenantID:   uuid.New(),
			Type:       "FOLLOW_UP",
			Summary:    "Call customer back",
			NextAction: "phone",
			Priority:   "HIGH",
			Status:     domain.ActivityPending,
		},
	}
}

func TestContactListSerializesDisplayFields(t *testing.T) {
	f := contactStub()
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/contacts", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var payload struct {
		Items []struct {
			Name   string   `json:"name"`
			Number string   `json:"number"`
			Email  string   `json:"email"`
			Tags   []string `json:"tags"`
			ID     string   `json:"id"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(payload.Items))
	}
	c := payload.Items[0]
	if c.Name != "Acme Customer" {
		t.Errorf("name=%q", c.Name)
	}
	if c.Number != "15551234567" {
		t.Errorf("number=%q", c.Number)
	}
	if c.Email != "customer@acme.test" {
		t.Errorf("email=%q", c.Email)
	}
	if len(c.Tags) != 2 || c.Tags[0] != "vip" {
		t.Errorf("tags=%v", c.Tags)
	}
	if c.ID == "" {
		t.Error("expected contact id in payload")
	}
}

func TestContactGetSerializesDisplayFields(t *testing.T) {
	f := contactStub()
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/contacts/"+f.contact.ID.String(), nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if got := strings.Contains(w.Body.String(), `"name":"Acme Customer"`); !got {
		t.Fatalf("name field missing: %s", w.Body.String())
	}
}

func TestAPIListContactActivities(t *testing.T) {
	f := contactStub()
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/contacts/"+f.contact.ID.String()+"/activities", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var payload []domain.Activity
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload) != 1 || payload[0].Summary != "Call customer back" {
		t.Fatalf("unexpected activities payload: %+v", payload)
	}
}

func TestAPICreateContactActivity(t *testing.T) {
	f := contactStub()
	s := &Server{Platform: f}
	body := `{"type":"FOLLOW_UP","summary":"Follow up on quote","next_action":"send pricing","priority":"HIGH"}`
	r := httptest.NewRequest("POST", "/api/v1/contacts/"+f.contact.ID.String()+"/activities", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestAPICreateContactActivityRequiresSummary(t *testing.T) {
	f := contactStub()
	s := &Server{Platform: f}
	body := `{"summary":""}`
	r := httptest.NewRequest("POST", "/api/v1/contacts/"+f.contact.ID.String()+"/activities", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400 for empty summary, got %d", w.Code)
	}
}

func TestAPIContactActivitiesBadID(t *testing.T) {
	f := contactStub()
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/contacts/not-a-uuid/activities", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400 for bad UUID, got %d", w.Code)
	}
}

// --- Conversations / Tickets ---

func TestAPIListConversationsAndTicketAlias(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	for _, path := range []string{
		"/api/v1/conversations",
		"/api/v1/conversations?status=OPEN",
		"/api/v1/tickets",
	} {
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("X-API-Key", "good")
		w := httptest.NewRecorder()
		s.APIHandler(w, r)
		if w.Code != 200 {
			t.Fatalf("path=%s status=%d body=%s", path, w.Code, w.Body.String())
		}
	}
}

func TestAPIConversationNotFound(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/conversations/9999999", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 404 {
		t.Fatalf("expected 404 for unknown ticket, got %d", w.Code)
	}
}

// summaryStub returns an enriched conversation summary to prove the list
// payload carries contact identity and last-message preview.
type summaryStub struct {
	apiRepoStub
	items []domain.ConversationSummary
}

func (f *summaryStub) ListConversationSummaries(uuid.UUID, string, int, int) ([]domain.ConversationSummary, error) {
	return f.items, nil
}

func TestAPIListConversationSummariesEnriched(t *testing.T) {
	contactID := uuid.New()
	summary := domain.ConversationSummary{
		Conversation: domain.Conversation{
			ID:             uuid.New(),
			TenantID:       uuid.New(),
			ContactID:      contactID,
			TicketNumber:   42,
			Status:         domain.ConversationOpen,
			Assignee:       "Ada",
			LastActivityAt: time.Now(),
		},
		ContactName:   "Ada Lovelace",
		ContactNumber: "15551234567",
		LastMessage:   "What is the status of my order?",
		LastActor:     domain.ActorContact,
	}
	f := &summaryStub{apiRepoStub: apiRepoStub{tenant: uuid.New()}, items: []domain.ConversationSummary{summary}}
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/conversations?status=OPEN", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(body))
	}
	got := body[0]
	for k, want := range map[string]any{
		"ticket_number":        float64(42),
		"status":               "OPEN",
		"assignee":             "Ada",
		"contact_name":         "Ada Lovelace",
		"contact_number":       "15551234567",
		"last_message_preview": "What is the status of my order?",
		"last_message_actor":   "CONTACT",
	} {
		if got[k] != want {
			t.Errorf("field %q = %v, want %v", k, got[k], want)
		}
	}
}

// --- Activities ---

func TestAPIListActivities(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	for _, path := range []string{"/api/v1/activities", "/api/v1/activities?status=PENDING"} {
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("X-API-Key", "good")
		w := httptest.NewRecorder()
		s.APIHandler(w, r)
		if w.Code != 200 {
			t.Fatalf("path=%s status=%d body=%s", path, w.Code, w.Body.String())
		}
	}
}

// --- Acknowledge ---

func TestAPIAcknowledgeActivityIdempotent(t *testing.T) {
	stub := &acknowledgeStub{apiRepoStub: apiRepoStub{tenant: uuid.New()}}
	s := &Server{Platform: stub}
	activityID := uuid.New()

	// Acknowledge twice — both calls must succeed (idempotency enforced by SQL in real DB).
	for i := 0; i < 2; i++ {
		r := httptest.NewRequest("POST", "/api/v1/activities/"+activityID.String()+"/acknowledge", nil)
		r.Header.Set("X-API-Key", "good")
		r.Header.Set("X-Actor", "operator-1")
		w := httptest.NewRecorder()
		s.APIHandler(w, r)
		if w.Code != 200 {
			t.Fatalf("ack attempt %d: status=%d body=%s", i+1, w.Code, w.Body.String())
		}
	}
	if stub.ackCalls != 2 {
		t.Fatalf("expected 2 acknowledge calls, got %d", stub.ackCalls)
	}
	if stub.ackActor != "operator-1" {
		t.Fatalf("actor not propagated: got %q", stub.ackActor)
	}
}

func TestAPIAcknowledgeDefaultsActorToAPI(t *testing.T) {
	stub := &acknowledgeStub{apiRepoStub: apiRepoStub{tenant: uuid.New()}}
	s := &Server{Platform: stub}
	r := httptest.NewRequest("POST", "/api/v1/activities/"+uuid.New().String()+"/acknowledge", nil)
	r.Header.Set("X-API-Key", "good")
	// No X-Actor header — should default to "api".
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
	if stub.ackActor != "api" {
		t.Fatalf("expected default actor 'api', got %q", stub.ackActor)
	}
}

func TestAPIAcknowledgeBadID(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	r := httptest.NewRequest("POST", "/api/v1/activities/not-a-uuid/acknowledge", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400 for bad UUID, got %d", w.Code)
	}
}

// --- Generic ---

func TestAPIMethodNotAllowed(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	r := httptest.NewRequest("DELETE", "/api/v1/accounts", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 405 {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestAPIUnknownResourceReturns404(t *testing.T) {
	f := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: f}
	r := httptest.NewRequest("GET", "/api/v1/nonexistent", nil)
	r.Header.Set("X-API-Key", "good")
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// Compile-time check: apiRepoStub satisfies domain.PlatformRepository.
var _ domain.PlatformRepository = (*apiRepoStub)(nil)

type sessionAuthStub struct {
	sessions map[uuid.UUID]domain.Session
}

func (s *sessionAuthStub) GetSessionByID(id uuid.UUID) (domain.Session, error) {
	if sess, ok := s.sessions[id]; ok {
		return sess, nil
	}
	return domain.Session{}, errors.New("session not found")
}

func (s *sessionAuthStub) GetOperatorByID(tenantID, operatorID uuid.UUID) (domain.Operator, error) {
	return domain.Operator{ID: operatorID, TenantID: tenantID, Name: "Test Operator"}, nil
}

func TestAPISessionAuthentication(t *testing.T) {
	tenantID := uuid.New()
	operatorID := uuid.New()
	sessionID := uuid.New()

	auth := &sessionAuthStub{
		sessions: map[uuid.UUID]domain.Session{
			sessionID: {
				ID:         sessionID,
				TenantID:   tenantID,
				OperatorID: operatorID,
				ExpiresAt:  time.Now().Add(8 * time.Hour),
			},
		},
	}
	repo := &apiRepoStub{tenant: tenantID}
	s := &Server{Platform: repo, Auth: auth}

	for _, path := range []string{
		"/api/v1/accounts",
		"/api/v1/contacts",
		"/api/v1/conversations",
		"/api/v1/activities",
	} {
		r := httptest.NewRequest("GET", path, nil)
		r.AddCookie(&http.Cookie{
			Name:  sessionCookieName,
			Value: sessionID.String(),
		})
		w := httptest.NewRecorder()
		s.APIHandler(w, r)
		if w.Code != 200 {
			t.Fatalf("path=%s status=%d body=%s", path, w.Code, w.Body.String())
		}
	}
}

func TestAPISessionAuthenticationInvalid(t *testing.T) {
	auth := &sessionAuthStub{sessions: map[uuid.UUID]domain.Session{}}
	repo := &apiRepoStub{tenant: uuid.New()}
	s := &Server{Platform: repo, Auth: auth}

	r := httptest.NewRequest("GET", "/api/v1/conversations", nil)
	r.AddCookie(&http.Cookie{
		Name:  sessionCookieName,
		Value: uuid.New().String(),
	})
	w := httptest.NewRecorder()
	s.APIHandler(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401 for unknown session, got %d", w.Code)
	}
}

type operatorActionStub struct {
	apiRepoStub
	lastAssignOperator  string
	lastHandoffOperator string
	lastCloseOperator   string
	lastReopenOperator  string
	lastNoteOperator    string
	lastMergeOperator   string
	lastSplitOperator   string
	lastAckOperator     string
}

func (s *operatorActionStub) AssignConversation(_ uuid.UUID, _ uuid.UUID, _ string, op string) (domain.Conversation, error) {
	s.lastAssignOperator = op
	return domain.Conversation{}, nil
}

func (s *operatorActionStub) HandoffConversation(_ uuid.UUID, _ uuid.UUID, op string) (domain.Conversation, error) {
	s.lastHandoffOperator = op
	return domain.Conversation{}, nil
}

func (s *operatorActionStub) CloseConversationWithReason(_ uuid.UUID, _ uuid.UUID, _ string, op string) (domain.Conversation, error) {
	s.lastCloseOperator = op
	return domain.Conversation{}, nil
}

func (s *operatorActionStub) ReopenConversation(_ uuid.UUID, _ uuid.UUID, op string) (domain.Conversation, error) {
	s.lastReopenOperator = op
	return domain.Conversation{}, nil
}

func (s *operatorActionStub) AddInternalNote(_ uuid.UUID, _ uuid.UUID, _ domain.Actor, op string, _ string) (domain.ConversationMessage, error) {
	s.lastNoteOperator = op
	return domain.ConversationMessage{}, nil
}

func (s *operatorActionStub) MergeConversations(_ uuid.UUID, _ uuid.UUID, _ uuid.UUID, op string) (domain.Conversation, error) {
	s.lastMergeOperator = op
	return domain.Conversation{}, nil
}

func (s *operatorActionStub) SplitConversation(_ uuid.UUID, _ uuid.UUID, _ []uuid.UUID, op string) (domain.Conversation, error) {
	s.lastSplitOperator = op
	return domain.Conversation{}, nil
}

func (s *operatorActionStub) AcknowledgeActivity(_ uuid.UUID, _ uuid.UUID, op string, _ time.Time) (domain.Activity, error) {
	s.lastAckOperator = op
	return domain.Activity{}, nil
}

func TestAPIOperatorActionsSessionActor(t *testing.T) {
	tenantID := uuid.New()
	operatorID := uuid.New()
	sessionID := uuid.New()
	convID := uuid.New()
	targetConvID := uuid.New()
	actID := uuid.New()
	msgID := uuid.New()

	auth := &sessionAuthStub{
		sessions: map[uuid.UUID]domain.Session{
			sessionID: {
				ID:         sessionID,
				TenantID:   tenantID,
				OperatorID: operatorID,
				ExpiresAt:  time.Now().Add(8 * time.Hour),
			},
		},
	}
	repo := &operatorActionStub{apiRepoStub: apiRepoStub{tenant: tenantID}}
	s := &Server{Platform: repo, Auth: auth}

	reqs := []struct {
		name     string
		path     string
		body     string
		expected *string
	}{
		{
			name:     "assign",
			path:     "/api/v1/operator/assign?id=" + convID.String(),
			body:     `{"assignee":"agent-1"}`,
			expected: &repo.lastAssignOperator,
		},
		{
			name:     "handoff",
			path:     "/api/v1/operator/handoff?id=" + convID.String(),
			body:     `{"reason":"escalation"}`,
			expected: &repo.lastHandoffOperator,
		},
		{
			name:     "close",
			path:     "/api/v1/operator/close?id=" + convID.String(),
			body:     `{"reason":"resolved"}`,
			expected: &repo.lastCloseOperator,
		},
		{
			name:     "reopen",
			path:     "/api/v1/operator/reopen?id=" + convID.String(),
			body:     `{}`,
			expected: &repo.lastReopenOperator,
		},
		{
			name:     "notes",
			path:     "/api/v1/notes?id=" + convID.String(),
			body:     `{"content":"called customer"}`,
			expected: &repo.lastNoteOperator,
		},
		{
			name:     "merge",
			path:     "/api/v1/merge",
			body:     `{"source_id":"` + convID.String() + `","target_id":"` + targetConvID.String() + `"}`,
			expected: &repo.lastMergeOperator,
		},
		{
			name:     "split",
			path:     "/api/v1/split?id=" + convID.String(),
			body:     `{"message_ids":["` + msgID.String() + `"]}`,
			expected: &repo.lastSplitOperator,
		},
		{
			name:     "acknowledge",
			path:     "/api/v1/activities/" + actID.String() + "/acknowledge",
			body:     `{}`,
			expected: &repo.lastAckOperator,
		},
	}

	for _, tc := range reqs {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", tc.path, strings.NewReader(tc.body))
			r.Header.Set("Content-Type", "application/json")
			r.AddCookie(&http.Cookie{
				Name:  sessionCookieName,
				Value: sessionID.String(),
			})
			w := httptest.NewRecorder()
			s.APIHandler(w, r)
			if w.Code >= 400 {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if *tc.expected != operatorID.String() {
				t.Fatalf("expected operator %s, got %s", operatorID.String(), *tc.expected)
			}
		})
	}
}
