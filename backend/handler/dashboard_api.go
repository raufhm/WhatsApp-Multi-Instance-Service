package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
	"github.com/raufhm/whops/internal/broadcast"
	"github.com/raufhm/whops/internal/storage"
	"github.com/raufhm/whops/whatsapp"
)

// DashboardAPIHandler routes session-authenticated endpoints used by the UI.
// All routes require a valid operator session.
type DashboardAPIHandler struct {
	Platform    domain.PlatformRepository
	Manager     accountManager
	Pairing     pairingService
	Store       dashboardDataStore
	Auth        dashboardAuth
	MediaStore  storage.MediaStore
	S3ObjectURL string
	Monitor     domain.MonitoringStore
	Broadcaster *broadcast.Broadcaster
}

// pairingService defines the pairing manager capabilities required by the dashboard.
type pairingService interface {
	Start(tenantID uuid.UUID, displayName string) (string, error)
	Get(id string) (whatsapp.PairingSnapshot, bool)
	Cancel(id string) error
}

// accountManager is the subset of WhatsAppManager used by the dashboard.
type accountManager interface {
	ListInstances() []domain.InstanceInfo
	GetInstance(host string) (domain.InstanceInfo, error)
	Disconnect(host string) error
	Reconnect(host string) error
}

// dashboardDataStore exposes dashboard-only data access.
type dashboardDataStore interface {
	ListUploadJobs(tenantID uuid.UUID, status string, limit, offset int) ([]domain.UploadJob, error)
	CreateOperator(tenantID uuid.UUID, email, name, role, password string) (domain.Operator, error)
}

