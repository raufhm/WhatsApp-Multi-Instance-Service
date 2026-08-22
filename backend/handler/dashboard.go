package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
	"github.com/raufhm/whops/internal/storage"
	"github.com/raufhm/whops/internal/totp"
)

// sessionCookieName is the HttpOnly session cookie used by the dashboard.
const sessionCookieName = "sid"

// dashboardAuth is the subset of storage functionality the dashboard needs.
type dashboardAuth interface {
	storage.OperatorAuth
	LogPermissionCheck(ctx context.Context, operatorID uuid.UUID, action, resource string, resourceID uuid.UUID, allowed bool, reason, ip, ua string) error
}

type whatsappSender interface {
	SendInvitation(to, message string) error
}

// contextKey is used to store the tenant ID in the request context.
type contextKey struct{}

// operatorIDContextKey is used to store the authenticated operator ID in the
// request context by DashboardSessionMiddleware.
type operatorIDContextKey struct{}

// DashboardSessionMiddleware validates the session cookie and injects the
// operator's tenant ID into the request context.
func DashboardSessionMiddleware(auth dashboardAuth, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookieName)
		if err != nil {
			writeAPIError(w, 401, "UNAUTHORIZED", "session cookie required")
			return
		}
		sid, err := uuid.Parse(c.Value)
		if err != nil {
			writeAPIError(w, 401, "UNAUTHORIZED", "invalid session cookie")
			return
		}
		s, err := auth.GetSessionByID(sid)
		if err != nil {
			writeAPIError(w, 401, "UNAUTHORIZED", "session expired or invalid")
			return
		}
		ctx := context.WithValue(r.Context(), contextKey{}, s.TenantID)
		ctx = context.WithValue(ctx, operatorIDContextKey{}, s.OperatorID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// tenantFromContext returns the tenant ID placed by DashboardSessionMiddleware.
func tenantFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(contextKey{}).(uuid.UUID)
	return id, ok
}

// operatorIDFromContext returns the operator ID placed by DashboardSessionMiddleware.
func operatorIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(operatorIDContextKey{}).(uuid.UUID)
	return id, ok
}

func serveIndex(w http.ResponseWriter, r *http.Request, staticFS fs.FS) {
	if staticFS == nil {
		http.Error(w, "Dashboard frontend not available", http.StatusNotFound)
		return
	}
	f, err := staticFS.Open("index.html")
	if err != nil {
		http.Error(w, "Dashboard frontend not built (run 'npm run build' inside frontend/)", http.StatusNotFound)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	if seeker, ok := f.(io.ReadSeeker); ok {
		http.ServeContent(w, r, "index.html", stat.ModTime(), seeker)
		return
	}

	data, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, "index.html", stat.ModTime(), bytes.NewReader(data))
}

// ServeDashboardStatic serves the built frontend from an embedded filesystem.
// The dist directory is embedded at build time via go:embed. Paths under
// /dashboard/* are mapped directly into the dist FS; any non-asset path falls
// back to index.html so the React Router SPA can render.
func ServeDashboardStatic(staticFS fs.FS) http.Handler {
	fileServer := http.StripPrefix("/dashboard", http.FileServer(http.FS(staticFS)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle GET and HEAD requests.
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Strip the /dashboard mount prefix so paths are relative to dist/.
		path := strings.TrimPrefix(r.URL.Path, "/dashboard")
		path = strings.TrimPrefix(path, "/")

		// Root path or direct index.html: serve dist/index.html directly without redirect.
		if path == "" || path == "index.html" {
			serveIndex(w, r, staticFS)
			return
		}

		// If the path points to a real static asset (file, not directory), serve it.
		if staticFS != nil {
			if f, err := staticFS.Open(path); err == nil {
				stat, statErr := f.Stat()
				f.Close()
				if statErr == nil && !stat.IsDir() {
					if strings.HasPrefix(path, "assets/") {
						w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
					}
					fileServer.ServeHTTP(w, r)
					return
				}
			}
		}

		// SPA fallback for any other route (e.g., /dashboard/inbox, /dashboard/login).
		serveIndex(w, r, staticFS)
	})
}

