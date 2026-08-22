package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

// channelAccountRepoStub returns a single account with a pre-defined display
// name so we can prove disconnect/reconnect never overwrites it.
type channelAccountRepoStub struct {
	apiRepoStub
	account domain.WhatsAppAccount
}

func (f *channelAccountRepoStub) GetAccount(tenantID, id uuid.UUID) (domain.WhatsAppAccount, error) {
	return f.account, nil
}

// channelManagerStub implements accountManager and records lifecycle calls.
type channelManagerStub struct {
	disconnected map[string]bool
	reconnected  map[string]bool
	connected    map[string]bool
}

func newChannelManagerStub() *channelManagerStub {
	return &channelManagerStub{
		disconnected: make(map[string]bool),
		reconnected:  make(map[string]bool),
		connected:    make(map[string]bool),
	}
}

func (m *channelManagerStub) ListInstances() []domain.InstanceInfo { return nil }
func (m *channelManagerStub) GetInstance(host string) (domain.InstanceInfo, error) {
	return domain.InstanceInfo{HostPhone: host, IsConnected: m.connected[host]}, nil
}
func (m *channelManagerStub) Disconnect(host string) error {
	m.disconnected[host] = true
	m.connected[host] = false
	return nil
}
func (m *channelManagerStub) Reconnect(host string) error {
	m.reconnected[host] = true
	m.connected[host] = true
	return nil
}

func TestDisconnectPreservesDisplayName(t *testing.T) {
	accountID := uuid.New()
	account := domain.WhatsAppAccount{
		ID:          accountID,
		TenantID:    uuid.New(),
		HostID:      "15551234567",
		Provider:    "whatsmeow",
		DisplayName: "Acme Support Channel",
	}
	srv := &DashboardAPIHandler{
		Platform: &channelAccountRepoStub{account: account},
		Manager:  newChannelManagerStub(),
	}
	auth := newFakeAuthWithRole(account.TenantID, "op@test", "secret", "admin")
	op := auth.operators["op@test"].op
	session, err := auth.CreateSession(op.ID, account.TenantID, time.Hour)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	srv.Auth = auth

	req := httptest.NewRequest(http.MethodPost, "/dashboard/api/accounts/"+accountID.String()+"/disconnect", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session.ID.String()})
	w := httptest.NewRecorder()
	DashboardSessionMiddleware(auth, srv).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var result accountWithHealth
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.DisplayName != "Acme Support Channel" {
		t.Fatalf("display_name was overwritten to %q", result.DisplayName)
	}
}

func TestReconnectPreservesDisplayName(t *testing.T) {
	accountID := uuid.New()
	account := domain.WhatsAppAccount{
		ID:          accountID,
		TenantID:    uuid.New(),
		HostID:      "15551234567",
		Provider:    "whatsmeow",
		DisplayName: "Acme Sales Channel",
	}
	srv := &DashboardAPIHandler{
		Platform: &channelAccountRepoStub{account: account},
		Manager:  newChannelManagerStub(),
	}
	auth := newFakeAuthWithRole(account.TenantID, "op@test", "secret", "admin")
	op := auth.operators["op@test"].op
	session, err := auth.CreateSession(op.ID, account.TenantID, time.Hour)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	srv.Auth = auth

	req := httptest.NewRequest(http.MethodPost, "/dashboard/api/accounts/"+accountID.String()+"/reconnect", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session.ID.String()})
	w := httptest.NewRecorder()
	DashboardSessionMiddleware(auth, srv).ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var result accountWithHealth
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.DisplayName != "Acme Sales Channel" {
		t.Fatalf("display_name was overwritten to %q", result.DisplayName)
	}
}
