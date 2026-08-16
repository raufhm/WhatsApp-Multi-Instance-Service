package handler

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raufhm/whatsapp-testing/domain"
)

// Permission identifies a single action that can be authorized.
type Permission string

const (
	// Operator management
	PermCreateOperator Permission = "operators:create"
	PermUpdateOperator Permission = "operators:update"
	PermDeleteOperator Permission = "operators:delete"
	PermViewOperators  Permission = "operators:view"
	PermResetTotp      Permission = "operators:reset_totp"

	// Audit logs
	PermViewAuditLogs        Permission = "audit:read"
	PermViewPermissionChecks Permission = "permissions:read"

	// Bot rules
	PermManageBotRules Permission = "bot_rules:manage"
	PermViewBotRules   Permission = "bot_rules:view"

	// Conversations
	PermCloseConversation   Permission = "conversations:close"
	PermReopenConversation  Permission = "conversations:reopen"
	PermHandoffConversation Permission = "conversations:handoff"
	PermAssignConversation  Permission = "conversations:assign"
	PermViewConversations   Permission = "conversations:view"
	PermSendMessages        Permission = "messages:send"
	PermAddInternalNote     Permission = "notes:create"

	// Team settings
	PermManageTeam       Permission = "team:manage"
	PermInviteOperators  Permission = "invitations:create"
	PermRevokeInvitation Permission = "invitations:revoke"

	// Account / WhatsApp
	PermManageAccounts Permission = "accounts:manage"
	PermViewAccounts   Permission = "accounts:view"

	// Tenant setup
	PermManageTenant Permission = "tenant:manage"

	// API keys
	PermManageApiKeys Permission = "api_keys:manage"
)

// RolePermissions maps each role to the set of permissions it is granted.
var RolePermissions = map[string][]Permission{
	"admin": {
		PermCreateOperator, PermUpdateOperator, PermDeleteOperator, PermViewOperators, PermResetTotp,
		PermViewAuditLogs, PermViewPermissionChecks,
		PermManageBotRules, PermViewBotRules,
		PermCloseConversation, PermReopenConversation, PermHandoffConversation,
		PermAssignConversation, PermViewConversations, PermSendMessages, PermAddInternalNote,
		PermManageTeam, PermInviteOperators, PermRevokeInvitation,
		PermManageAccounts, PermViewAccounts,
		PermManageTenant,
		PermManageApiKeys,
	},
	"operator": {
		PermViewOperators,
		PermViewBotRules,
		PermCloseConversation, PermReopenConversation, PermHandoffConversation,
		PermAssignConversation, PermViewConversations, PermSendMessages, PermAddInternalNote,
		PermViewAccounts,
	},
	"viewer": {
		PermViewOperators,
		PermViewBotRules,
		PermViewConversations,
		PermViewAccounts,
	},
}

// HasPermission reports whether a role is granted the given permission.
func HasPermission(role string, permission Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// operatorContextKey stores the authenticated operator in the request context.
type operatorContextKey struct{}

// withOperator returns a context carrying the authenticated operator.
func withOperator(ctx context.Context, op domain.Operator) context.Context {
	return context.WithValue(ctx, operatorContextKey{}, op)
}

// operatorFromContext returns the authenticated operator, if present.
func operatorFromContext(ctx context.Context) (domain.Operator, bool) {
	op, ok := ctx.Value(operatorContextKey{}).(domain.Operator)
	return op, ok
}

// permissionAuditor persists permission checks for the audit trail. It is the
// subset of storage needed by the permission middleware and handlers.
type permissionAuditor interface {
	LogPermissionCheck(ctx context.Context, operatorID uuid.UUID, action, resource string, resourceID uuid.UUID, allowed bool, reason, ip, ua string) error
}

// RequirePermission wraps a handler so that only operators holding the given
// permission may invoke it. Denials and grants are both recorded in the audit
// trail. The operator must be injected into the request context by the caller.
func RequirePermission(required Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			op, ok := operatorFromContext(r.Context())
			if !ok {
				writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}
			if !authorize(w, r, op, required) {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission wraps a handler so that an operator holding any of the
// supplied permissions may invoke it.
func RequireAnyPermission(permissions ...Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			op, ok := operatorFromContext(r.Context())
			if !ok {
				writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}
			for _, perm := range permissions {
				if HasPermission(op.Role, perm) {
					auditPermission(r.Context(), nil, op, perm, extractResource(r), extractResourceID(r), true, getClientIP(r), r.UserAgent())
					next.ServeHTTP(w, r)
					return
				}
			}
			// Record a representative denial using the first permission.
			auditPermission(r.Context(), nil, op, permissions[0], extractResource(r), extractResourceID(r), false, getClientIP(r), r.UserAgent())
			writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		})
	}
}

// authorize checks a single permission, records the outcome, and writes a 403
// response when the operator is not permitted. It returns true on success.
func authorize(w http.ResponseWriter, r *http.Request, op domain.Operator, required Permission) bool {
	allowed := HasPermission(op.Role, required)
	auditPermission(r.Context(), nil, op, required, extractResource(r), extractResourceID(r), allowed, getClientIP(r), r.UserAgent())
	if !allowed {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions for action: "+string(required))
		return false
	}
	return true
}

// auditPermission asynchronously persists a permission check. A nil auditor is
// a no-op so handler tests can exercise permission logic without a database.
func auditPermission(ctx context.Context, auditor permissionAuditor, op domain.Operator, action Permission, resource string, resourceID uuid.UUID, allowed bool, ip, ua string) {
	if auditor == nil {
		return
	}
	reason := "allowed"
	if !allowed {
		reason = "denied"
	}
	go func() {
		logCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := auditor.LogPermissionCheck(logCtx, op.ID, string(action), resource, resourceID, allowed, reason, ip, ua); err != nil {
			log.Printf("failed to log permission check: %v", err)
		}
	}()
}

// extractResource derives the resource name from the request path, e.g.
// "/dashboard/api/operators" -> "operators".
func extractResource(r *http.Request) string {
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	// dashboard/api/RESOURCE/...
	if len(parts) >= 3 {
		return parts[2]
	}
	return "unknown"
}

// extractResourceID extracts a UUID resource id from the request path when one
// is present (e.g. "/dashboard/api/operators/<id>/totp-reset").
func extractResourceID(r *http.Request) uuid.UUID {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	for _, p := range parts {
		if id, err := uuid.Parse(p); err == nil {
			return id
		}
	}
	return uuid.Nil
}

// getClientIP returns the client IP without the port, defaulting to empty when
// it cannot be determined. Empty values are stored as SQL NULL by the storage
// layer.
func getClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	host := r.RemoteAddr
	if host == "" {
		return ""
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}
