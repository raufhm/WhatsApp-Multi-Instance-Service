# Developer & AI Agent Guide

This document is a practical orientation for contributors and agents. It is not a product spec, and it should be treated as a working guide that can change as the repository evolves.

## Project Snapshot

WhatsApp Multi-Instance Service is a self-hosted WhatsApp inbox and operator dashboard. The codebase uses a Go backend, a React/Vite frontend, PostgreSQL for persistence, and `whatsmeow` for WhatsApp connectivity.

What the project currently appears to focus on:

- connecting WhatsApp accounts through QR pairing
- showing conversations in a shared dashboard
- sending messages and media
- handling notes, follow-ups, and conversation state
- basic contact and deal-stage tracking
- optional media storage through S3-compatible storage

This repository does include tenant-aware storage and operator authentication flows, but avoid assuming enterprise-grade guarantees unless you verify the specific code path you are working on.

## What To Be Careful About

- Linked-device WhatsApp flows can break if WhatsApp changes behavior or applies account restrictions.
- The product is not the official Meta Cloud API.
- Some workflows are intentionally manual, especially around CRM-style tracking and operator actions.
- Media may fall back to local disk if S3 is not configured, and remote media URLs can expire.
- Any claim about permissions, lifecycle states, or background workers should be verified in code before depending on it.

## Key Files

Use these as the first places to look before broad searches:

| Area | Files |
|---|---|
| Backend entrypoint | `backend/main.go` |
| Dashboard embed | `backend/dashboard_embed.go` |
| Domain models | `backend/domain/models.go` |
| WhatsApp integration | `backend/whatsapp/subsystem.go`, `backend/whatsapp/pairing.go`, `backend/whatsapp/templates.go` |
| Bot logic | `backend/internal/bot/engine.go` |
| Conversation helpers | `backend/internal/conversation/contracts.go` |
| Storage layer | `backend/internal/storage/*.go` |
| Upload worker | `backend/internal/upload/*.go` |
| Auth / TOTP | `backend/internal/totp/totp.go` |
| Migrations | `backend/migrations/` |
| Frontend app | `frontend/src/App.tsx`, `frontend/src/main.tsx`, `frontend/src/routes/` |
| Frontend pages | `frontend/src/pages/` |
| Shared UI | `frontend/src/components/` |
| Frontend API client | `frontend/src/lib/apiClient.ts` |
| Types | `frontend/src/types/index.ts` |
| OpenSpec docs | `openspec/` |

## How To Work In This Repo

### Exploration

- Prefer targeted searches with `rg` over broad directory scans.
- Open only the files you need for the current task.
- Reuse context that is already visible in the conversation instead of rereading unchanged files.

### Code Changes

- Keep diffs small and focused.
- Match the existing style of the file you are editing.
- Avoid speculative refactors unless the user asked for them.
- Do not remove generated or embedded assets unless you know the build pipeline will regenerate them.

### Verification

- Run scoped tests first when possible.
- Use full-suite checks only when you are finalizing a broader change.
- If a command fails because of sandboxing or missing network access, retry with escalation only if the command is important to the task.

## Common Commands

### Backend

```sh
cd backend
go test ./...
go run .
```

### Frontend

```sh
cd frontend
npm install
npm run build
npm test -- --run
```

### Docker Compose

```sh
docker compose up --build -d
docker compose logs -f app
```

## Configuration Notes

- Copy `.env.example` to `.env` before local runs.
- `PG_DSN` is required for the backend.
- `PORT`, `LOG_LEVEL`, and `BOT_SESSION_TIMEOUT` are the main runtime knobs.
- Leave `S3_BUCKET` empty if you want to use local disk storage for media.

## Scope Notes

If you are editing or describing features, stay aligned with what is visible in the code and README.

Prefer language like:

- “supports”
- “appears to”
- “currently implements”
- “is intended to”

Avoid language like:

- “enterprise-grade”
- “fully featured”
- “guaranteed”
- “production ready” unless the specific change has actually been validated

## Future Work

The repository has room for improvement in areas like:

- onboarding and setup polish
- reliability and reconnect handling
- observability and error reporting
- tests and regression coverage
- documentation and examples
- dashboard UX refinements

Treat those as opportunities, not commitments.
