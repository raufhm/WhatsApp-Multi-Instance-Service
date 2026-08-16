package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
	"github.com/raufhm/whatsapp-testing/internal/storage"
	"github.com/raufhm/whatsapp-testing/whatsapp"
	"golang.org/x/crypto/bcrypt"
)

// fakeAuth implements storage.OperatorAuth for dashboard handler tests.
type fakeAuth struct {
	operators map[string]struct {
		op   domain.Operator
		hash string
	}
	sessions map[string]struct {
		s domain.Session
	}
	nextSession uuid.UUID
}

func newFakeAuth(tenantID uuid.UUID, email, password string) *fakeAuth {
	return newFakeAuthWithRole(tenantID, email, password, "operator")
}

func newFakeAuthWithRole(tenantID uuid.UUID, email, password, role string) *fakeAuth {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &fakeAuth{
		operators: map[string]struct {
			op   domain.Operator
			hash string
		}{
			email: {
				op: domain.Operator{
					ID: uuid.New(), TenantID: tenantID, Email: email, Name: "Test Op",
					Role: role, IsActive: true,
				},
				hash: string(hash),
			},
		},
		sessions:    map[string]struct{ s domain.Session }{},
		nextSession: uuid.New(),
	}
}

func (f *fakeAuth) FindOperatorByEmail(tenantID uuid.UUID, email string) (domain.Operator, string, error) {
	rec, ok := f.operators[email]
	if !ok || rec.op.TenantID != tenantID {
		return domain.Operator{}, "", storage.ErrOperatorNotFound
	}
	return rec.op, rec.hash, nil
}
func (f *fakeAuth) GetOperatorByID(tenantID, operatorID uuid.UUID) (domain.Operator, error) {
	for _, rec := range f.operators {
		if rec.op.ID == operatorID && rec.op.TenantID == tenantID {
			return rec.op, nil
		}
	}
	return domain.Operator{}, storage.ErrOperatorNotFound
}
func (f *fakeAuth) CreateOperator(uuid.UUID, string, string, string, string) (domain.Operator, error) {
	return domain.Operator{}, nil
}
func (f *fakeAuth) CreateSession(operatorID, tenantID uuid.UUID, ttl time.Duration) (domain.Session, error) {
	s := domain.Session{ID: f.nextSession, OperatorID: operatorID, TenantID: tenantID, ExpiresAt: time.Now().Add(ttl)}
	f.sessions[s.ID.String()] = struct{ s domain.Session }{s}
	return s, nil
}
func (f *fakeAuth) GetSessionByID(sessionID uuid.UUID) (domain.Session, error) {
	rec, ok := f.sessions[sessionID.String()]
	if !ok {
		return domain.Session{}, storage.ErrOperatorNotFound
	}
	if rec.s.ExpiresAt.Before(time.Now()) {
		return domain.Session{}, storage.ErrOperatorNotFound
	}
	return rec.s, nil
}
func (f *fakeAuth) DeleteSession(sessionID uuid.UUID) error {
	delete(f.sessions, sessionID.String())
	return nil
}
func (f *fakeAuth) TouchSession(sessionID uuid.UUID) error { return nil }

func (f *fakeAuth) FindOperatorByIdentifier(tenantID uuid.UUID, identifier string) (domain.Operator, string, string, error) {
	for _, rec := range f.operators {
		if rec.op.TenantID == tenantID && (rec.op.Email == identifier || rec.op.WhatsappNumber == identifier) {
			return rec.op, "", rec.hash, nil
		}
	}
	return domain.Operator{}, "", "", storage.ErrOperatorNotFound
}

func (f *fakeAuth) GetOperatorByIDWithSecret(tenantID, operatorID uuid.UUID) (domain.Operator, string, error) {
	for _, rec := range f.operators {
		if rec.op.ID == operatorID && rec.op.TenantID == tenantID {
			return rec.op, "", nil
		}
	}
	return domain.Operator{}, "", storage.ErrOperatorNotFound
}

func (f *fakeAuth) ListOperators(tenantID uuid.UUID) ([]domain.Operator, error) {
	var list []domain.Operator
	for _, rec := range f.operators {
		if rec.op.TenantID == tenantID {
			list = append(list, rec.op)
		}
	}
	return list, nil
}

