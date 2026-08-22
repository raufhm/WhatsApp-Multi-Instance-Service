package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whops/domain"
)

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		permission Permission
		expected   bool
	}{
		{"admin can create operator", "admin", PermCreateOperator, true},
		{"operator cannot create operator", "operator", PermCreateOperator, false},
		{"viewer cannot create operator", "viewer", PermCreateOperator, false},
		{"operator can view conversations", "operator", PermViewConversations, true},
		{"viewer can view conversations", "viewer", PermViewConversations, true},
		{"operator can close conversations", "operator", PermCloseConversation, true},
		{"viewer cannot close conversations", "viewer", PermCloseConversation, false},
		{"operator cannot manage tenant", "operator", PermManageTenant, false},
		{"operator cannot invite operators", "operator", PermInviteOperators, false},
		{"admin can invite operators", "admin", PermInviteOperators, true},
		{"unknown role has no permissions", "unknown", PermViewConversations, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPermission(tt.role, tt.permission); got != tt.expected {
				t.Fatalf("HasPermission(%q, %q) = %v, want %v", tt.role, tt.permission, got, tt.expected)
			}
		})
	}
}

func TestRequirePermission(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	admin := domain.Operator{ID: uuid.New(), TenantID: uuid.New(), Role: "admin"}
	viewer := domain.Operator{ID: uuid.New(), TenantID: uuid.New(), Role: "viewer"}

	handler := RequirePermission(PermCreateOperator)(next)

	t.Run("admin can access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/dashboard/api/operators", nil)
		req = req.WithContext(withOperator(req.Context(), admin))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("viewer cannot access", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/dashboard/api/operators", nil)
		req = req.WithContext(withOperator(req.Context(), viewer))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("missing operator is unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/dashboard/api/operators", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
		}
	})
}

func TestRequireAnyPermission(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	operator := domain.Operator{ID: uuid.New(), TenantID: uuid.New(), Role: "operator"}
	viewer := domain.Operator{ID: uuid.New(), TenantID: uuid.New(), Role: "viewer"}

	t.Run("operator with one matching permission passes", func(t *testing.T) {
		handler := RequireAnyPermission(PermViewAccounts, PermManageAccounts)(next)
		req := httptest.NewRequest(http.MethodGet, "/dashboard/api/accounts", nil)
		req = req.WithContext(withOperator(req.Context(), operator))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("viewer without any matching permission denied", func(t *testing.T) {
		handler := RequireAnyPermission(PermManageBotRules, PermCreateOperator)(next)
		req := httptest.NewRequest(http.MethodGet, "/dashboard/api/bot-rules", nil)
		req = req.WithContext(withOperator(req.Context(), viewer))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
		}
	})
}

type permissionCheckRecord struct {
	operatorID uuid.UUID
	action     string
	allowed    bool
}

// recordingAuditor captures permission checks for assertions.
type recordingAuditor struct {
	checks chan permissionCheckRecord
}

func (r *recordingAuditor) LogPermissionCheck(ctx context.Context, operatorID uuid.UUID, action, resource string, resourceID uuid.UUID, allowed bool, reason, ip, ua string) error {
	r.checks <- permissionCheckRecord{operatorID: operatorID, action: action, allowed: allowed}
	return nil
}

func TestAuditPermission(t *testing.T) {
	auditor := &recordingAuditor{checks: make(chan permissionCheckRecord, 1)}
	op := domain.Operator{ID: uuid.New(), TenantID: uuid.New(), Role: "viewer"}

	auditPermission(context.Background(), auditor, op, PermCreateOperator, "operators", uuid.Nil, false, "127.0.0.1", "test-agent")

	select {
	case rec := <-auditor.checks:
		if rec.operatorID != op.ID {
			t.Fatalf("expected operator %s, got %s", op.ID, rec.operatorID)
		}
		if rec.action != string(PermCreateOperator) {
			t.Fatalf("expected action %q, got %q", PermCreateOperator, rec.action)
		}
		if rec.allowed {
			t.Fatalf("expected allowed=false, got true")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for audit log")
	}
}

func TestExtractResource(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/dashboard/api/operators", nil)
	if got := extractResource(req); got != "operators" {
		t.Fatalf("expected operators, got %q", got)
	}
}

func TestExtractResourceID(t *testing.T) {
	id := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/dashboard/api/operators/"+id.String()+"/totp-reset", nil)
	if got := extractResourceID(req); got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestGetClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.5:443"
	if got := getClientIP(req); got != "192.168.1.5" {
		t.Fatalf("expected 192.168.1.5, got %q", got)
	}
}
