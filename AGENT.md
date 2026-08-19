# Developer & AI Agent Guide

**Project**: WhatsApp Multi-Instance Service  
**Repository Layout**: Modular Go backend (`backend/`) + Vite/React dashboard (`frontend/`)

---

## 1. Project Overview

WhatsApp Multi-Instance Service is a multi-tenant conversation platform built on top of [whatsmeow](https://github.com/tulir/whatsmeow). It pairs multiple WhatsApp accounts (via QR codes), ingests messages into ticketed conversation threads, executes deterministic bot rules, and provides an operator dashboard to manage inboxes, timelines, contacts, and deal stages.

### Key Capabilities
- **Multi-Tenant & RBAC**: Tenant isolation, operator auth (password + TOTP), roles (`admin`, `operator`, `viewer`).
- **WhatsApp Gateway**: Multi-account QR pairing, event ingestion, and outbound message queuing.
- **Conversation & Contacts**: Contact normalization (`@s.whatsapp.net` vs `@g.us`), sequential ticketing (`TKT-NNNN`), lifecycle states (`OPEN`, `BOT_ACTIVE`, `WAITING`, `HANDED_OFF`, `CLOSED`), pipeline/deal stages.
- **Bot Engine**: Deterministic rules engine with auditing and runtime rule management.
- **Operator Dashboard**: React + Vite + TypeScript SPA in `frontend/`, embedded into Go binary via `backend/dist`.
- **Media Pipeline**: Asynchronous S3 upload worker with durable job tracking and exponential backoff.

---

## 2. Fast Architecture & Key File Map

Use this map to navigate directly to the relevant files without broad exploratory scans:

| Component / Feature | Key Paths & Files |
|---|---|
| **Entry & Router** | `backend/main.go`, `backend/dashboard_embed.go` |
| **HTTP Handlers & DTOs** | `backend/handler/http.go`, `backend/handler/*_dto.go`, `backend/handler/auth_handlers.go` |
| **Domain Models & Interfaces** | `backend/domain/models.go` |
| **Database & Repositories** | `backend/internal/storage/postgres.go`, `backend/internal/storage/queries.go` |
| **SQL Migrations** | `backend/migrations/000N_*.up.sql` |
| **WhatsApp Subsystem** | `backend/whatsapp/subsystem.go`, `backend/whatsapp/manager.go` |
| **Bot Rules Engine** | `backend/internal/bot/engine.go`, `backend/internal/bot/rules.go` |
| **Conversation Normalization** | `backend/internal/conversation/normalizer.go`, `backend/internal/conversation/contracts.go` |
| **TOTP & Auth Security** | `backend/internal/totp/totp.go` |
| **Media Upload Worker** | `backend/internal/upload/worker.go` |
| **Frontend Routing & App** | `frontend/src/App.tsx`, `frontend/src/components/Layout.tsx` |
| **Frontend Pages** | `frontend/src/pages/` (`Inbox.tsx`, `Contacts.tsx`, `ContactDetail.tsx`, `Accounts.tsx`, etc.) |
| **Frontend API & Hooks** | `frontend/src/api/`, `frontend/src/hooks/` (`useInbox.ts`, `useContacts.ts`, etc.) |
| **Frontend Types** | `frontend/src/types/index.ts` |
| **Specs & Proposals** | `openspec/specs/`, `openspec/proposals/` |

---

## 3. AI Agent Token Efficiency & Execution Rules

To preserve context window tokens, reduce latency, and maintain maximum accuracy, all AI agents must follow these operational guidelines:

### Context & Exploration
- **Targeted Lookups**: Use the Key File Map above and `grep_search` / `glob_search` with specific paths/patterns instead of listing full directories or performing unbounded searches.
- **Specific Line Ranges**: Open specific line windows around relevant functions/types rather than reading full files.
- **Do Not Re-Read Unchanged Context**: If a file's content was already retrieved in the conversation history and has not changed, reuse that knowledge directly.
- **Avoid Large File Dumps**: Never dump whole files or massive logs into the conversation context.

### Precise Edits & Code Style
- **Minimal, Atomic Diffs**: Make surgical search-and-replace edits. Do not rewrite entire files or introduce unnecessary formatting changes.
- **Follow Local Conventions**: Match existing error handling (`fmt.Errorf("...: %w", err)`), naming patterns, and type definitions.
- **Preserve Embeds & Builds**: Never delete `backend/dist/.gitkeep`. Ensure Go builds succeed even before frontend build output is present.

### Scoped Verification & Testing
- **Run Scoped Tests First**:
  - Backend: Run only the affected package, e.g. `go test ./internal/storage -run TestSpecificName` or `go test ./backend/handler/...`.
  - Frontend: Run targeted tests, e.g. `npm test -- src/pages/Contacts.test.tsx --run`.
- **Run Full Suites Only for Final Validation**: Execute `go test ./...` and `npm test -- --run` only when finalizing changes.
- **Suppress Verbose Output**: Avoid `-v` flags unless debugging a specific test failure to avoid flooding the context with log lines.

### Response & Communication Economy
- **Direct & Concise**: Avoid boilerplate summaries, conversational filler, or repeating entire source code blocks in explanations.
- **Reference by Path and Line**: Point directly to `path/to/file.go:line` rather than quoting long snippets.

---

## 4. Development Workflow & Commands

### Prerequisites
- Go ≥ 1.25, Node.js ≥ 20 / npm, PostgreSQL ≥ 14, Docker (optional)

### Build & Run Commands

```sh
# Backend build & test
cd backend
go test ./...                    # All backend tests
go test ./internal/storage -v    # Scoped package test

# Frontend build & test
cd frontend
npm run build                    # Build SPA to ../backend/dist
npm test -- --run                # Run all frontend tests in CI mode
npm test -- src/pages/Contacts.test.tsx --run # Scoped test

# Docker Compose (local full-stack)
docker compose up -d --build
```

### Database Migrations
- Migration files are in `backend/migrations/` (`000N_<name>.up.sql` / `000N_<name>.down.sql`).
- Apply migrations using `migrate -path ./migrations -database "$DATABASE_URL" up`.

---

## 5. Known Roadmap & Future Scope
- **Webhooks**: Configurable webhooks with payload signing and delivery retries.
- **Meta Cloud API**: Alternative transport adapter behind existing transport interfaces.
- **Streaming Media**: Direct S3 streaming for large media payloads.
- **Dead-Letter & Replay**: Durable queues and CLI replay tools for projection failures.
- **API-Key Lifecycle**: Admin endpoints for key rotation and revocation.