func (d *DashboardAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := tenantFromContext(r.Context())
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "dashboard session required")
		return
	}

	path := r.URL.Path
	switch {
	case path == "/dashboard/api/accounts" && r.Method == http.MethodGet:
		if !d.requirePermission(w, r, tenantID, PermViewAccounts) {
			return
		}
		d.listAccounts(w, r, tenantID)
		return
	case strings.HasPrefix(path, "/dashboard/api/accounts/") && strings.HasSuffix(path, "/disconnect") && r.Method == http.MethodPost:
		if !d.requirePermission(w, r, tenantID, PermManageAccounts) {
			return
		}
		d.disconnectAccount(w, r, tenantID)
		return
	case strings.HasPrefix(path, "/dashboard/api/accounts/") && strings.HasSuffix(path, "/reconnect") && r.Method == http.MethodPost:
		if !d.requirePermission(w, r, tenantID, PermManageAccounts) {
			return
		}
		d.reconnectAccount(w, r, tenantID)
		return
	case (path == "/dashboard/api/accounts" || path == "/dashboard/api/pairing") && r.Method == http.MethodPost:
		if !d.requirePermission(w, r, tenantID, PermManageAccounts) {
			return
		}
		d.startPairing(w, r, tenantID)
		return
	case strings.HasPrefix(path, "/dashboard/api/pairing/") && strings.HasSuffix(path, "/cancel") && r.Method == http.MethodPost:
		if !d.requirePermission(w, r, tenantID, PermManageAccounts) {
			return
		}
		d.cancelPairing(w, r, tenantID)
		return
	case strings.HasPrefix(path, "/dashboard/api/pairing/") && r.Method == http.MethodGet:
		if !d.requirePermission(w, r, tenantID, PermManageAccounts) {
			return
		}
		d.getPairing(w, r, tenantID)
		return
	case path == "/dashboard/api/bot-rules" && r.Method == http.MethodGet:
		if !d.requirePermission(w, r, tenantID, PermViewBotRules) {
			return
		}
		d.listBotRules(w, r, tenantID)
		return
	case path == "/dashboard/api/bot-rules" && r.Method == http.MethodPost:
		if !d.requirePermission(w, r, tenantID, PermManageBotRules) {
			return
		}
		d.createBotRuleSet(w, r, tenantID)
		return
	case path == "/dashboard/api/bot-rules/activate" && r.Method == http.MethodPost:
		if !d.requirePermission(w, r, tenantID, PermManageBotRules) {
			return
		}
		d.activateBotRuleSet(w, r, tenantID)
		return
	case path == "/dashboard/api/upload-jobs" && r.Method == http.MethodGet:
		d.listUploadJobs(w, r, tenantID)
		return
	case path == "/dashboard/api/media" && r.Method == http.MethodPost:
		d.uploadMedia(w, r, tenantID)
		return
	case strings.HasPrefix(path, "/dashboard/api/media/") && r.Method == http.MethodGet:
		d.serveMedia(w, r, tenantID)
		return
	case path == "/dashboard/api/operators" && r.Method == http.MethodPost:
		if !d.requirePermission(w, r, tenantID, PermCreateOperator) {
			return
		}
		d.createOperator(w, r, tenantID)
		return
	case path == "/dashboard/api/monitoring/status" && r.Method == http.MethodGet:
		if !d.requirePermission(w, r, tenantID, PermViewAccounts) {
			return
		}
		d.monitoringStatus(w, r, tenantID)
		return
	case path == "/dashboard/api/monitoring/status-events" && r.Method == http.MethodGet:
		if !d.requirePermission(w, r, tenantID, PermViewAccounts) {
			return
		}
		d.monitoringStatusEvents(w, r, tenantID)
		return
	case path == "/dashboard/api/monitoring/metrics" && r.Method == http.MethodGet:
		if !d.requirePermission(w, r, tenantID, PermViewAccounts) {
			return
		}
		d.monitoringMetrics(w, r, tenantID)
		return
	case path == "/dashboard/api/monitoring/events" && r.Method == http.MethodGet:
		if !d.requirePermission(w, r, tenantID, PermViewAccounts) {
			return
		}
		d.monitoringEvents(w, r, tenantID)
		return
	case path == "/dashboard/api/monitoring/queue-depth" && r.Method == http.MethodGet:
		if !d.requirePermission(w, r, tenantID, PermViewAccounts) {
			return
		}
		d.monitoringQueueDepth(w, r, tenantID)
		return
	case path == "/dashboard/api/monitoring/stream" && r.Method == http.MethodGet:
		if !d.requirePermission(w, r, tenantID, PermViewAccounts) {
			return
		}
		d.monitoringStream(w, r, tenantID)
		return
	}
	writeAPIError(w, 404, "NOT_FOUND", "dashboard api endpoint not found")
}

// requirePermission enforces a permission for the operator identified by the
// session middleware and records the outcome in the audit trail. It returns
// false and writes a response when access is denied.
func (d *DashboardAPIHandler) requirePermission(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID, perm Permission) bool {
	opID, ok := operatorIDFromContext(r.Context())
	if !ok || d.Auth == nil {
		writeAPIError(w, 401, "UNAUTHORIZED", "operator not found")
		return false
	}
	op, err := d.Auth.GetOperatorByID(tenantID, opID)
	if err != nil {
		writeAPIError(w, 401, "UNAUTHORIZED", "operator not found")
		return false
	}
	allowed := HasPermission(op.Role, perm)
	auditPermission(r.Context(), d.Auth, op, perm, extractResource(r), extractResourceID(r), allowed, getClientIP(r), r.UserAgent())
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions for action: "+string(perm))
		return false
	}
	return true
}

type accountWithHealth struct {
	domain.WhatsAppAccount
	Health      string `json:"health"`
	IsConnected bool   `json:"is_connected"`
	QueueSize   int    `json:"queue_size"`
}

