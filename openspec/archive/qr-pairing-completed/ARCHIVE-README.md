# Archived: Scan-to-connect WhatsApp pairing

**Archived**: August 16, 2026
**Status**: ✅ FULLY IMPLEMENTED

This change has been fully implemented, built, and tested. The specification is
preserved here for historical reference.

## What Was Implemented

A browser-based QR pairing flow that connects a WhatsApp number and — critically
— links it to the operator's tenant so its messages are projected.

### Account registration (the missing link)
- `domain/models.go` — `AccountRegistrar` interface
- `internal/storage/postgres.go` — `RegisterAccount` idempotent upsert into
  `whatsapp_accounts` (tenant + host phone + provider + display name)

### Pairing manager
- `whatsapp/pairing.go` — `PairingManager` with in-memory sessions, tenant index,
  TTL janitor, `Start`/`Get`/`Cancel`, QR-channel event handling
  (`code`/`success`/`timeout`/`error`), `EncodeQRDataURL`
- `whatsapp/subsystem.go` — `ResolveJIDPhone`, `WhatsAppManager.Pairing`
- `main.go` — wires `manager.Pairing.SetStore(pgStore)` and pairing routes

### HTTP endpoints (session + `PermManageAccounts`)
- `handler/dashboard_api.go` — `POST /dashboard/api/pairing`,
  `GET /dashboard/api/pairing/{id}`, `POST /dashboard/api/pairing/{id}/cancel`;
  `POST /dashboard/api/accounts` now starts pairing
- `handler/http.go` — legacy `/api/onboard` updated to `RegisterAccount` on
  success

### Frontend
- `frontend/src/components/ui/PairingModal.tsx` — QR render, poll, status, cancel
- `frontend/src/lib/apiClient.ts` — `pairingApi`
- `frontend/src/pages/Accounts.tsx` — "Add account" button + modal
- `frontend/src/pages/SetupWizard.tsx` — replaced fake toggle with real modal
- `frontend/src/types/index.ts` — `PairingStatus`, `PairingSnapshot`

## Acceptance Criteria Met

- Authorized operator can start pairing and receive a browser-renderable QR ✅
- QR refreshes on expiry; state is pollable ✅
- Successful scan upserts `whatsapp_accounts` and spawns the instance ✅
- Endpoints gated by `PermManageAccounts` (403 for unauthorized) ✅
- `go build ./...` and `go test ./...` pass ✅
- `npm run build` (frontend) passes ✅

## Related

- `../operator-permissions-completed/` — `PermManageAccounts` RBAC (archived)
- `../tenant-onboarding-flow-completed/` — tenant/setup onboarding (archived)
- `../media-storage-and-delivery-completed/` — media serving (archived)