// DashboardHandler is the top-level dashboard router. It dispatches the auth
// API and static asset serving under a single mount.
type DashboardHandler struct {
	Auth     dashboardAuth
	WhatsApp whatsappSender
	StaticFS fs.FS
}

func (d *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/dashboard")
	path = strings.Trim(path, "/")

	// Auth & Onboarding API routes.
	if strings.HasPrefix(path, "api/") {
		// Public onboarding & auth endpoints
		switch {
		case path == "api/signup/tenant" && r.Method == http.MethodPost:
			d.handleSignupTenant(w, r)
			return
		case path == "api/verify-email" && r.Method == http.MethodPost:
			d.handleVerifyEmail(w, r)
			return
		case strings.HasPrefix(path, "api/totp/setup/") && r.Method == http.MethodGet:
			token := strings.TrimPrefix(path, "api/totp/setup/")
			d.handleTOTPSetup(w, r, token)
			return
		case path == "api/totp/verify-setup" && r.Method == http.MethodPost:
			d.handleTOTPVerifySetup(w, r)
			return
		case strings.HasPrefix(path, "api/invitations/accept/") && r.Method == http.MethodGet:
			token := strings.TrimPrefix(path, "api/invitations/accept/")
			d.handleAcceptInvitationInfo(w, r, token)
			return
		case path == "api/signup/operator" && r.Method == http.MethodPost:
			d.handleSignupOperator(w, r)
			return
		case path == "api/login" && r.Method == http.MethodPost:
			d.handleLogin(w, r)
			return
		case path == "api/login/backup-code" && r.Method == http.MethodPost:
			d.handleBackupCodeLogin(w, r)
			return
		case path == "api/recovery/request" && r.Method == http.MethodPost:
			d.handleRecoveryRequest(w, r)
			return
		case strings.HasPrefix(path, "api/recovery/") && r.Method == http.MethodGet:
			token := strings.TrimPrefix(path, "api/recovery/")
			d.handleRecoveryToken(w, r, token)
			return
		case path == "api/logout" && r.Method == http.MethodPost:
			d.handleLogout(w, r)
			return
		}

		// Protected endpoints (require valid session)
		op, ok := d.currentOperator(r)
		if !ok {
			writeAPIError(w, 401, "UNAUTHORIZED", "not authenticated")
			return
		}

		switch {
		case path == "api/me" && r.Method == http.MethodGet:
			d.handleMeWithOp(w, r, op)
			return
		case path == "api/account/totp" && r.Method == http.MethodGet:
			d.handleGetAccountTOTP(w, r, op.TenantID, op.ID)
			return
		case path == "api/account/totp/regenerate-backup-codes" && r.Method == http.MethodPost:
			d.handleRegenerateBackupCodes(w, r, op.ID)
			return
		case path == "api/operators" && r.Method == http.MethodGet:
			if !d.requirePermission(w, r, op, PermViewOperators) {
				return
			}
			d.handleListOperators(w, r, op.TenantID)
			return
		case strings.HasPrefix(path, "api/operators/") && strings.HasSuffix(path, "/totp-reset") && r.Method == http.MethodPost:
			if !d.requirePermission(w, r, op, PermResetTotp) {
				return
			}
			targetIDStr := strings.TrimPrefix(path, "api/operators/")
			targetIDStr = strings.TrimSuffix(targetIDStr, "/totp-reset")
			targetID, err := uuid.Parse(targetIDStr)
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid operator id")
				return
			}
			d.handleAdminResetTOTP(w, r, op.TenantID, op.ID, targetID)
			return
		case strings.HasPrefix(path, "api/operators/") && strings.HasSuffix(path, "/totp-status") && r.Method == http.MethodGet:
			if !d.requirePermission(w, r, op, PermResetTotp) {
				return
			}
			targetIDStr := strings.TrimPrefix(path, "api/operators/")
			targetIDStr = strings.TrimSuffix(targetIDStr, "/totp-status")
			targetID, err := uuid.Parse(targetIDStr)
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid operator id")
				return
			}
			d.handleAdminGetTOTPStatus(w, r, op.TenantID, op.ID, targetID)
			return
		case path == "api/invitations/whatsapp" && r.Method == http.MethodPost:
			if !d.requirePermission(w, r, op, PermInviteOperators) {
				return
			}
			d.handleCreateWhatsAppInvitation(w, r, op.TenantID, op.ID)
			return
		case path == "api/invitations/email" && r.Method == http.MethodPost:
			if !d.requirePermission(w, r, op, PermInviteOperators) {
				return
			}
			d.handleCreateEmailInvitation(w, r, op.TenantID, op.ID)
			return
		case path == "api/invitations" && r.Method == http.MethodGet:
			if !d.requirePermission(w, r, op, PermInviteOperators) {
				return
			}
			d.handleListInvitations(w, r, op.TenantID)
			return
		case strings.HasPrefix(path, "api/invitations/") && r.Method == http.MethodDelete:
			if !d.requirePermission(w, r, op, PermRevokeInvitation) {
				return
			}
			invIDStr := strings.TrimPrefix(path, "api/invitations/")
			invID, err := uuid.Parse(invIDStr)
			if err != nil {
				writeAPIError(w, 400, "INVALID_ID", "invalid invitation id")
				return
			}
			d.handleRevokeInvitation(w, r, op.TenantID, invID)
			return
		case path == "api/tenant/setup-status" && r.Method == http.MethodGet:
			d.handleGetTenantSetupStatus(w, r, op.TenantID)
			return
		case path == "api/tenant/setup" && r.Method == http.MethodPut:
			if !d.requirePermission(w, r, op, PermManageTenant) {
				return
			}
			d.handleUpdateTenantSetup(w, r, op.TenantID)
			return
		case path == "api/tenant/complete-setup" && r.Method == http.MethodPost:
			if !d.requirePermission(w, r, op, PermManageTenant) {
				return
			}
			d.handleCompleteTenantSetup(w, r, op.TenantID)
			return
		default:
			writeAPIError(w, 404, "NOT_FOUND", "dashboard api endpoint not found")
			return
		}
	}

	// Static assets / SPA.
	ServeDashboardStatic(d.StaticFS).ServeHTTP(w, r)
}

