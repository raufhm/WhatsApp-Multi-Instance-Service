package whatsapp

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Pairing session status constants.
const (
	PairingStatusAwaitingScan = "awaiting_scan"
	PairingStatusConnected    = "connected"
	PairingStatusFailed       = "failed"
	PairingStatusCancelled    = "cancelled"
	PairingStatusExpired      = "expired"
)

// EncodeQRDataURL converts a QR code string into a base64-encoded PNG data URL.
func EncodeQRDataURL(code string) (string, error) {
	png, err := qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}

// PairingSnapshot represents the public, pollable state of a pairing session.
type PairingSnapshot struct {
	ID        string `json:"id,omitempty"`
	Status    string `json:"status"`
	QRDataURL string `json:"qr_data_url,omitempty"`
	HostPhone string `json:"host_phone,omitempty"`
	Error     string `json:"error,omitempty"`
}

// PairingSession holds internal state and references for an in-flight QR pairing process.
type PairingSession struct {
	ID          string
	TenantID    uuid.UUID
	DisplayName string
	Status      string
	QRDataURL   string
	HostPhone   string
	Error       string
	Client      *whatsmeow.Client
	CancelFunc  context.CancelFunc
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PairingManager manages lifecycle and state transitions for WhatsApp web pairing sessions.
type PairingManager struct {
	sessions    map[string]*PairingSession
	tenantIndex map[uuid.UUID]string // tenantID -> active sessionID
	manager     *WhatsAppManager
	store       domain.AccountRegistrar
	ttl         time.Duration
	mu          sync.RWMutex
	stopJanitor chan struct{}
}

// NewPairingManager constructs a new PairingManager and starts background session reaping.
func NewPairingManager(manager *WhatsAppManager, store domain.AccountRegistrar) *PairingManager {
	pm := &PairingManager{
		sessions:    make(map[string]*PairingSession),
		tenantIndex: make(map[uuid.UUID]string),
		manager:     manager,
		store:       store,
		ttl:         10 * time.Minute,
		stopJanitor: make(chan struct{}),
	}
	go pm.runJanitor(1 * time.Minute)
	return pm
}

// SetStore updates the account registrar store used upon pairing completion.
func (p *PairingManager) SetStore(store domain.AccountRegistrar) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.store = store
}

// Close stops background janitor processing.
func (p *PairingManager) Close() {
	select {
	case <-p.stopJanitor:
	default:
		close(p.stopJanitor)
	}
}

func (p *PairingManager) runJanitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-p.stopJanitor:
			return
		case now := <-ticker.C:
			p.cleanupExpired(now)
		}
	}
}

func (p *PairingManager) cleanupExpired(now time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id, s := range p.sessions {
		maxAge := p.ttl
		if s.Status == PairingStatusConnected || s.Status == PairingStatusFailed || s.Status == PairingStatusCancelled {
			maxAge = 5 * time.Minute
		}
		if now.Sub(s.UpdatedAt) > maxAge {
			if s.CancelFunc != nil {
				s.CancelFunc()
			}
			if s.Client != nil && s.Client.IsConnected() {
				s.Client.Disconnect()
			}
			delete(p.sessions, id)
			if p.tenantIndex[s.TenantID] == id {
				delete(p.tenantIndex, s.TenantID)
			}
		}
	}
}

