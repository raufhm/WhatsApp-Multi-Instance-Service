package whatsapp

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
)

type mockAccountRegistrar struct {
	mu       sync.Mutex
	accounts []domain.WhatsAppAccount
	err      error
}

func (m *mockAccountRegistrar) RegisterAccount(tenantID uuid.UUID, hostID, displayName, provider string) (domain.WhatsAppAccount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return domain.WhatsAppAccount{}, m.err
	}
	a := domain.WhatsAppAccount{
		ID:          uuid.New(),
		TenantID:    tenantID,
		HostID:      hostID,
		DisplayName: displayName,
		Provider:    provider,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.accounts = append(m.accounts, a)
	return a, nil
}

func TestEncodeQRDataURL(t *testing.T) {
	code := "2@testpairingcode123456"
	dataURL, err := EncodeQRDataURL(code)
	if err != nil {
		t.Fatalf("unexpected error encoding QR: %v", err)
	}
	if !strings.HasPrefix(dataURL, "data:image/png;base64,") {
		t.Fatalf("expected data URL prefix, got %q", dataURL)
	}
}

func TestPairingManager_GetAndCancel(t *testing.T) {
	pm := &PairingManager{
		sessions:    make(map[string]*PairingSession),
		tenantIndex: make(map[uuid.UUID]string),
		ttl:         10 * time.Minute,
		stopJanitor: make(chan struct{}),
	}
	defer pm.Close()

	tenantID := uuid.New()
	sessionID := "sess-123"
	cancelled := false

	session := &PairingSession{
		ID:          sessionID,
		TenantID:    tenantID,
		DisplayName: "Support Host",
		Status:      PairingStatusAwaitingScan,
		QRDataURL:   "data:image/png;base64,abc",
		CancelFunc: func() {
			cancelled = true
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	pm.sessions[sessionID] = session
	pm.tenantIndex[tenantID] = sessionID

	// Test Get
	snap, found := pm.Get(sessionID)
	if !found {
		t.Fatalf("expected session %s to be found", sessionID)
	}
	if snap.Status != PairingStatusAwaitingScan {
		t.Fatalf("expected status %s, got %s", PairingStatusAwaitingScan, snap.Status)
	}
	if snap.QRDataURL != "data:image/png;base64,abc" {
		t.Fatalf("expected qr data url, got %s", snap.QRDataURL)
	}

	// Test Cancel
	if err := pm.Cancel(sessionID); err != nil {
		t.Fatalf("unexpected cancel error: %v", err)
	}
	if !cancelled {
		t.Fatalf("expected CancelFunc to be called")
	}

	snapAfter, found := pm.Get(sessionID)
	if !found || snapAfter.Status != PairingStatusCancelled {
		t.Fatalf("expected status %s, got %s", PairingStatusCancelled, snapAfter.Status)
	}
	if snapAfter.QRDataURL != "" {
		t.Fatalf("expected qr_data_url to be cleared on cancel")
	}
	if _, exists := pm.tenantIndex[tenantID]; exists {
		t.Fatalf("expected tenantIndex to be cleared on cancel")
	}
}

func TestPairingManager_CleanupExpired(t *testing.T) {
	pm := &PairingManager{
		sessions:    make(map[string]*PairingSession),
		tenantIndex: make(map[uuid.UUID]string),
		ttl:         10 * time.Minute,
		stopJanitor: make(chan struct{}),
	}
	defer pm.Close()

	tenantID := uuid.New()
	activeID := "active-sess"
	expiredID := "expired-sess"
	oldConnectedID := "connected-sess"

	now := time.Now()

	pm.sessions[activeID] = &PairingSession{
		ID:        activeID,
		TenantID:  tenantID,
		Status:    PairingStatusAwaitingScan,
		UpdatedAt: now.Add(-2 * time.Minute),
	}
	pm.tenantIndex[tenantID] = activeID

	pm.sessions[expiredID] = &PairingSession{
		ID:        expiredID,
		TenantID:  uuid.New(),
		Status:    PairingStatusAwaitingScan,
		UpdatedAt: now.Add(-15 * time.Minute),
	}

	pm.sessions[oldConnectedID] = &PairingSession{
		ID:        oldConnectedID,
		TenantID:  uuid.New(),
		Status:    PairingStatusConnected,
		UpdatedAt: now.Add(-6 * time.Minute),
	}

	pm.cleanupExpired(now)

	if _, ok := pm.sessions[activeID]; !ok {
		t.Fatalf("expected active session to remain")
	}
	if _, ok := pm.sessions[expiredID]; ok {
		t.Fatalf("expected expired session to be reaped")
	}
	if _, ok := pm.sessions[oldConnectedID]; ok {
		t.Fatalf("expected old connected session to be reaped")
	}
}
