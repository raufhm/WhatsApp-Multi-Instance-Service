# WhatsApp Multi-Instance Service

An API-first conversation platform built on top of [whatsmeow](https://github.com/tulir/whatsmeow).  
Link personal/business WhatsApp numbers to the service; every inbound and outbound message is stored against a contact and ticketed conversation; a deterministic bot handles initial exchanges; completed sessions produce acknowledgeable follow-up activities for customer service teams.

> **Disclaimer:** Personal portfolio project built for research and development. Using linked WhatsApp accounts (QR-paired) may violate WhatsApp's Terms of Service. Run at your own risk.

---

## Architecture overview

```
whatsmeow event → Dispatcher fan-out
    ├── PostgresStore (raw whatsmeow_* persistence)
    └── AsyncProjector
            ├── ProjectMessageContext (contact + conversation + timeline)
            └── bot.Processor (deterministic rules → WhatsApp queue → BOT timeline entry)
                        └── on terminal/handoff → CloseConversation → Activity (PENDING)

Outgoing media  → durable upload_jobs row (object key generated once)
Upload worker   → claims due jobs (FOR UPDATE SKIP LOCKED), S3 upload,
                  retries with bounded exponential backoff + jitter,
                  marks COMPLETED/FAILED and attaches the archive URL to the message

HTTP API (/api/v1/*)  ← API-key authenticated, tenant-scoped
Background ticker    → CloseAllTimedOut (every 1 min)
```

### Key packages

| Package | Purpose |
|---|---|
| `backend/whatsapp/` | Multi-instance transport, outbound queue, `AsyncProjector` |
| `backend/internal/bot/` | Deterministic rules engine + per-conversation processor |
| `backend/internal/storage/` | PostgreSQL repositories (raw + application layer) |
| `backend/internal/conversation/` | Address normalisation utilities |
| `backend/domain/` | Shared types, interfaces, enums |
| `backend/handler/` | HTTP server (legacy + `/api/v1` platform API) |
| `backend/config/` | Environment-variable configuration |
| `backend/migrations/` | golang-migrate SQL migrations |

---

## Lifecycle model

```
OPEN → BOT_ACTIVE → WAITING → HANDED_OFF
                            └→ CLOSED
```

Every conversation receives a unique, sequentially-allocated **ticket number** (`TKT-NNNN`).  
A single **PENDING activity** is created when a session closes (terminal rule, handoff, or timeout); once acknowledged by an operator the activity transitions to `ACKNOWLEDGED`.

## Reliable media uploads

Outgoing media is archived to S3 asynchronously through durable `upload_jobs`. On the send path a job is persisted with a **single object key generated once**; the upload worker claims due jobs atomically (`FOR UPDATE SKIP LOCKED`), uploads, and:

- **success** → `COMPLETED`, the archive URL is persisted and attached to the message;
- **transient** S3/network error → re-queued (`PENDING`) with exponential backoff + jitter;
- **permanent** error (e.g. missing config) or hitting `UPLOAD_MAX_ATTEMPTS` → `FAILED`, retaining the last error and attempt count for diagnosis.

Because the object key is reused for every attempt, a crash after S3 accepts an upload but before completion is persisted converges to a single archive object on the next attempt. The message is never treated as having an available archive URL until S3 confirms. The worker logs `job`, `message`, `key`, `attempt`, and outcome (never media bytes); the poll loop uses `UPLOAD_*` settings and shuts down gracefully on `SIGINT`/`SIGTERM`.

---

## Getting started

### Prerequisites

- Go ≥ 1.25
- PostgreSQL ≥ 14
- AWS S3 bucket (optional — only required for media uploads)

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `PG_DSN` | *(required)* | PostgreSQL connection string |
| `PORT` | `8080` | HTTP listen port |
| `S3_BUCKET` | `""` | S3 bucket name for outgoing media |
| `BOT_SESSION_TIMEOUT` | `30m` | Idle session timeout before auto-closure |
| `BOT_FALLBACK_REPLY` | *(generic message)* | Reply when no rule matches |
| `BOT_RULES_VERSION` | `default` | Version tag stored in `bot_sessions` for audit |
| `LOG_LEVEL` | `INFO` | Application log verbosity |
| `UPLOAD_WORKER_ENABLED` | `true` | Enable the retryable S3 upload worker |
| `UPLOAD_POLL_INTERVAL` | `5s` | Upload worker poll interval |
| `UPLOAD_MAX_ATTEMPTS` | `5` | Max upload attempts before permanent failure |
| `UPLOAD_INITIAL_BACKOFF` | `1s` | Initial retry backoff |
| `UPLOAD_MAX_BACKOFF` | `60s` | Retry backoff cap |
| `UPLOAD_LEASE` | `60s` | Claim lease; expired leases are reclaimed after a crash |
| `UPLOAD_JITTER` | `0.2` | Random backoff jitter factor (0 disables) |

### Run locally

```sh
# 1. Apply migrations
migrate -path ./backend/migrations -database "$PG_DSN" up

# 2. Set environment and start
export PG_DSN="postgres://user:pass@localhost:5432/whatsapp?sslmode=disable"
export PORT=8080
cd backend && go run .

# 3. Pair a WhatsApp account
# Either via the Operator Dashboard (http://localhost:8080/dashboard -> Accounts -> Link Device)
# or via the API:
curl -X POST http://localhost:8080/api/onboard -d '{"email":"you@example.com"}'
# → scan the QR code printed to stdout / base64 in logs / UI pairing modal
```

### Run the operator dashboard (frontend)

The dashboard is a Vite + React + TypeScript app in `frontend/`. It is served by
Go from the embedded `backend/dist/` build at `/dashboard/`.

```sh
# Build the frontend (produces backend/dist, embedded by the Go binary)
cd frontend && npm install && npm run build && cd ..

# Start the backend, then open http://localhost:8080/dashboard
cd backend && go run .
# (if the frontend isn't built, the server still runs API-only)
```

For frontend development with hot reload, run the Vite dev server, which proxies
`/api` and `/dashboard/api` to the Go backend on `:8080`:

```sh
cd frontend && npm run dev
# open http://localhost:5173/dashboard
```

To log in you first need an operator row (see `backend/migrations/0005_operator_dashboard.up.sql`)
and an `X-Tenant` header pointing at the owning tenant. The login endpoint is
`POST /dashboard/api/login` with `{ "email": ..., "password": ... }`.

### Docker Compose (recommended for local development)

A `docker-compose.yml` is included that spins up PostgreSQL and the service in one
command. The frontend is built into the image and migrations run automatically.

```sh
# Copy the example environment
cp .env.example .env

# Start everything
docker compose up --build -d

# View logs
docker compose logs -f app

# Open dashboard
open http://localhost:8080/dashboard

# Stop
docker compose down
```

The compose file mounts a persistent volume for PostgreSQL and a separate volume
for WhatsApp session state. Update `.env` to configure S3 credentials if you want
to enable media uploads.

### Docker image only

```sh
# Build and run the image manually
docker build -t whatsapp-service .
docker run -e PG_DSN="postgres://..." -p 8080:8080 whatsapp-service
```

### Legacy Docker snippet

```sh
docker build -t whatsapp-service .
docker run -e PG_DSN="..." -p 8080:8080 whatsapp-service
```

All `/api/v1/*` endpoints require a tenant API key, passed as:
- Header: `X-API-Key: <key>`
- Bearer: `Authorization: Bearer <key>`

Pagination: `?limit=50&offset=0` (limit 1–100).

### Accounts

```
GET  /api/v1/accounts               → list linked WhatsApp accounts
POST /api/v1/accounts/{account}/messages  → enqueue outbound message (202 Queued)
```

### Contacts

```
GET  /api/v1/contacts               → list contacts (paginated)
GET  /api/v1/contacts/{id}          → contact detail
```

### Conversations / Tickets

```
GET  /api/v1/conversations          → inbox (optional ?status=OPEN|BOT_ACTIVE|WAITING|HANDED_OFF|CLOSED)
GET  /api/v1/conversations/{id}     → conversation + message timeline (by UUID or ticket number)
GET  /api/v1/conversations/{id}/messages  → same as above
GET  /api/v1/tickets                → alias for /api/v1/conversations
```

### Activities

```
GET  /api/v1/activities             → follow-up queue (optional ?status=PENDING|ACKNOWLEDGED|DISMISSED)
POST /api/v1/activities/{id}/acknowledge  → idempotent acknowledge (X-Actor header optional)
```

### Bot rules

```
GET  /api/v1/bot-rules                    → list all ruleset versions
POST /api/v1/bot-rules                    → create a new ruleset version (inactive)
GET  /api/v1/bot-rules/active             → currently active ruleset
POST /api/v1/bot-rules/activate?version=N → activate a version (deactivates previous)
```

Rule JSON body: `{ "rules": [{ "name": "...", "pattern": "...", "match": "CONTAINS|EXACT|PREFIX", "response": "...", "terminal": false, "handoff": false }] }`

### Operator actions

All endpoints accept an optional `X-Actor` header (defaults to `api`).

```
POST /api/v1/operator/assign?id=<conv>   → body {"assignee":"name"} assigns operator
POST /api/v1/operator/handoff?id=<conv>   → body {"reason":"..."} hands off to human
POST /api/v1/operator/close?id=<conv>     → body {"reason":"..."} closes with reason
POST /api/v1/operator/reopen?id=<conv>   → reopens a closed conversation
```

### Internal notes

```
POST /api/v1/notes?id=<conv>              → body {"content":"note", "actor":"OPERATOR"} adds internal note
```

Notes use `is_internal=true` and are never sent to WhatsApp contacts.

### Conversation merge / split

```
POST /api/v1/merge                        → body {"source_id":"<uuid>", "target_id":"<uuid>"} merges source into target
POST /api/v1/split?id=<source>             → body {"message_ids":["uuid",...]} splits messages into new conversation
GET  /api/v1/audit-logs?limit=&offset=   → operator audit log (newest first)
```

Merge moves all messages from source to target, closes source with `closure_reason='merged'`. Split creates a new conversation from the same account/contact and moves the specified messages.

### Dashboard

The operator dashboard is served at `/dashboard/` and consumes the `/api/v1/*` API.

```
GET  /dashboard/inbox          → conversation inbox
GET  /dashboard/conversations/:id → conversation detail + reply composer
GET  /dashboard/contacts       → contact directory
GET  /dashboard/accounts       → WhatsApp account health
GET  /dashboard/bot-rules       → bot rules versions + editor
GET  /dashboard/upload-jobs    → outgoing media upload status/failures
```

### Legacy endpoints (backwards-compatible)

```
POST /api/onboard       → pair a new WhatsApp account via QR
POST /api/send          → raw send (no tenant auth)
GET  /api/bots          → list all paired instances
GET  /api/bots/detail?host=  → instance detail
GET  /api/health        → liveness probe
```

### Error format

```json
{ "error": "human-readable message", "code": "MACHINE_READABLE_CODE" }
```

---

## Multi-tenancy

Each API key is hashed (SHA-256) and stored in `api_keys.key_hash` — raw keys are never persisted.  
A tenant is created by inserting a row into `tenants`; API keys and linked WhatsApp accounts then belong to that tenant.  
Every query is scoped by `tenant_id`; cross-tenant access is structurally impossible at the repository layer.

---

## Known limitations / risks

- **WhatsApp policy risk**: linking personal accounts violates WhatsApp ToS; use this for research only.
- **No Cloud API support** (MVP scope): Meta Business API integration is deferred.
- **Dashboard is API-first**: a React operator dashboard is served at `/dashboard/`, but some management screens (accounts, bot rules) are still pending.
- **Large media**: outgoing media is read fully into RAM before upload; streaming is deferred.
- **No durable work queue**: the bounded async projector drops events when full under sustained load.
- **`Processor.locks` map** grows unbounded for long-lived processes; future: add eviction.

---

## Roadmap

- [x] Configurable bot rules loaded from database / file at runtime
- [x] Operator dashboard (auth, inbox, conversation detail, ticket actions, activity queue)
- [ ] Webhook dispatch for inbound events
- [ ] Human agent hand-off UI
- [ ] Meta Cloud API transport adapter
- [ ] AI/LLM-based transcript summarisation
- [x] Reliable outgoing-media S3 uploads (durable jobs, bounded retries, idempotent keys)
- [ ] Streaming large-media upload (defer RAM loading)
- [ ] Retry/dead-letter queue for failed projections
