# Operator Roles and Permissions Specification

**Priority**: CRITICAL  
**Effort**: 2-3 days  
**Agent**: Backend Specialist  
**Status**: READY FOR IMPLEMENTATION

---

## Overview

Implement role-based access control (RBAC) for the operator dashboard to enforce permission checks on all sensitive operations. Currently, the `operators` table has a `role` column but it's not enforced in the application layer.

---

## Current State

### Database Schema

```sql
-- operators table already has role column
CREATE TABLE operators (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    email TEXT,
    whatsapp_number TEXT,
    name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'operator',  -- EXISTS BUT NOT ENFORCED
    is_active BOOLEAN DEFAULT true,
    -- ... other columns
);
```

### Problem

- Role column exists but no enforcement
- Any authenticated operator can perform admin actions
- No permission audit trail
- Security risk in production

---

## Required Implementation

### 1. Database Migration

**File**: `migrations/0007_operator_permissions.up.sql`

```sql
-- Add constraint to ensure valid roles
ALTER TABLE operators 
ADD CONSTRAINT chk_operator_role 
CHECK (role IN ('admin', 'operator', 'viewer'));

-- Create permission audit table
CREATE TABLE operator_permission_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    resource TEXT,
    resource_id UUID,
    allowed BOOLEAN NOT NULL,
    reason TEXT,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_permission_checks_operator ON operator_permission_checks(operator_id);
CREATE INDEX idx_permission_checks_action ON operator_permission_checks(action);
CREATE INDEX idx_permission_checks_created ON operator_permission_checks(created_at);

-- Add comment explaining the table
COMMENT ON TABLE operator_permission_checks IS 
'Audit log for all permission checks, successful and denied';
```

**Rollback**: `migrations/0007_operator_permissions.down.sql`

```sql
DROP TABLE operator_permission_checks;
ALTER TABLE operators DROP CONSTRAINT chk_operator_role;
```

---

### 2. Permission System Core

**File**: `handler/permissions.go` (NEW)

```go
package handler

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/google/uuid"
)

// Permission represents a specific action that can be authorized
type Permission string

const (
    // Operator management
    PermCreateOperator    Permission = "operators:create"
    PermUpdateOperator    Permission = "operators:update"
    PermDeleteOperator    Permission = "operators:delete"
    PermViewOperators     Permission = "operators:view"
    PermResetTotp         Permission = "operators:reset_totp"
    
    // Audit logs
    PermViewAuditLogs     Permission = "audit:read"
    PermViewPermissionChecks Permission = "permissions:read"
    
    // Bot rules
    PermManageBotRules    Permission = "bot_rules:manage"
    PermViewBotRules      Permission = "bot_rules:view"
    
    // Conversations
    PermCloseConversation Permission = "conversations:close"
    PermReopenConversation Permission = "conversations:reopen"
    PermHandoffConversation Permission = "conversations:handoff"
    PermAssignConversation Permission = "conversations:assign"
    PermViewConversations Permission = "conversations:view"
    PermSendMessages      Permission = "messages:send"
    PermAddInternalNote   Permission = "notes:create"
    
    // Team settings
    PermManageTeam        Permission = "team:manage"
    PermInviteOperators   Permission = "invitations:create"
    PermRevokeInvitation  Permission = "invitations:revoke"
    
    // Account/WhatsApp
    PermManageAccounts    Permission = "accounts:manage"
    PermViewAccounts      Permission = "accounts:view"
    
    // API Keys
    PermManageApiKeys     Permission = "api_keys:manage"
)

// RolePermissions maps roles to their granted permissions
var RolePermissions = map[string][]Permission{
    "admin": {
        // All permissions
        PermCreateOperator, PermUpdateOperator, PermDeleteOperator, PermViewOperators, PermResetTotp,
        PermViewAuditLogs, PermViewPermissionChecks,
        PermManageBotRules, PermViewBotRules,
        PermCloseConversation, PermReopenConversation, PermHandoffConversation,
        PermAssignConversation, PermViewConversations, PermSendMessages, PermAddInternalNote,
        PermManageTeam, PermInviteOperators, PermRevokeInvitation,
        PermManageAccounts, PermViewAccounts,
        PermManageApiKeys,
    },
    "operator": {
        // Standard operator permissions - no admin/management
        PermViewOperators,
        PermViewBotRules,
        PermCloseConversation, PermReopenConversation, PermHandoffConversation,
        PermAssignConversation, PermViewConversations, PermSendMessages, PermAddInternalNote,
        PermViewAccounts,
    },
    "viewer": {
        // Read-only access
        PermViewOperators,
        PermViewBotRules,
        PermViewConversations,
        PermViewAccounts,
    },
}

// HasPermission checks if an operator has a specific permission
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
```

