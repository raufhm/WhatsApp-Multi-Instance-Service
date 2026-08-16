# Developer & AI Agent Guide

**Project**: WhatsApp Multi-Instance Service  
**Repository Layout**: Modular Go backend (`backend/`) + Vite/React dashboard (`frontend/`)

---

## 1. Project Overview

WhatsApp Multi-Instance Service is a multi-tenant conversation platform built on top of [whatsmeow](https://github.com/tulir/whatsmeow). It allows businesses to pair multiple WhatsApp accounts (via QR codes), ingest incoming messages into ticketed conversation threads, run automated deterministic bot rules, and provide a web-based operator dashboard for customer service teams to manage inboxes, review timelines, and send replies.

### Current Implementation Status
- **Multi-Tenant Foundation**: Complete with tenant scoping, operator authentication (password + TOTP), and RBAC (`admin`, `operator`, `viewer`).
- **WhatsApp Integration**: Multi-account linking via QR pairing, event ingestion, and outbound message queuing.
- **Conversation & Ticketing Engine**: Contact normalization, sequential ticket numbering (`TKT-NNNN`), lifecycle states (`OPEN`, `BOT_ACTIVE`, `WAITING`, `HANDED_OFF`, `CLOSED`), and follow-up activities.
- **Bot Engine**: Deterministic rules engine with auditing and runtime rule management.
- **Operator Dashboard (Web UI)**: React + Vite + TypeScript application in `frontend/`, embedded into the Go binary at compile-time via `backend/dist`.
- **Reliable Media**: Asynchronous S3 upload worker with durable job tracking, exponential backoff, and retry jitter.

---

## 2. Repository Layout & Architecture

### Structure
```
.
├── backend/                    # Go service & database migrations
│   ├── main.go                 # Entry point, router setup, lifecycle init
│   ├── dashboard_embed.go      # Embeds backend/dist via //go:embed all:dist
│   ├── config/                 # Environment configuration loader
│   ├── domain/                 # Domain entities, interfaces, and DTOs
│   ├── handler/                # HTTP route handlers (API v1 & Dashboard API)
│   ├── internal/
│   │   ├── bot/                # Deterministic bot rules engine
│   │   ├── conversation/       # Contact/conversation normalizers
│   │   ├── storage/            # PostgreSQL & S3 repositories
│   │   ├── totp/               # TOTP verification & enrollment
│   │   └── upload/             # S3 media upload worker & claim manager
│   ├── migrations/             # golang-migrate SQL migrations
│   ├── test/                   # Subsystem & integration tests
│   ├── go.mod                  # Go module definition
│   └── go.sum
├── frontend/                   # React + Vite + TypeScript SPA
│   ├── src/                    # UI components, pages, hooks, api clients
│   ├── package.json
│   └── vite.config.ts          # Builds assets to ../backend/dist
├── openspec/                   # Specifications, architecture proposals & archives
├── docker-compose.yml          # Local multi-container development environment
├── Dockerfile                  # Multi-stage container build
├── entrypoint.sh               # Container startup script (runs migrations)
├── AGENT.md                    # Agent & contributor documentation
└── README.md                   # Project overview & quickstart
```

### Event Flow & Request Lifecycle
1. **WhatsApp Ingestion**: `whatsmeow` events receive fan-out handling via `Dispatcher` → raw persistence in PostgreSQL and asynchronous projection via `AsyncProjector`.
2. **Conversation Projection**: Contacts and conversations are created or updated; incoming messages are added to the conversation timeline.
3. **Bot Rule Processing**: The deterministic bot evaluates rules against active conversation sessions and queues automated replies if matched.
4. **Outbound Queue & Media**: Outbound messages are dispatched to WhatsApp; media files create durable `upload_jobs` processed by the background upload worker.
5. **Operator Web Dashboard**: Embedded at `/dashboard/` in the Go server, communicating with `/dashboard/api/*` and `/api/v1/*`.

### Authentication & Authorization
- **API v1**: Tenant API Key authentication via `X-API-Key` header or `Authorization: Bearer <key>`.
- **Dashboard API**: Session cookie or JWT authentication, tenant identification via `X-Tenant` header, and operator RBAC (`admin`, `operator`, `viewer`).

---

## 3. Development Workflow

### Prerequisites
- Go ≥ 1.25
- Node.js ≥ 20 / npm
- PostgreSQL ≥ 14
- Docker & Docker Compose (optional, recommended)

### Building and Testing

#### Backend
```sh
# Build the backend
cd backend
go build ./...

# Run all tests
go test ./...
```

#### Frontend
```sh
# Install dependencies
cd frontend
npm install

# Build production assets (outputs directly to backend/dist)
npm run build

# Start dev server with hot-reloading (proxies API to localhost:8080)
npm run dev
```

#### Docker Compose
```sh
# Build and run PostgreSQL + backend with embedded dashboard
docker compose up --build

# Run in detached mode
docker compose up -d --build
```

### Development Notes & Conventions
- **Frontend Embedding**: `backend/dashboard_embed.go` uses `//go:embed all:dist`. A `.gitkeep` is maintained in `backend/dist/` so Go builds succeed even if frontend assets have not been built yet.
- **Database Migrations**: SQL migration files live in `backend/migrations/` and follow the `000N_<name>.up.sql` / `000N_<name>.down.sql` naming convention. When running under Docker, `entrypoint.sh` runs `migrate -path ./migrations` automatically.
- **Git Hygiene**: `backend/dist/*` (except `.gitkeep`), `node_modules/`, local `.env`, and AI harness directories (`.junie/`, `.claude/`, `.cursor/`, etc.) are ignored. All documentation and specifications in `openspec/**/*.md` must remain version-controlled.

---

## 4. Known Roadmap & Gaps

- **Webhooks**: Configurable webhooks with payload signing, delivery retries, and delivery logging.
- **Meta Cloud API**: Alternative transport adapter behind existing transport boundaries.
- **Streaming Media**: Streaming large media uploads directly rather than in-memory buffering.
- **Dead-Letter & Replay**: Durable dead-letter queues and CLI/API replay tools for projection failures.
- **API-Key Lifecycle**: Admin endpoints for creating, rotating, revoking, and tracking API keys.
