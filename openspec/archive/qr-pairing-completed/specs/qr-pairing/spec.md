# QR pairing

## ADDED Requirements

### Requirement: Operators can start pairing and receive a scannable QR code

An authenticated operator with account-management permission MUST be able to
start a pairing session and receive a QR code over HTTP (not server logs).

#### Scenario: Start pairing

- **WHEN** an authorized operator starts pairing from the dashboard
- **THEN** the service MUST create a pairing session and return a QR code image
- **AND** the QR MUST be renderable in the browser

### Requirement: Pairing state is pollable and QR refreshes on expiry

The pairing session MUST expose its status, and MUST issue a fresh QR code when
the current one expires.

#### Scenario: QR expires

- **WHEN** a QR code expires before scanning
- **THEN** the service MUST emit a new QR code for the same session
- **AND** polling MUST reflect the updated QR and status

### Requirement: A successful scan registers the account and connects the instance

On a successful scan, the service MUST persist the number into
`whatsapp_accounts` scoped to the operator's tenant, and MUST spawn the WhatsApp
instance.

#### Scenario: Scan succeeds

- **WHEN** the operator scans the QR in WhatsApp
- **THEN** the service MUST upsert a `whatsapp_accounts` row for the tenant and
  host phone
- **AND** the instance MUST be spawned and reachable via `ListInstances` and
  `ListAccounts`
- **AND** subsequent messages from that host MUST be projected (tenant boundary
  resolves)

### Requirement: Pairing is permission-gated and cancellable

Only operators with the manage-accounts permission MAY start or cancel pairing,
and a pending session MUST be cancellable.

#### Scenario: Unauthorized pairing

- **WHEN** an operator without manage-accounts permission attempts to start
  pairing
- **THEN** the service MUST return 403

#### Scenario: Cancel pairing

- **WHEN** an operator cancels a pending pairing session
- **THEN** the session MUST be terminated and its temporary client disconnected