---

### 3. Permission Middleware

**File**: `handler/permissions.go` (continued)

```go
// RequirePermission creates middleware that checks for a specific permission
func RequirePermission(required Permission) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Get operator from context (set by auth middleware)
            op := getOperatorFromContext(r.Context())
            if op == nil {
                writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
                return
            }
            
            // Check permission
            allowed := HasPermission(op.Role, required)
            
            // Log permission check
            logPermissionCheck(
                r.Context(),
                op.ID,
                op.TenantID,
                required,
                allowed,
                extractResource(r),
                extractResourceID(r),
                getClientIP(r),
                r.UserAgent(),
            )
            
            if !allowed {
                writeAPIError(
                    w,
                    http.StatusForbidden,
                    "FORBIDDEN",
                    "insufficient permissions for action: "+string(required),
                )
                return
            }
            
            // Permission granted, proceed
            next.ServeHTTP(w, r)
        })
    }
}

// RequireAnyPermission creates middleware that checks for any of the given permissions
func RequireAnyPermission(permissions ...Permission) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            op := getOperatorFromContext(r.Context())
            if op == nil {
                writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
                return
            }
            
            allowed := false
            for _, perm := range permissions {
                if HasPermission(op.Role, perm) {
                    allowed = true
                    break
                }
            }
            
            // Log permission check
            logPermissionCheck(
                r.Context(),
                op.ID,
                op.TenantID,
                permissions[0], // Log first permission as representative
                allowed,
                extractResource(r),
                extractResourceID(r),
                getClientIP(r),
                r.UserAgent(),
            )
            
            if !allowed {
                writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Helper functions
func logPermissionCheck(
    ctx context.Context,
    operatorID, tenantID uuid.UUID,
    action Permission,
    allowed bool,
    resource string,
    resourceID uuid.UUID,
    ip string,
    ua string,
) {
    // Get storage from context or use global
    store := getStoreFromContext(ctx)
    if store == nil {
        // Fallback: log but don't block
        return
    }
    
    reason := "allowed"
    if !allowed {
        reason = "denied"
    }
    
    // Async logging to avoid blocking request
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        err := store.LogPermissionCheck(ctx, operatorID, action, resource, resourceID, allowed, reason, ip, ua)
        if err != nil {
            // Log error but don't fail the request
            log.Printf("failed to log permission check: %v", err)
        }
    }()
}

func extractResource(r *http.Request) string {
    // Extract resource type from URL path
    // e.g., /dashboard/api/operators -> "operators"
    path := r.URL.Path
    // Simple parsing - can be enhanced
    if len(path) > 0 && path[0] == '/' {
        path = path[1:]
    }
    parts := strings.Split(path, "/")
    if len(parts) >= 3 {
        return parts[2] // dashboard/api/RESOURCE/...
    }
    return "unknown"
}

func extractResourceID(r *http.Request) uuid.UUID {
    // Extract resource ID from URL if present
    // e.g., /dashboard/api/operators/:id -> returns the ID
    path := r.URL.Path
    parts := strings.Split(path, "/")
    if len(parts) >= 4 {
        id, err := uuid.Parse(parts[3])
        if err == nil {
            return id
        }
    }
    return uuid.Nil
}
```

---

### 4. Storage Layer

**File**: `internal/storage/permissions.go` (NEW)

