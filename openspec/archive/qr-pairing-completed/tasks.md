# Tasks

## Phase 1 — Account registration (backend foundation)

- [x] Add `RegisterAccount` to the repository boundary and `PostgresStore`.
- [x] Add a domain/repository interface method for idempotent upsert into `whatsapp_accounts`.
- [x] Wire `RegisterAccount` so a paired host is tenant-scoped and visible in `ListAccounts`.

## Phase 2 — Pairing manager

- [x] Add `whatsapp.PairingManager` with in-memory session map + mutex.
- [x] Implement `Start` (create device, connect, open QR channel, event goroutine).
- [x] Handle `code`, `success`, `timeout`, and error channel events.
- [x] On success: resolve `HostPhone`, call `RegisterAccount`, then `SpawnInstance`.
- [x] Implement `Cancel`, session TTL, and janitor reaping.
- [x] Extract `EncodeQRDataURL` helper from `OnboardHandler`.

## Phase 3 — HTTP endpoints

- [x] `POST /dashboard/api/pairing` (start).
- [x] `GET /dashboard/api/pairing/{id}` (poll snapshot).
- [x] `POST /dashboard/api/pairing/{id}/cancel`.
- [x] Gate all endpoints with `DashboardSessionMiddleware` + `PermManageAccounts`.
- [x] Update legacy `/api/onboard` to register the account (or mark superseded).

## Phase 4 — Frontend

- [x] Add "Add account" button to Accounts page.
- [x] Build pairing modal with QR render + status + cancel.
- [x] Poll pairing state every ~2s and invalidate accounts on success.
- [x] Update SetupWizard to link to the real pairing flow.

## Phase 5 — Tests and docs

- [x] Unit tests for `PairingManager` state transitions and registration.
- [x] Storage tests for `RegisterAccount` (idempotent upsert).
- [x] Handler tests for permission gating and endpoint behavior.
- [x] Update README (pairing workflow) and openspec master/roadmap references.