type loginRequest struct {
	TenantID   string `json:"tenant_id"`
	TenantSlug string `json:"tenant_slug"`
	Identifier string `json:"identifier"`
	Email      string `json:"email"`
	Whatsapp   string `json:"whatsapp_number"`
	Password   string `json:"password"`
	Code       string `json:"code"`
	TOTPCode   string `json:"totp_code"`
	RememberMe bool   `json:"remember_me"`
}

var errTenantRequired = errors.New("tenant identifier is required")

func (d *DashboardHandler) resolveTenant(r *http.Request, explicitID, explicitSlug string) (domain.Tenant, error) {
	tenantStr := strings.TrimSpace(explicitSlug)
	if tenantStr == "" {
		tenantStr = strings.TrimSpace(explicitID)
	}
	if tenantStr == "" {
		tenantStr = strings.TrimSpace(r.Header.Get("X-Tenant-Slug"))
	}
	if tenantStr == "" {
		tenantStr = strings.TrimSpace(r.Header.Get("X-Tenant"))
	}
	if tenantStr == "" {
		return domain.Tenant{}, errTenantRequired
	}

	// If it parses as a UUID, attempt to look it up by ID first
	if parsedUUID, err := uuid.Parse(tenantStr); err == nil {
		tenant, err := d.Auth.GetTenantByID(parsedUUID)
		if err == nil {
			return tenant, nil
		}
	}

	// Look up by slug
	tenant, err := d.Auth.FindTenantBySlug(tenantStr)
	if err != nil {
		return domain.Tenant{}, err
	}
	return tenant, nil
}