```go
package storage

import (
    "context"
    "database/sql"
    "time"

    "github.com/google/uuid"
    "github.com/lib/pq"
)

func (p *PostgresStore) LogPermissionCheck(
    ctx context.Context,
    operatorID uuid.UUID,
    action Permission,
    resource string,
    resourceID uuid.UUID,
    allowed bool,
    reason string,
    ip string,
    ua string,
) error {
    query := `
        INSERT INTO operator_permission_checks
        (operator_id, action, resource, resource_id, allowed, reason, ip_address, user_agent, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
    `
    
    _, err := p.db.ExecContext(ctx, query,
        operatorID,
        string(action),
        nullString(resource),
        nullUUID(resourceID),
        allowed,
        reason,
        nullString(ip),
        nullString(ua),
    )
    
    return err
}

// Helper functions for null values
func nullString(s string) interface{} {
    if s == "" {
        return nil
    }
    return s
}

func nullUUID(id uuid.UUID) interface{} {
    if id == uuid.Nil {
        return nil
    }
    return id
}
```

---

### 5. Update Existing Handlers

**File**: `handler/dashboard_api.go` (MODIFY)

Add permission middleware to existing routes:

```go
func (d *DashboardAPIHandler) setupRoutes() {
    d.handlers = map[string]http.Handler{
        // Operator management - ADMIN ONLY
        "GET /operators": RequirePermission(PermViewOperators)(
            http.HandlerFunc(d.handleListOperators),
        ),
        "POST /operators": RequirePermission(PermCreateOperator)(
            http.HandlerFunc(d.handleCreateOperator),
        ),
        "PUT /operators/:id": RequirePermission(PermUpdateOperator)(
            http.HandlerFunc(d.handleUpdateOperator),
        ),
        "DELETE /operators/:id": RequirePermission(PermDeleteOperator)(
            http.HandlerFunc(d.handleDeleteOperator),
        ),
        "POST /operators/:id/totp-reset": RequirePermission(PermResetTotp)(
            http.HandlerFunc(d.handleTotpReset),
        ),
        
        // Audit logs - ADMIN ONLY
        "GET /audit-logs": RequirePermission(PermViewAuditLogs)(
            http.HandlerFunc(d.handleGetAuditLogs),
        ),
        "GET /permission-checks": RequirePermission(PermViewPermissionChecks)(
            http.HandlerFunc(d.handleGetPermissionChecks),
        ),
        
        // Bot rules
        "GET /bot-rules": RequireAnyPermission(PermViewBotRules, PermManageBotRules)(
            http.HandlerFunc(d.handleGetBotRules),
        ),
        "POST /bot-rules/activate": RequirePermission(PermManageBotRules)(
            http.HandlerFunc(d.handleActivateBotRules),
        ),
        
        // Conversations
        "GET /conversations": RequirePermission(PermViewConversations)(
            http.HandlerFunc(d.handleListConversations),
        ),
        "GET /conversations/:id": RequirePermission(PermViewConversations)(
            http.HandlerFunc(d.handleGetConversation),
        ),
        "POST /conversations/:id/assign": RequirePermission(PermAssignConversation)(
            http.HandlerFunc(d.handleAssignConversation),
        ),
        "POST /conversations/:id/close": RequirePermission(PermCloseConversation)(
            http.HandlerFunc(d.handleCloseConversation),
        ),
        "POST /conversations/:id/reopen": RequirePermission(PermReopenConversation)(
            http.HandlerFunc(d.handleReopenConversation),
        ),
        "POST /conversations/:id/note": RequirePermission(PermAddInternalNote)(
            http.HandlerFunc(d.handleAddNote),
        ),
        
        // Messages
        "POST /conversations/:id/message": RequirePermission(PermSendMessages)(
            http.HandlerFunc(d.handleSendMessage),
        ),
        
        // Invitations
        "GET /invitations": RequirePermission(PermInviteOperators)(
            http.HandlerFunc(d.handleListInvitations),
        ),
        "POST /invitations/whatsapp": RequirePermission(PermInviteOperators)(
            http.HandlerFunc(d.handleCreateWhatsAppInvitation),
        ),
        "POST /invitations/email": RequirePermission(PermInviteOperators)(
            http.HandlerFunc(d.handleCreateEmailInvitation),
        ),
        "DELETE /invitations/:id": RequirePermission(PermRevokeInvitation)(
            http.HandlerFunc(d.handleRevokeInvitation),
        ),
        
        // Team management
        "GET /team": RequirePermission(PermManageTeam)(
            http.HandlerFunc(d.handleGetTeam),
        ),
        
        // Accounts
        "GET /accounts": RequireAnyPermission(PermViewAccounts, PermManageAccounts)(
            http.HandlerFunc(d.handleListAccounts),
        ),
        "POST /accounts": RequirePermission(PermManageAccounts)(
            http.HandlerFunc(d.handleCreateAccount),
        ),
    }
}
```

---

### 6. Tests

**File**: `handler/permissions_test.go` (NEW)

```go
package handler

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
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
        {"unknown role has no permissions", "unknown", PermViewConversations, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := HasPermission(tt.role, tt.permission)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestRequirePermission(t *testing.T) {
    admin := &domain.Operator{
        ID:       uuid.New(),
        TenantID: uuid.New(),
        Role:     "admin",
    }
    
    viewer := &domain.Operator{
        ID:       uuid.New(),
        TenantID: uuid.New(),
        Role:     "viewer",
    }
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    protected := RequirePermission(PermCreateOperator)(handler)
    
    t.Run("admin can access", func(t *testing.T) {
        req := httptest.NewRequest("POST", "/dashboard/api/operators", nil)
        req = req.WithContext(context.WithValue(req.Context(), operatorContextKey, admin))
        rr := httptest.NewRecorder()
        
        protected.ServeHTTP(rr, req)
        
        assert.Equal(t, http.StatusOK, rr.Code)
    })
    
    t.Run("viewer cannot access", func(t *testing.T) {
        req := httptest.NewRequest("POST", "/dashboard/api/operators", nil)
        req = req.WithContext(context.WithValue(req.Context(), operatorContextKey, viewer))
        rr := httptest.NewRecorder()
        
        protected.ServeHTTP(rr, req)
        
        assert.Equal(t, http.StatusForbidden, rr.Code)
        assert.Contains(t, rr.Body.String(), "insufficient permissions")
    })
}

func TestPermissionCheckLogging(t *testing.T) {
    // Test that permission checks are logged
    // This requires a mock store or test database
}
```

---

## Acceptance Criteria

- ✅ Viewer role cannot POST to `/dashboard/api/operators` (returns 403)
- ✅ Admin can perform all administrative actions
- ✅ Operator has limited permissions (no admin/management)
- ✅ All permission checks logged to `operator_permission_checks` table
- ✅ Permission check logs include IP and user agent
- ✅ Tests passing (unit + integration)
- ✅ Migration runs successfully
- ✅ Rollback migration works
- ✅ No breaking changes to existing functionality

---

## Files to Create/Modify

### Create
- `migrations/0007_operator_permissions.up.sql`
- `migrations/0007_operator_permissions.down.sql`
- `handler/permissions.go`
- `handler/permissions_test.go`
- `internal/storage/permissions.go`

### Modify
- `handler/dashboard_api.go` (add permission middleware to routes)
- `handler/http.go` (add helper functions if needed)

---

## Implementation Steps

1. **Create database migration**
   - Write up migration
   - Write down migration
   - Test migration locally

2. **Implement permission system core**
   - Define permissions
   - Define role mappings
   - Implement `HasPermission()` function

3. **Implement middleware**
   - Create `RequirePermission()` middleware
   - Create `RequireAnyPermission()` middleware
   - Implement logging

4. **Implement storage layer**
   - Add `LogPermissionCheck()` method
   - Test with PostgreSQL

5. **Update handlers**
   - Wrap all routes with appropriate permission checks
   - Ensure no routes are missed

6. **Write tests**
   - Unit tests for permission logic
   - Integration tests for middleware
   - Test all roles

7. **Manual testing**
   - Test as admin
   - Test as operator
   - Test as viewer
   - Verify logs

8. **Documentation**
   - Update API docs
   - Add permission matrix to README

---

## Security Considerations

- **Defense in depth**: Permission checks at middleware level, but also validate in business logic
- **Fail closed**: If permission system fails, deny access rather than grant
- **Audit trail**: All permission checks logged, both allowed and denied
- **Least privilege**: Default roles should have minimal necessary permissions
- **Regular review**: Permission matrix should be reviewed periodically

---

## Performance Considerations

- **Async logging**: Permission check logging is asynchronous to avoid blocking requests
- **In-memory permission map**: RolePermissions is a static map for O(1) lookup
- **Database indexing**: Permission check logs are indexed for efficient querying
- **Connection pooling**: Uses existing PostgreSQL connection pool

---

## Related Specifications

- [`02-api-key-lifecycle/spec.md`](../02-api-key-lifecycle/spec.md) - API key management
- [`04-audit-logging/spec.md`](../04-audit-logging/spec.md) - Comprehensive audit logging

---

## Testing Checklist

- [ ] Admin can create operators
- [ ] Operator cannot create operators
- [ ] Viewer cannot view audit logs
- [ ] All permission checks are logged
- [ ] Logs include IP and user agent
- [ ] Migration runs without errors
- [ ] Rollback works
- [ ] No existing functionality broken
- [ ] Performance acceptable under load