func (f *fakeAuth) SignupTenant(tenantName, adminName, adminEmail, adminWhatsapp string) (domain.Tenant, domain.Operator, string, error) {
	tenantID := uuid.New()
	t := domain.Tenant{ID: tenantID, Name: tenantName, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	op := domain.Operator{ID: uuid.New(), TenantID: tenantID, Name: adminName, Email: adminEmail, WhatsappNumber: adminWhatsapp, Role: "admin", IsActive: true}
	return t, op, "test-verify-token", nil
}

func (f *fakeAuth) VerifyEmailToken(rawToken string) (domain.Tenant, domain.Operator, string, error) {
	tenantID := uuid.New()
	t := domain.Tenant{ID: tenantID, Name: "Test Org"}
	op := domain.Operator{ID: uuid.New(), TenantID: tenantID, Name: "Admin", Email: "admin@test.com", Role: "admin", IsActive: true}
	return t, op, "test-setup-token", nil
}

func (f *fakeAuth) GetTOTPSetupInfo(rawSetupToken string) (domain.Operator, domain.Tenant, string, error) {
	tenantID := uuid.New()
	t := domain.Tenant{ID: tenantID, Name: "Test Org"}
	op := domain.Operator{ID: uuid.New(), TenantID: tenantID, Name: "Admin", Email: "admin@test.com", Role: "admin", IsActive: true}
	return op, t, "JBSWY3DPEHPK3PXP", nil
}

func (f *fakeAuth) VerifyTOTPSetup(rawSetupToken, code string) (domain.Operator, []string, domain.Session, error) {
	tenantID := uuid.New()
	op := domain.Operator{ID: uuid.New(), TenantID: tenantID, Name: "Admin", Email: "admin@test.com", Role: "admin", IsActive: true}
	session := domain.Session{ID: uuid.New(), OperatorID: op.ID, TenantID: tenantID, ExpiresAt: time.Now().Add(8 * time.Hour)}
	return op, []string{"1111-2222", "3333-4444"}, session, nil
}

func (f *fakeAuth) CreateInvitation(tenantID uuid.UUID, createdBy *uuid.UUID, recipient, channel, role, whatsappNumber, email string) (domain.Invitation, string, error) {
	inv := domain.Invitation{ID: uuid.New(), TenantID: tenantID, Recipient: recipient, Channel: channel, Role: role, Status: "pending", ExpiresAt: time.Now().Add(7 * 24 * time.Hour)}
	return inv, "invite-token-123", nil
}

func (f *fakeAuth) GetInvitationByToken(rawToken string) (domain.Invitation, domain.Tenant, error) {
	tenantID := uuid.New()
	inv := domain.Invitation{ID: uuid.New(), TenantID: tenantID, Recipient: "operator@test.com", Channel: "whatsapp", Role: "operator", Status: "pending"}
	t := domain.Tenant{ID: tenantID, Name: "Test Tenant"}
	return inv, t, nil
}

func (f *fakeAuth) AcceptInvitationAndSignupOperator(rawToken, name, whatsappNumber, email string) (domain.Operator, string, error) {
	tenantID := uuid.New()
	op := domain.Operator{ID: uuid.New(), TenantID: tenantID, Name: name, Email: email, WhatsappNumber: whatsappNumber, Role: "operator", IsActive: true}
	return op, "setup-token-456", nil
}

func (f *fakeAuth) ListInvitations(tenantID uuid.UUID) ([]domain.Invitation, error) {
	return []domain.Invitation{}, nil
}

func (f *fakeAuth) RevokeInvitation(tenantID, invitationID uuid.UUID) error {
	return nil
}

func (f *fakeAuth) VerifyBackupCodeAndLogin(tenantID uuid.UUID, identifier, code string) (domain.Operator, domain.Session, int, error) {
	op := domain.Operator{ID: uuid.New(), TenantID: tenantID, Email: identifier, Role: "operator", IsActive: true}
	session := domain.Session{ID: uuid.New(), OperatorID: op.ID, TenantID: tenantID, ExpiresAt: time.Now().Add(8 * time.Hour)}
	return op, session, 9, nil
}

func (f *fakeAuth) RegenerateBackupCodes(operatorID uuid.UUID) ([]string, error) {
	return []string{"AAAA-BBBB", "CCCC-DDDD"}, nil
}

func (f *fakeAuth) CountRemainingBackupCodes(operatorID uuid.UUID) (int, error) {
	return 10, nil
}

func (f *fakeAuth) ResetOperatorTOTPByAdmin(tenantID, adminID, targetOperatorID uuid.UUID) (string, error) {
	return "new-setup-token", nil
}

func (f *fakeAuth) RequestRecovery(tenantID uuid.UUID, identifier string) (string, error) {
	return "recovery-token-xyz", nil
}

func (f *fakeAuth) ValidateRecoveryToken(rawToken string) (domain.RecoveryToken, domain.Operator, error) {
	return domain.RecoveryToken{ID: uuid.New(), OperatorID: uuid.New()}, domain.Operator{ID: uuid.New(), Email: "op@test.com"}, nil
}

func (f *fakeAuth) GetTenantSetupStatus(tenantID uuid.UUID) (domain.TenantSetupStatus, error) {
	return domain.TenantSetupStatus{SetupStep: 1, IsSetupComplete: false}, nil
}

func (f *fakeAuth) UpdateTenantSetup(tenantID uuid.UUID, setupStep int, orgDetails map[string]any) (domain.TenantSetupStatus, error) {
	return domain.TenantSetupStatus{SetupStep: setupStep, IsSetupComplete: false}, nil
}

func (f *fakeAuth) CompleteTenantSetup(tenantID uuid.UUID) error {
	return nil
}

func (f *fakeAuth) GetTenantByID(tenantID uuid.UUID) (domain.Tenant, error) {
	return domain.Tenant{ID: tenantID, Name: "Test Tenant"}, nil
}

func (f *fakeAuth) TrackInvitationDelivery(invitationID uuid.UUID, status, messageID, errorMessage string) error {
	return nil
}

func (f *fakeAuth) LogRecoveryAudit(tenantID uuid.UUID, operatorID *uuid.UUID, action, ip, userAgent string, details map[string]any) error {
	return nil
}

func (f *fakeAuth) LogPermissionCheck(ctx context.Context, operatorID uuid.UUID, action, resource string, resourceID uuid.UUID, allowed bool, reason, ip, ua string) error {
	return nil
}

type fakeWhatsApp struct {
	lastTo      string
	lastMessage string
	shouldFail  bool
}

func (fw *fakeWhatsApp) SendInvitation(to, message string) error {
	if fw.shouldFail {
		return errors.New("delivery failed")
	}
	fw.lastTo = to
	fw.lastMessage = message
	return nil
}

func TestDashboardLogin(t *testing.T) {
	tenantID := uuid.New()
	auth := newFakeAuth(tenantID, "op@example.com", "secret")
	d := &DashboardHandler{Auth: auth, StaticFS: nil}

	body := `{"email":"op@example.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/api/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant", tenantID.String())

	rr := httptest.NewRecorder()
	d.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	cookies := rr.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookieName {
		t.Fatalf("expected session cookie, got %v", cookies)
	}
	if cookies[0].Path != "/" {
		t.Fatalf("expected session cookie path '/', got %q", cookies[0].Path)
	}
}

func TestDashboardLoginWrongPassword(t *testing.T) {
	tenantID := uuid.New()
	auth := newFakeAuth(tenantID, "op@example.com", "secret")
	d := &DashboardHandler{Auth: auth}

	body := `{"email":"op@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/api/login", bytes.NewBufferString(body))
	req.Header.Set("X-Tenant", tenantID.String())

	rr := httptest.NewRecorder()
	d.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDashboardMeUnauthenticated(t *testing.T) {
	auth := newFakeAuth(uuid.New(), "op@example.com", "secret")
	d := &DashboardHandler{Auth: auth}

	req := httptest.NewRequest(http.MethodGet, "/dashboard/api/me", nil)
	rr := httptest.NewRecorder()
	d.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestServeDashboardStatic(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data:    []byte("<!DOCTYPE html><html><body>Dashboard Root</body></html>"),
			Mode:    0644,
			ModTime: time.Now(),
		},
		"assets/app.js": &fstest.MapFile{
			Data:    []byte("console.log('app')"),
			Mode:    0644,
			ModTime: time.Now(),
		},
	}

	d := &DashboardHandler{StaticFS: mockFS}

	testCases := []struct {
		name         string
		path         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "root dashboard",
			path:         "/dashboard",
			expectedCode: http.StatusOK,
			expectedBody: "Dashboard Root",
		},
		{
			name:         "trailing slash dashboard",
			path:         "/dashboard/",
			expectedCode: http.StatusOK,
			expectedBody: "Dashboard Root",
		},
		{
			name:         "SPA route fallback",
			path:         "/dashboard/inbox",
			expectedCode: http.StatusOK,
			expectedBody: "Dashboard Root",
		},
		{
			name:         "SPA route login fallback",
			path:         "/dashboard/login",
			expectedCode: http.StatusOK,
			expectedBody: "Dashboard Root",
		},
		{
			name:         "static asset serving",
			path:         "/dashboard/assets/app.js",
			expectedCode: http.StatusOK,
			expectedBody: "console.log('app')",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rr := httptest.NewRecorder()
			d.ServeHTTP(rr, req)

			if rr.Code != tc.expectedCode {
				t.Fatalf("path %s: expected status %d, got %d (headers: %v)", tc.path, tc.expectedCode, rr.Code, rr.Header())
			}
			if !strings.Contains(rr.Body.String(), tc.expectedBody) {
				t.Fatalf("path %s: expected body containing %q, got %q", tc.path, tc.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestDashboardSignupTenant(t *testing.T) {
	auth := newFakeAuth(uuid.New(), "admin@test.com", "secret")
	d := &DashboardHandler{Auth: auth}

	body := `{"tenant_name":"Acme Corp","admin_name":"Admin Alice","admin_email":"admin@acme.com","admin_whatsapp":"+1234567890"}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/api/signup/tenant", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	d.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "verification_token") {
		t.Fatalf("expected verification_token in body, got: %s", rr.Body.String())
	}
}

func TestDashboardVerifyEmail(t *testing.T) {
	auth := newFakeAuth(uuid.New(), "admin@test.com", "secret")
	d := &DashboardHandler{Auth: auth}

	body := `{"token":"test-verify-token"}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/api/verify-email", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	d.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "setup_token") {
		t.Fatalf("expected setup_token in body, got: %s", rr.Body.String())
	}
}

func TestDashboardTOTPSetupAndVerify(t *testing.T) {
	auth := newFakeAuth(uuid.New(), "admin@test.com", "secret")
	d := &DashboardHandler{Auth: auth}

	// GET setup info
	req := httptest.NewRequest(http.MethodGet, "/dashboard/api/totp/setup/test-setup-token", nil)
	rr := httptest.NewRecorder()
	d.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for TOTP setup, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "qr_code_data_url") {
		t.Fatalf("expected qr_code_data_url in response: %s", rr.Body.String())
	}

	// POST verify setup
	verifyBody := `{"setup_token":"test-setup-token","code":"123456"}`
	verifyReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/totp/verify-setup", bytes.NewBufferString(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRR := httptest.NewRecorder()
	d.ServeHTTP(verifyRR, verifyReq)
	if verifyRR.Code != http.StatusOK {
		t.Fatalf("expected 200 for verify setup, got %d: %s", verifyRR.Code, verifyRR.Body.String())
	}
	if !strings.Contains(verifyRR.Body.String(), "backup_codes") {
		t.Fatalf("expected backup_codes in response: %s", verifyRR.Body.String())
	}
}

func TestDashboardOperatorInvitationAndSignup(t *testing.T) {
	auth := newFakeAuth(uuid.New(), "admin@test.com", "secret")
	d := &DashboardHandler{Auth: auth}

	// GET accept invite info
	req := httptest.NewRequest(http.MethodGet, "/dashboard/api/invitations/accept/invite-token-123", nil)
	rr := httptest.NewRecorder()
	d.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for accept invite, got %d: %s", rr.Code, rr.Body.String())
	}

	// POST operator signup
	signupBody := `{"invitation_token":"invite-token-123","name":"Bob Operator","totp_code":"123456"}`
	signupReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/signup/operator", bytes.NewBufferString(signupBody))
	signupReq.Header.Set("Content-Type", "application/json")
	signupRR := httptest.NewRecorder()
	d.ServeHTTP(signupRR, signupReq)
	if signupRR.Code != http.StatusCreated {
		t.Fatalf("expected 201 for operator signup, got %d: %s", signupRR.Code, signupRR.Body.String())
	}
}

func TestDashboardBackupCodeLoginAndRecovery(t *testing.T) {
	tenantID := uuid.New()
	auth := newFakeAuth(tenantID, "admin@test.com", "secret")
	d := &DashboardHandler{Auth: auth}

	// Backup code login
	loginBody := `{"tenant_id":"` + tenantID.String() + `","identifier":"admin@test.com","code":"1111-2222"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/login/backup-code", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	d.ServeHTTP(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("expected 200 for backup code login, got %d: %s", loginRR.Code, loginRR.Body.String())
	}

	// Recovery request
	recBody := `{"tenant_id":"` + tenantID.String() + `","identifier":"admin@test.com"}`
	recReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/recovery/request", bytes.NewBufferString(recBody))
	recReq.Header.Set("Content-Type", "application/json")
	recRR := httptest.NewRecorder()
	d.ServeHTTP(recRR, recReq)
	if recRR.Code != http.StatusOK {
		t.Fatalf("expected 200 for recovery request, got %d: %s", recRR.Code, recRR.Body.String())
	}
}

func TestDashboardCreateWhatsAppInvitation(t *testing.T) {
	tenantID := uuid.New()
	auth := newFakeAuthWithRole(tenantID, "admin@test.com", "secret", "admin")
	fakeWA := &fakeWhatsApp{}
	d := &DashboardHandler{Auth: auth, WhatsApp: fakeWA}

	// First login to get a session
	loginBody := `{"email":"admin@test.com","password":"secret"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-Tenant", tenantID.String())
	loginRR := httptest.NewRecorder()
	d.ServeHTTP(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("login failed: %d", loginRR.Code)
	}
	cookie := loginRR.Result().Cookies()[0]

	// Send WhatsApp invitation
	inviteBody := `{"whatsapp_number":"+1234567890","role":"operator"}`
	inviteReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/invitations/whatsapp", bytes.NewBufferString(inviteBody))
	inviteReq.Header.Set("Content-Type", "application/json")
	inviteReq.AddCookie(cookie)

	inviteRR := httptest.NewRecorder()
	d.ServeHTTP(inviteRR, inviteReq)

	if inviteRR.Code != http.StatusCreated {
		t.Fatalf("expected 201 for whatsapp invitation, got %d: %s", inviteRR.Code, inviteRR.Body.String())
	}
	if !strings.Contains(inviteRR.Body.String(), `"whatsapp_sent":true`) {
		t.Fatalf("expected whatsapp_sent:true in response: %s", inviteRR.Body.String())
	}
	if fakeWA.lastTo != "+1234567890" {
		t.Fatalf("expected whatsapp sent to +1234567890, got: %s", fakeWA.lastTo)
	}
	if !strings.Contains(fakeWA.lastMessage, "invite-token-123") {
		t.Fatalf("expected invite token in whatsapp message: %s", fakeWA.lastMessage)
	}

	// Test WhatsApp failure fallback
	fakeWA.shouldFail = true
	failRR := httptest.NewRecorder()
	failReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/invitations/whatsapp", bytes.NewBufferString(inviteBody))
	failReq.Header.Set("Content-Type", "application/json")
	failReq.AddCookie(cookie)
	d.ServeHTTP(failRR, failReq)

	if failRR.Code != http.StatusCreated {
		t.Fatalf("expected 201 for fallback invitation, got %d: %s", failRR.Code, failRR.Body.String())
	}
	if !strings.Contains(failRR.Body.String(), `"whatsapp_sent":false`) {
		t.Fatalf("expected whatsapp_sent:false in response: %s", failRR.Body.String())
	}
	if !strings.Contains(failRR.Body.String(), "manual_instructions") {
		t.Fatalf("expected manual_instructions in response: %s", failRR.Body.String())
	}
}

type fakePairingService struct {
	startID   string
	startErr  error
	snapshots map[string]whatsapp.PairingSnapshot
	cancelErr error
	cancelled []string
}

func (f *fakePairingService) Start(tenantID uuid.UUID, displayName string) (string, error) {
	if f.startErr != nil {
		return "", f.startErr
	}
	id := f.startID
	if id == "" {
		id = "pairing-test-id"
	}
	return id, nil
}

func (f *fakePairingService) Get(id string) (whatsapp.PairingSnapshot, bool) {
	if f.snapshots == nil {
		return whatsapp.PairingSnapshot{}, false
	}
	snap, ok := f.snapshots[id]
	return snap, ok
}

func (f *fakePairingService) Cancel(id string) error {
	if f.cancelErr != nil {
		return f.cancelErr
	}
	f.cancelled = append(f.cancelled, id)
	return nil
}

func TestDashboardPairingAPI(t *testing.T) {
	tenantID := uuid.New()
	adminAuth := newFakeAuthWithRole(tenantID, "admin@test.com", "secret", "admin")
	opAuth := newFakeAuthWithRole(tenantID, "op@test.com", "secret", "operator")

	adminSession, _ := adminAuth.CreateSession(adminAuth.operators["admin@test.com"].op.ID, tenantID, time.Hour)
	opSession, _ := opAuth.CreateSession(opAuth.operators["op@test.com"].op.ID, tenantID, time.Hour)

	fakePairing := &fakePairingService{
		startID: "test-pair-1",
		snapshots: map[string]whatsapp.PairingSnapshot{
			"test-pair-1": {
				ID:        "test-pair-1",
				Status:    whatsapp.PairingStatusAwaitingScan,
				QRDataURL: "data:image/png;base64,mockqr",
			},
		},
	}

	adminAPI := &DashboardAPIHandler{
		Auth:    adminAuth,
		Pairing: fakePairing,
	}
	adminHandler := DashboardSessionMiddleware(adminAuth, adminAPI)

	opAPI := &DashboardAPIHandler{
		Auth:    opAuth,
		Pairing: fakePairing,
	}
	opHandler := DashboardSessionMiddleware(opAuth, opAPI)

	// 1. Admin start pairing -> 201
	startReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/pairing", bytes.NewBufferString(`{"display_name":"Office Phone"}`))
	startReq.Header.Set("Content-Type", "application/json")
	startReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: adminSession.ID.String()})
	startRR := httptest.NewRecorder()
	adminHandler.ServeHTTP(startRR, startReq)

	if startRR.Code != http.StatusCreated {
		t.Fatalf("expected 201 for start pairing, got %d: %s", startRR.Code, startRR.Body.String())
	}
	if !strings.Contains(startRR.Body.String(), `"pairing_id":"test-pair-1"`) {
		t.Fatalf("expected pairing_id in response, got %s", startRR.Body.String())
	}

	// 2. Operator start pairing -> 403
	opStartReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/pairing", bytes.NewBufferString(`{}`))
	opStartReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: opSession.ID.String()})
	opStartRR := httptest.NewRecorder()
	opHandler.ServeHTTP(opStartRR, opStartReq)

	if opStartRR.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for operator start pairing, got %d: %s", opStartRR.Code, opStartRR.Body.String())
	}

	// 3. Admin get pairing snapshot -> 200
	getReq := httptest.NewRequest(http.MethodGet, "/dashboard/api/pairing/test-pair-1", nil)
	getReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: adminSession.ID.String()})
	getRR := httptest.NewRecorder()
	adminHandler.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected 200 for get pairing, got %d: %s", getRR.Code, getRR.Body.String())
	}
	if !strings.Contains(getRR.Body.String(), `"status":"awaiting_scan"`) {
		t.Fatalf("expected awaiting_scan status, got %s", getRR.Body.String())
	}

	// 4. Admin get non-existent pairing snapshot -> 404
	badGetReq := httptest.NewRequest(http.MethodGet, "/dashboard/api/pairing/unknown-id", nil)
	badGetReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: adminSession.ID.String()})
	badGetRR := httptest.NewRecorder()
	adminHandler.ServeHTTP(badGetRR, badGetReq)

	if badGetRR.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown pairing, got %d", badGetRR.Code)
	}

	// 5. Admin cancel pairing -> 200
	cancelReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/pairing/test-pair-1/cancel", nil)
	cancelReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: adminSession.ID.String()})
	cancelRR := httptest.NewRecorder()
	adminHandler.ServeHTTP(cancelRR, cancelReq)

	if cancelRR.Code != http.StatusOK {
		t.Fatalf("expected 200 for cancel pairing, got %d: %s", cancelRR.Code, cancelRR.Body.String())
	}
	if len(fakePairing.cancelled) != 1 || fakePairing.cancelled[0] != "test-pair-1" {
		t.Fatalf("expected test-pair-1 cancelled, got %v", fakePairing.cancelled)
	}

	// 6. Operator cancel pairing -> 403
	opCancelReq := httptest.NewRequest(http.MethodPost, "/dashboard/api/pairing/test-pair-1/cancel", nil)
	opCancelReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: opSession.ID.String()})
	opCancelRR := httptest.NewRecorder()
	opHandler.ServeHTTP(opCancelRR, opCancelReq)

	if opCancelRR.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for operator cancel pairing, got %d", opCancelRR.Code)
	}
}