func (d *DashboardAPIHandler) listAccounts(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	accounts, err := d.Platform.ListAccounts(tenantID)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to list accounts")
		return
	}

	instances := d.Manager.ListInstances()
	instanceMap := make(map[string]domain.InstanceInfo, len(instances))
	for _, i := range instances {
		instanceMap[i.HostPhone] = i
	}

	result := make([]accountWithHealth, 0, len(accounts))
	for _, a := range accounts {
		info, ok := instanceMap[a.HostID]
		health := "unknown"
		isConnected := false
		queueSize := 0
		if ok {
			if info.IsConnected {
				health = "healthy"
			} else {
				health = "disconnected"
			}
			isConnected = info.IsConnected
			queueSize = info.QueueSize
		}
		result = append(result, accountWithHealth{
			WhatsAppAccount: a,
			Health:          health,
			IsConnected:     isConnected,
			QueueSize:       queueSize,
		})
	}
	WriteJSON(w, 200, result)
}

func (d *DashboardAPIHandler) accountAction(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID, action func(host string) error) bool {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeAPIError(w, 404, "NOT_FOUND", "invalid account path")
		return false
	}
	accountID, err := uuid.Parse(parts[3])
	if err != nil {
		writeAPIError(w, 400, "INVALID_ID", "invalid account id")
		return false
	}
	account, err := d.Platform.GetAccount(tenantID, accountID)
	if err != nil {
		writeAPIError(w, 404, "NOT_FOUND", "account not found")
		return false
	}
	if err := action(account.HostID); err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", err.Error())
		return false
	}

	// Re-read instance state after the action so the response reflects the
	// current connection status.
	info, err := d.Manager.GetInstance(account.HostID)
	isConnected := err == nil && info.IsConnected
	WriteJSON(w, 200, accountWithHealth{
		WhatsAppAccount: account,
		Health:          "",
		IsConnected:     isConnected,
		QueueSize:       0,
	})
	return true
}

func (d *DashboardAPIHandler) disconnectAccount(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	d.accountAction(w, r, tenantID, d.Manager.Disconnect)
}

func (d *DashboardAPIHandler) reconnectAccount(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	d.accountAction(w, r, tenantID, d.Manager.Reconnect)
}

func (d *DashboardAPIHandler) pairingService() pairingService {
	if d.Pairing != nil {
		return d.Pairing
	}
	if wm, ok := d.Manager.(*whatsapp.WhatsAppManager); ok && wm != nil && wm.Pairing != nil {
		return wm.Pairing
	}
	return nil
}

func (d *DashboardAPIHandler) startPairing(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	pairing := d.pairingService()
	if pairing == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "pairing service unavailable")
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
	}
	if r.Body != nil && r.ContentLength > 0 {
		_ = DecodeJSONBody(r, &req)
	}

	pairingID, err := pairing.Start(tenantID, req.DisplayName)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", fmt.Sprintf("failed to start pairing: %v", err))
		return
	}

	WriteJSON(w, 201, map[string]string{
		"pairing_id": pairingID,
	})
}

func (d *DashboardAPIHandler) getPairing(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	pairing := d.pairingService()
	if pairing == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "pairing service unavailable")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/dashboard/api/pairing/")
	if id == "" || strings.Contains(id, "/") {
		writeAPIError(w, 404, "NOT_FOUND", "invalid pairing id")
		return
	}

	snap, found := pairing.Get(id)
	if !found {
		writeAPIError(w, 404, "NOT_FOUND", "pairing session not found")
		return
	}

	WriteJSON(w, 200, snap)
}

func (d *DashboardAPIHandler) cancelPairing(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	pairing := d.pairingService()
	if pairing == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "pairing service unavailable")
		return
	}

	trimmed := strings.TrimPrefix(r.URL.Path, "/dashboard/api/pairing/")
	id := strings.TrimSuffix(trimmed, "/cancel")
	if id == "" || strings.Contains(id, "/") {
		writeAPIError(w, 404, "NOT_FOUND", "invalid pairing id")
		return
	}

	if err := pairing.Cancel(id); err != nil {
		writeAPIError(w, 404, "NOT_FOUND", err.Error())
		return
	}

	WriteJSON(w, 200, map[string]string{"status": "cancelled"})
}