// Start initiates a new pairing session for the given tenant.
// If an active pairing session already exists for this tenant, it is cancelled first.
func (p *PairingManager) Start(tenantID uuid.UUID, displayName string) (string, error) {
	if tenantID == uuid.Nil {
		return "", errors.New("tenant_id is required")
	}
	if p.manager == nil || p.manager.Container == nil {
		return "", errors.New("whatsapp container unavailable")
	}

	p.mu.Lock()
	if oldID, exists := p.tenantIndex[tenantID]; exists {
		if oldSession, ok := p.sessions[oldID]; ok {
			if oldSession.CancelFunc != nil {
				oldSession.CancelFunc()
			}
			if oldSession.Client != nil && oldSession.Client.IsConnected() {
				oldSession.Client.Disconnect()
			}
			oldSession.Status = PairingStatusCancelled
			oldSession.UpdatedAt = time.Now()
		}
		delete(p.tenantIndex, tenantID)
	}

	sessionID := uuid.New().String()
	ctx, cancel := context.WithCancel(context.Background())

	device := p.manager.Container.NewDevice()
	client := whatsmeow.NewClient(device, waLog.Stdout("Pairing", "WARN", true))

	qrChan, err := client.GetQRChannel(ctx)
	if err != nil {
		cancel()
		p.mu.Unlock()
		return "", fmt.Errorf("failed to get qr channel: %w", err)
	}

	if err := client.Connect(); err != nil {
		cancel()
		p.mu.Unlock()
		return "", fmt.Errorf("failed to connect client: %w", err)
	}

	session := &PairingSession{
		ID:          sessionID,
		TenantID:    tenantID,
		DisplayName: displayName,
		Status:      PairingStatusAwaitingScan,
		Client:      client,
		CancelFunc:  cancel,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	p.sessions[sessionID] = session
	p.tenantIndex[tenantID] = sessionID
	p.mu.Unlock()

	go p.handleQRChannel(ctx, sessionID, device, client, qrChan)

	return sessionID, nil
}

func (p *PairingManager) handleQRChannel(ctx context.Context, sessionID string, device *store.Device, client *whatsmeow.Client, qrChan <-chan whatsmeow.QRChannelItem) {
	for {
		select {
		case <-ctx.Done():
			p.mu.Lock()
			if s, ok := p.sessions[sessionID]; ok && s.Status == PairingStatusAwaitingScan {
				s.Status = PairingStatusCancelled
				s.UpdatedAt = time.Now()
			}
			p.mu.Unlock()
			return
		case evt, ok := <-qrChan:
			if !ok {
				return
			}
			p.mu.Lock()
			session, ok := p.sessions[sessionID]
			if !ok {
				p.mu.Unlock()
				return
			}

			switch evt.Event {
			case "code":
				dataURL, err := EncodeQRDataURL(evt.Code)
				if err == nil {
					session.QRDataURL = dataURL
					session.Status = PairingStatusAwaitingScan
					session.UpdatedAt = time.Now()
				}
				p.mu.Unlock()

			case "success":
				var hostPhone string
				if client.Store != nil && client.Store.ID != nil {
					hostPhone = ResolveJIDPhone(*client.Store.ID)
				} else if device.ID != nil {
					hostPhone = ResolveJIDPhone(*device.ID)
				}
				session.HostPhone = hostPhone
				session.Status = PairingStatusConnected
				session.QRDataURL = ""
				session.UpdatedAt = time.Now()
				delete(p.tenantIndex, session.TenantID)

				tenantID := session.TenantID
				dispName := session.DisplayName
				storeRef := p.store
				p.mu.Unlock()

				if storeRef != nil && hostPhone != "" {
					if _, err := storeRef.RegisterAccount(tenantID, hostPhone, dispName, "whatsmeow"); err != nil {
						log.Printf("[pairing] error registering account %s for tenant %s: %v", hostPhone, tenantID, err)
					}
				}

				client.Disconnect()

				go func() {
					time.Sleep(1 * time.Second)
					if p.manager != nil {
						if err := p.manager.SpawnInstance(device); err != nil {
							log.Printf("[pairing] error spawning instance for %s: %v", hostPhone, err)
						}
					}
				}()
				return

			case "timeout":
				session.QRDataURL = ""
				session.Status = PairingStatusExpired
				session.UpdatedAt = time.Now()
				p.mu.Unlock()

			case "error":
				session.Status = PairingStatusFailed
				if evt.Error != nil {
					session.Error = evt.Error.Error()
				} else {
					session.Error = "pairing failed"
				}
				session.UpdatedAt = time.Now()
				delete(p.tenantIndex, session.TenantID)
				p.mu.Unlock()
				client.Disconnect()
				return

			default:
				p.mu.Unlock()
			}
		}
	}
}

// Get returns the snapshot of a pairing session by its ID.
func (p *PairingManager) Get(id string) (PairingSnapshot, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	session, ok := p.sessions[id]
	if !ok {
		return PairingSnapshot{}, false
	}
	return PairingSnapshot{
		ID:        session.ID,
		Status:    session.Status,
		QRDataURL: session.QRDataURL,
		HostPhone: session.HostPhone,
		Error:     session.Error,
	}, true
}

// Cancel terminates an in-flight pairing session by ID.
func (p *PairingManager) Cancel(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	session, ok := p.sessions[id]
	if !ok {
		return errors.New("pairing session not found")
	}

	if session.CancelFunc != nil {
		session.CancelFunc()
	}
	if session.Client != nil && session.Client.IsConnected() {
		session.Client.Disconnect()
	}
	session.Status = PairingStatusCancelled
	session.QRDataURL = ""
	session.UpdatedAt = time.Now()
	if p.tenantIndex[session.TenantID] == id {
		delete(p.tenantIndex, session.TenantID)
	}
	return nil
}
