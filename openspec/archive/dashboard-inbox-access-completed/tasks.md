# Tasks

## Phase 1 — Session-aware API auth

- [ ] Add a `sessionAuth` interface (or reuse existing) to `handler.Server`.
- [ ] Extend `Server.tenant()` to fall back to a valid session when no API key is present.
- [ ] Derive operator identity from the session for audit/actor fields.
- [ ] Keep the API-key path unchanged and first in precedence.

## Phase 2 — Cookie scope & wiring

- [ ] Broaden the `sid` cookie `Path` from `/dashboard` to `/` in `handler/dashboard.go`.
- [ ] Wire the auth store into `handler.Server` in `main.go`.

## Phase 3 — Frontend verification

- [ ] Confirm `apiClient` sends the session cookie on `/api/v1/*` (withCredentials already set).
- [ ] Confirm the inbox loads conversations, activities, contacts, and can send messages.

## Phase 4 — Tests

- [ ] Handler tests: session-authenticated inbox read/write returns 200.
- [ ] Handler tests: missing/expired session returns 401; API key still works.
- [ ] Handler tests: cross-tenant access is rejected (tenant scoping).
- [ ] Regression: operator actions and notes record the correct operator id.

## Phase 5 — End-to-end MVP verification

- [ ] Sign up, sign in, onboard via QR.
- [ ] Receive a text and an image; both appear in the inbox timeline.
- [ ] Send a text and an image reply from the inbox; both deliver to WhatsApp.
- [ ] Update README with the verified MVP walkthrough.