func (d *DashboardHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}

	tenant, err := d.resolveTenant(r, req.TenantID, req.TenantSlug)
	if err != nil {
		if errors.Is(err, errTenantRequired) {
			writeAPIError(w, 400, "TENANT_REQUIRED", "tenant ID or company name is required")
			return
		}
		writeAPIError(w, 401, "INVALID_CREDENTIALS", "invalid credentials")
		return
	}

	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		identifier = strings.TrimSpace(req.Email)
	}
	if identifier == "" {
		identifier = strings.TrimSpace(req.Whatsapp)
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		code = strings.TrimSpace(req.TOTPCode)
	}

	if identifier == "" {
		writeAPIError(w, 400, "INVALID_REQUEST", "email or whatsapp identifier is required")
		return
	}

	op, plainSecret, hash, err := d.Auth.FindOperatorByIdentifier(tenant.ID, identifier)
	if err != nil {
		writeAPIError(w, 401, "INVALID_CREDENTIALS", "invalid credentials")
		return
	}
	if !op.IsActive {
		writeAPIError(w, 403, "ACCOUNT_DISABLED", "operator account is disabled")
		return
	}

	// If TOTP code provided, verify TOTP
	if code != "" && plainSecret != "" {
		if !totp.VerifyTOTP(plainSecret, code) {
			writeAPIError(w, 401, "INVALID_CREDENTIALS", "invalid totp code")
			return
		}
	} else if req.Password != "" && hash != "" {
		if !storage.VerifyOperatorPassword(hash, req.Password) {
			writeAPIError(w, 401, "INVALID_CREDENTIALS", "invalid email or password")
			return
		}
	} else if code != "" && plainSecret == "" {
		writeAPIError(w, 401, "INVALID_CREDENTIALS", "totp not set up for account")
		return
	} else {
		writeAPIError(w, 400, "INVALID_REQUEST", "totp code or password required")
		return
	}

	ttl := 8 * time.Hour
	if req.RememberMe {
		ttl = 30 * 24 * time.Hour
	}

	session, err := d.Auth.CreateSession(op.ID, tenant.ID, ttl)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "could not create session")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	})

	WriteJSON(w, 200, map[string]any{
		"user":        op,
		"session_id":  session.ID,
		"tenant_id":   tenant.ID,
		"tenant_slug": tenant.Slug,
		"tenant_name": tenant.Name,
	})
}

func (d *DashboardHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookieName)
	if err == nil {
		if sid, perr := uuid.Parse(c.Value); perr == nil {
			_ = d.Auth.DeleteSession(sid)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	WriteJSON(w, 200, map[string]any{"status": "logged_out"})
}

func (d *DashboardHandler) currentOperator(r *http.Request) (domain.Operator, bool) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return domain.Operator{}, false
	}
	sid, err := uuid.Parse(c.Value)
	if err != nil {
		return domain.Operator{}, false
	}
	s, err := d.Auth.GetSessionByID(sid)
	if err != nil {
		return domain.Operator{}, false
	}
	op, err := d.Auth.GetOperatorByID(s.TenantID, s.OperatorID)
	if err != nil {
		return domain.Operator{}, false
	}
	return op, true
}

// requirePermission enforces a single permission for the current operator and
// records the outcome in the audit trail. It writes a 403 response and returns
// false when the operator is not permitted.
func (d *DashboardHandler) requirePermission(w http.ResponseWriter, r *http.Request, op domain.Operator, perm Permission) bool {
	allowed := HasPermission(op.Role, perm)
	auditPermission(r.Context(), d.Auth, op, perm, extractResource(r), extractResourceID(r), allowed, getClientIP(r), r.UserAgent())
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions for action: "+string(perm))
		return false
	}
	return true
}

func (d *DashboardHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	op, ok := d.currentOperator(r)
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "not authenticated")
		return
	}
	d.handleMeWithOp(w, r, op)
}

func (d *DashboardHandler) handleMeWithOp(w http.ResponseWriter, r *http.Request, op domain.Operator) {
	full, err := d.Auth.GetOperatorByID(op.TenantID, op.ID)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "could not load operator")
		return
	}
	tenantName := ""
	tenantSlug := ""
	if tenant, err := d.Auth.GetTenantByID(op.TenantID); err == nil {
		tenantName = tenant.Name
		tenantSlug = tenant.Slug
	}
	WriteJSON(w, 200, map[string]any{
		"user":        full,
		"tenant_id":   op.TenantID,
		"tenant_slug": tenantSlug,
		"tenant_name": tenantName,
	})
}

// Logger is kept for clarity; replace with the project logger if needed.
var _ = log.Printf
