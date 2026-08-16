# Design: Scan-to-connect WhatsApp pairing

## State machine

A pairing session moves through a small, pollable lifecycle:

```text
AWAITING_SCAN ──scan success──▶ CONNECTED
      │  ▲
      │  └── QR timeout (auto-regenerate)
      ├── cancel ──▶ CANCELLED
      └── error ───▶ FAILED
```

Each session stores:

- `id` (uuid)
- `tenant_id`
- `status`
- `qr_data_url` (base64 PNG) or empty once connected
- `host_phone` (resolved after success)
- `error` (last error message)
- timestamps and the underlying `*whatsmeow.Client` + cancel func

## Pairing manager

A new `whatsapp.PairingManager` owns sessions in a mutex-protected map. It is
single-process and intentionally in-memory: a pairing is ephemeral and tied to a
live `whatsmeow.Client`. (Multi-instance deployments behind a load balancer are
out of scope for v1; noted below.)

```go
type PairingManager struct {
    sessions map[string]*PairingSession
    // + mu, manager *WhatsAppManager, store AccountRegistrar
}

func (p *PairingManager) Start(tenantID uuid.UUID, displayName string) (string, error)
func (p *PairingManager) Get(id string) (PairingSnapshot, bool)
func (p *PairingManager) Cancel(id string) error
```

`Start` creates a device via the existing `Container.NewDevice()`, builds a
`whatsmeow.Client`, opens the QR channel, and launches a goroutine that reads
channel events:

- `code` → encode PNG → base64 data URL, set `AWAITING_SCAN`
- `success` → resolve `HostPhone` from `client.Store.ID`, register the account,
  `SpawnInstance(device)`, set `CONNECTED`
- `timeout` → clear `qr_data_url`, set a brief `EXPIRED`/refreshing state so the
  client can emit the next `code`
- error → set `FAILED`

On success the session is retained for a short TTL so the poller can read the
final `CONNECTED` snapshot, then is reaped.

## Account registration (the missing link)

Add to the repository boundary:

```go
RegisterAccount(tenantID uuid.UUID, hostID, displayName, provider string) (domain.WhatsAppAccount, error)
```

Implementation is an idempotent upsert into `whatsapp_accounts`:

```sql
INSERT INTO whatsapp_accounts (tenant_id, host_id, provider, display_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tenant_id, provider, host_id)
DO UPDATE SET display_name = EXCLUDED.display_name, updated_at = CURRENT_TIMESTAMP
RETURNING id, tenant_id, host_id, provider, display_name, created_at, updated_at;
```

This is what makes a paired number actually visible in `ListAccounts` and,
critically, lets `ProjectMessage` resolve the tenant boundary for its events.

## API endpoints (dashboard, session + permission)

Reuse the existing `PermManageAccounts` permission and `DashboardSessionMiddleware`.

- `POST /dashboard/api/pairing` — body `{ display_name }`; returns `{ pairing_id }`
- `GET /dashboard/api/pairing/{id}` — returns `{ status, qr_data_url, host_phone, error }`
- `POST /dashboard/api/pairing/{id}/cancel` — cancels and disconnects the client

Repurpose the current `startPairing` (`POST /dashboard/api/accounts`) to start a
pairing session (or alias it to `POST /dashboard/api/pairing`).

## QR rendering

Move the existing PNG/base64 logic out of `OnboardHandler` into a helper, e.g.
`whatsapp.EncodeQRDataURL(code string) (string, error)` using `qrcode.Encode`
with `qrcode.Medium, 256`. The frontend renders it as `<img src={qr_data_url}>`.

## Frontend

Accounts page gains an "Add account" button that opens a pairing modal:

- On open, `POST /dashboard/api/pairing`.
- Poll `GET /dashboard/api/pairing/{id}` every ~2s.
- Render the QR image, a "Scan with WhatsApp → Linked Devices → Link a Device"
  hint, and a status line (awaiting scan / connected / failed).
- On `CONNECTED`, invalidate the accounts query, close the modal, and show the
  new account.
- A "Cancel" button calls the cancel endpoint and closes the modal.

Optionally, replace the `SetupWizard` fake toggle with a link to the real
pairing flow.

## Cleanup and safety

- Cancel and error paths disconnect the temporary client.
- Sessions carry a TTL and are reaped by the poll loop or a small janitor.
- Only one active pairing session per tenant is allowed by default (starting a
  new one cancels the previous) to avoid orphaned QR clients.

## Out of scope (v1)

- Persistent pairing sessions across process restarts / multi-instance workers
- SSE/WebSocket push (polling matches the existing dashboard refresh pattern)
- WhatsApp "pairing code" (phone-number + 8-char code) flow
