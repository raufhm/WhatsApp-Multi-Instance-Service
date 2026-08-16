# Change: Dashboard inbox & messaging access

## Why

The WhatsApp pipeline works end to end at the library level: after QR onboarding
(`qr-pairing`), the host is registered in `whatsapp_accounts`, and incoming and
outgoing messages are projected into `conversations` / `conversation_messages`
via the `AsyncProjector`.

But an operator **cannot see or send them** from the dashboard, because every
inbox operation is served by `/api/v1/*`, which is authenticated by API key
only (`Server.tenant()` checks `X-API-Key`/`Bearer`), and:

1. No API key is ever minted — there is no create/list/revoke endpoint and no
   code writes to `api_keys` (the "API Key Lifecycle" gap is still open).
2. The dashboard authenticates with a session cookie scoped to `Path=/dashboard`,
   so it is not sent to `/api/v1/*`.

Net effect: the inbox returns `401`, and the MVP (sign up → sign in → onboard →
send/receive text & media) is blocked at the inbox boundary.

## What Changes

- Allow the existing `/api/v1/*` inbox endpoints to authenticate a logged-in
  operator's dashboard session as an alternative to an API key.
- Broaden the dashboard session cookie scope so the browser sends it to the
  inbox endpoints.
- Derive the operator identity for audit/actor fields from the session when
  authenticated that way.
- Keep API-key auth unchanged for external/programmatic clients.
- Add coverage proving both auth modes work and that tenant scoping still holds.

## Impact

- Affected specs: `dashboard-inbox-access`
- Affected code: `handler` (session-aware tenant resolution), `handler/dashboard.go`
  (cookie scope), `main.go` (wire auth into the API server), `frontend`
  (no functional change expected), and tests
- API compatibility: `/api/v1/*` continues to accept API keys; session-based
  access is additive. No breaking changes.
