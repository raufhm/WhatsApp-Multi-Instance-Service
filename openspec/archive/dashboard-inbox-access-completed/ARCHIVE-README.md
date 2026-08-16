# Archived: Dashboard inbox & messaging access

**Archived**: August 16, 2026
**Status**: ✅ FULLY IMPLEMENTED

This change has been fully implemented, built, and tested. The specification is
preserved here for historical reference.

## What Was Implemented

Bridged the gap between the dashboard's session auth and the API-key-only
`/api/v1/*` inbox endpoints, so a logged-in operator can read and write the
inbox without an API key.

### Session-aware API auth
- `handler/http.go` — `Server.Auth sessionAuth` field; `Server.tenant()` now
  resolves tenant from API key first, then `?token=`/`?api_key=` query params,
  then a valid `sid` session cookie (`GetSessionByID`)
- `handler/operator_actions.go` — `operatorID` converted to a method that
  resolves the operator from the session, falling back to `X-Actor`/`"api"`;
  all callers updated to `s.operatorID(r)`
- `handler/notes.go`, `handler/merge_audit.go` — use `s.operatorID(r)`

### Cookie scope
- `handler/dashboard.go` — `sid` cookie `Path` broadened from `/dashboard` to `/`
  in both `handleLogin` and `handleBackupCodeLogin`

### Wiring
- `main.go` — `handler.Server{Auth: pgStore, ...}`

## Acceptance Criteria Met

- Session-authenticated inbox reads/writes accepted without an API key ✅
- API-key access preserved and still first in precedence ✅
- Missing/expired session returns 401; tenant scoping retained ✅
- Operator identity recorded from session (not `"api"`) ✅
- `go build ./...` and `go test ./...` pass ✅

## Related

- `../qr-pairing-completed/` — onboarding that precedes this inbox access (archived)
- `../operator-permissions-completed/` — RBAC/permission layer (archived)
- `../agent-implementation-plan/` — API key lifecycle (gap 02) remains open
