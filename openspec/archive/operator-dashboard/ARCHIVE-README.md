# Archived: Operator Dashboard

**Archived**: August 16, 2026
**Status**: ✅ IMPLEMENTED (core workflow complete; some polish items deferred)

This change has been implemented and the dashboard is functional. It is archived
here for historical reference. Remaining polish items are noted below.

## What Was Implemented

A React + TypeScript + TanStack + Tailwind dashboard served from `/dashboard/*`,
backed by session authentication (TOTP, no passwords) and consuming the existing
`/api/v1/*` and `/dashboard/api/*` endpoints.

### Frontend
- `frontend/` — Vite + React + TanStack Router/Query + Tailwind
- Pages: `Inbox`, `ConversationDetail`, `Contacts`, `Accounts`, `BotRules`,
  `UploadJobs`, `Team`, `Login`, `SignupTenant`, `SignupChoice`,
  `JoinWithCode`, `VerifyEmail`, `SetupWizard`, `TotpSettings`, `Recovery`,
  `OperatorInvitation`
- `frontend/src/hooks/useInbox.ts`, `useDashboard.ts`, `useAuth.tsx`
- `frontend/src/lib/apiClient.ts` — axios client with session cookie + `X-Tenant`

### Backend
- `handler/dashboard.go` — session middleware, auth/onboarding routes, static serving
- `handler/dashboard_api.go` — protected dashboard API (accounts, bot-rules, operators, upload-jobs, media)
- `handler/permissions.go` — RBAC middleware + audit (see `operator-permissions-completed`)
- `internal/storage/*` — operators, sessions, invitations, tenant onboarding, notes, operator actions, merge audit

### Auth & onboarding
- TOTP authentication, backup codes, recovery, email verification
- WhatsApp/email operator invitations, "join with code" flow
- Tenant setup wizard (4 steps)

## Deferred / Not Completed

From `tasks.md`, the following remain open:

- [ ] Keyboard shortcuts (navigate, send, close)
- [ ] Theme and branding (light/dark mode optional)
- [ ] Accessibility audit (ARIA labels, focus management)
- [ ] Unit tests for components and hooks

## Verification Evidence

- `go build ./...` passes
- `go test ./...` passes
- `npm run build` (frontend) passes

## Related

- `../operator-permissions-completed/` — RBAC + permission audit (archived)
- `../tenant-onboarding-flow-completed/` — TOTP onboarding (archived)
- `../media-storage-and-delivery-completed/` — media serving/upload (archived)
- `../agent-implementation-plan/` — roadmap (gaps 02-16 still open)