func (d *DashboardAPIHandler) listBotRules(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	sets, err := d.Platform.ListBotRuleSets(tenantID)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to list bot rules")
		return
	}
	if sets == nil {
		sets = []domain.BotRuleSet{}
	}
	WriteJSON(w, 200, sets)
}

func (d *DashboardAPIHandler) createBotRuleSet(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	var req struct {
		Rules []domain.BotRule `json:"rules"`
	}
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}
	if err := ValidateBotRules(req.Rules); err != nil {
		writeAPIError(w, 400, "VALIDATION_ERROR", err.Error())
		return
	}
	set, err := d.Platform.SaveBotRuleSet(tenantID, req.Rules)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to save bot rules")
		return
	}
	WriteJSON(w, 201, set)
}

func (d *DashboardAPIHandler) activateBotRuleSet(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	versionStr := r.URL.Query().Get("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		writeAPIError(w, 400, "INVALID_VERSION", "version must be an integer")
		return
	}
	set, err := d.Platform.ActivateBotRuleSetVersion(tenantID, version)
	if err != nil {
		writeAPIError(w, 404, "NOT_FOUND", err.Error())
		return
	}
	WriteJSON(w, 200, set)
}

func (d *DashboardAPIHandler) listUploadJobs(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	status := r.URL.Query().Get("status")
	limit, offset, valid := page(r)
	if !valid {
		writeAPIError(w, 400, "INVALID_PAGINATION", "invalid pagination")
		return
	}
	jobs, err := d.Store.ListUploadJobs(tenantID, status, limit, offset)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to list upload jobs")
		return
	}
	if jobs == nil {
		jobs = []domain.UploadJob{}
	}
	WriteJSON(w, 200, jobs)
}

func (d *DashboardAPIHandler) createOperator(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Role     string `json:"role"`
		Password string `json:"password"`
	}
	if err := DecodeJSONBody(r, &req); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeAPIError(w, 400, "VALIDATION_ERROR", "email and password are required")
		return
	}
	op, err := d.Store.CreateOperator(tenantID, req.Email, req.Name, req.Role, req.Password)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to create operator")
		return
	}
	WriteJSON(w, 201, op)
}

func (d *DashboardAPIHandler) uploadMedia(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	if d.MediaStore == nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "media store not configured")
		return
	}

	// Limit upload size to 50MB
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "file exceeds maximum allowed size (50MB) or malformed multipart form")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "missing 'file' form field")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to read uploaded file")
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = http.DetectContentType(data)
	}

	key := fmt.Sprintf("media/%s", uuid.NewString())
	if err := d.MediaStore.Put(r.Context(), key, mimeType, data); err != nil {
		writeAPIError(w, 500, "INTERNAL_ERROR", "failed to store media")
		return
	}

	if d.Platform != nil {
		if err := d.Platform.RecordMediaObject(r.Context(), tenantID, key, mimeType, int64(len(data))); err != nil {
			log.Printf("Failed to record media object in DB: %v", err)
		}
	}

	mediaURL := storage.ResolveMediaURL(key, d.S3ObjectURL, "dashboard")

	WriteJSON(w, http.StatusCreated, map[string]any{
		"media_key": key,
		"media_url": mediaURL,
		"mime_type": mimeType,
		"size":      len(data),
	})
}

func (d *DashboardAPIHandler) serveMedia(w http.ResponseWriter, r *http.Request, tenantID uuid.UUID) {
	key := strings.TrimPrefix(r.URL.Path, "/dashboard/api/media/")
	if key == "" || d.MediaStore == nil {
		writeAPIError(w, 404, "NOT_FOUND", "media not found")
		return
	}

	obj, err := d.Platform.GetMediaObject(r.Context(), tenantID, key)
	if err != nil {
		writeAPIError(w, 404, "NOT_FOUND", "media not found")
		return
	}

	rc, err := d.MediaStore.Open(r.Context(), key)
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
}
