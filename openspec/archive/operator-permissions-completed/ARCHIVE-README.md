# Archived: 01 — Operator Permissions (RBAC)

**Archived**: August 16, 2025  
**Status**: ✅ FULLY IMPLEMENTED

This change has been fully implemented, tested, and archived. The specification
is preserved here for historical reference.

## What Was Implemented

Role-based access control (`admin`, `operator`, `viewer`) enforced at the
handler/middleware layer with asynchronous permission audit logging.

### Files
- `migrations/0007_operator_permissions.up.sql` / `.down.sql` — `operator_permission_checks` audit table + indexes
- `handler/permissions.go` — permission constants, `RolePermissions` matrix, `HasPermission`, `RequirePermission`/`RequireAnyPermission` middleware, async `auditPermission`
- `handler/permissions_test.go` — unit tests
- `internal/storage/permissions.go` — `PostgresStore.LogPermissionCheck`
- `internal/storage/permissions_test.go` — sqlmock tests
- Modified: `handler/dashboard.go`, `handler/dashboard_api.go`, `handler/dashboard_test.go`, `main.go`

## Roles & Permissions

| Role | Scope |
|------|-------|
| `admin` | Full access — operators, invitations, tenant setup, audit |
| `operator` | Operational access — view/close conversations, send messages, notes, view accounts; **no** operator/tenant management |
| `viewer` | Read-only — view conversations, accounts, bot rules |

## Acceptance Criteria Met

- Viewer → `POST /dashboard/api/operators` returns **403** ✅
- Permission checks logged with operator ID, action, resource, IP, user agent ✅
- `go build ./...` and `go test ./...` pass ✅

## Related

- `../../changes/agent-implementation-plan/` — active implementation plan (gaps 02-16)
- `../tenant-onboarding-flow-completed/` — TOTP onboarding (archived)
- `../manual-onboarding-whatsapp-completed/` — WhatsApp invitations (archived)
