# Design: Dashboard inbox & messaging access

## Problem

The conversation/messaging API and the dashboard use two different auth schemes
and there is no bridge between them:

- `/api/v1/*` → `Server.tenant()` → `X-API-Key` / `Authorization: Bearer`
- `/dashboard/api/*` → `DashboardSessionMiddleware` → `sid` cookie (scoped to
  `Path=/dashboard`)

The inbox UI (`frontend/src/hooks/useInbox.ts`) calls `/api/v1/*`, so it has
neither a key nor a cookie and receives `401`.

## Recommended approach: session fallback on the inbox API

Extend the public API's tenant resolution to accept a valid dashboard session
when no API key is present, and widen the session cookie scope.

### 1. Broaden the session cookie

In `handler/dashboard.go` `handleLogin` / `handleBackupCodeLogin`, change the
`sid` cookie `Path` from `/dashboard` to `/`. The `HttpOnly`, `SameSite=Lax`,
and `MaxAge` attributes stay. This makes the browser include the session cookie
on `/api/v1/*` requests.

### 2. Session-aware tenant resolution

Give `Server` a small session-auth dependency (the same storage methods the
middleware uses):

```go
type sessionAuth interface {
    GetSessionByID(id uuid.UUID) (domain.Session, error)
}
```

Update `Server.tenant(r)`:

```go
func (s *Server) tenant(r *http.Request) (uuid.UUID, bool) {
    // 1. API key (unchanged)
    // 2. else, session cookie sid -> GetSessionByID -> Session.TenantID
}
```

An `operatorID` helper is derived from the resolved session's `OperatorID` for
audit/actor fields; direct API-key callers keep the existing `X-Actor` /
`"api"` fallback.

### 3. Wiring

`main.go` already constructs `handler.Server{Platform: pgStore, ...}`. Add the
auth store (or a `sessionAuth` wrapper) to `Server` so `tenant()` can resolve
sessions. `PostgresStore` already implements `GetSessionByID`.

## Alternative considered: dedicated `/dashboard/api/*` inbox endpoints

Replicate conversations/activities/contacts/messages/operator/notes under
`/dashboard/api/*` and repoint `useInbox.ts`. Cleaner separation, but ~9
endpoints of duplication. Deferred in favour of the fallback for MVP; can be
introduced later without removing the fallback.

## Security notes

- `SameSite=Lax` mitigates cross-site request forgery; the frontend already
  sends `X-CSRF-Token`, which can optionally be verified on mutating
  `/api/v1/*` routes as a follow-up.
- Session access must resolve the same tenant as the API key would; both paths
  must reject unknown/expired sessions with `401`.
- No tenant can be inferred from a resource id alone; the session tenant is
  authoritative.

## Out of scope

- API key lifecycle (create/list/revoke UI) — tracked separately (gap 02)
- Realtime push/WebSocket inbox updates (polling is retained)
