# Change: Scan-to-connect WhatsApp pairing

## Why

Connecting a WhatsApp number today is incomplete and unusable from the product
surface:

1. The QR code is only emitted to **server logs** (`handler/http.go:367`
   `OnboardHandler` prints half-block terminal art and a base64 PNG to stdout).
   No HTTP endpoint returns it, so there is no way to scan it from a browser.
2. The dashboard has no "add account" flow: `POST /dashboard/api/accounts`
   returns `501 NOT_IMPLEMENTED` (`handler/dashboard_api.go:165`), and the
   Accounts page just says "use the onboarding endpoint".
3. The `SetupWizard` "pairing" is a fake manual toggle
   (`frontend/src/pages/SetupWizard.tsx`).
4. **No code ever inserts into `whatsapp_accounts`.** A paired number is never
   linked to a tenant, so `ProjectMessage` ignores its events and `ListAccounts`
   shows nothing.

## What Changes

- Add an in-memory pairing manager that drives whatsmeow's QR channel and
  exposes a pollable state machine (`awaiting_scan` → `connected` / `failed` /
  `cancelled`), regenerating the QR on expiry.
- Add a repository method to register a paired host into `whatsapp_accounts`
  (tenant + host phone + display name), idempotently.
- On successful scan, register the account and spawn the instance.
- Add session-authenticated, permission-gated dashboard endpoints to start,
  poll, and cancel pairing.
- Add an "Add account" pairing modal to the Accounts page: live QR, status,
  expiry/refresh, and auto-close on success.
- Make the legacy `/api/onboard` endpoint also register the account (or clearly
  mark it superseded).

## Impact

- Affected specs: `qr-pairing`
- Affected code: `whatsapp` (pairing manager), `internal/storage`
  (account registration), `handler` (dashboard + legacy onboard), `frontend`
  (Accounts page, SetupWizard), `domain`
- API compatibility: `POST /api/onboard` keeps working; new dashboard endpoints
  are additive. No breaking changes.
