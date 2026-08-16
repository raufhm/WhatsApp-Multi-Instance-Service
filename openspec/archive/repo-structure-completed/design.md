# Design: Repository tidy-up and structure

## Target layout

```
.
├── backend/                    # Go service
│   ├── main.go
│   ├── dashboard_embed.go      # //go:embed all:dist
│   ├── config/
│   ├── domain/
│   ├── handler/
│   ├── internal/
│   ├── whatsapp/
│   ├── test/
│   ├── migrations/
│   ├── go.mod
│   └── go.sum
├── frontend/                   # Vite + React + TS
│   ├── src/
│   ├── package.json
│   ├── vite.config.ts
│   └── ... (node_modules untracked; dist emitted to backend/dist)
├── .env.example
├── .gitignore
├── .dockerignore
├── docker-compose.yml
├── Dockerfile
├── entrypoint.sh
├── AGENT.md
├── README.md
├── openspec/
└── dist/                       # build output (gitignored), embedded by Go
```

Notes:

- `dist/` is the frontend build output. It is committed only as a `.gitkeep`
  placeholder (so `//go:embed all:dist` compiles when the frontend is unbuilt),
  otherwise gitignored.
- `migrations/` moves under `backend/` next to the Go module so
  `migrate -path ./migrations` and the embed/environment stay cohesive.

## Go module & packages

The module is currently `github.com/raufhm/whatsapp-testing`. Two options:

1. **Rename** to something meaningful (e.g. `github.com/raufhm/whatsapp-multi-instance-service`)
   and rewrite all import paths. More churn via `go.mod` + `sed` across imports.
2. **Keep** the existing module path and simply move the directory. Lowest
   risk; only build context (`-C backend` / `WORKDIR backend`) changes.

Recommended: **keep the module path** for this tidy-up to minimize risk, and
rename in a separate follow-up if desired. The move is purely a directory shift,
not an import-path change (imports are module-path-relative, not filesystem
relative).

## Go embed directive

`dashboard_embed.go` embeds `//go:embed all:dist`. After the move:

- Backend lives in `backend/` and runs with `WORKDIR backend` (Docker) or
  `cd backend && go run .` (local).
- `frontend/vite.config.ts` `build.outDir` must point to `../backend/dist`
  (currently `../dist`).
- The `dist/` placeholder must exist at `backend/dist/.gitkeep` so the embed
  compiles before a frontend build.

## Docker changes

- `Dockerfile`: frontend builder already `COPY frontend/`; keep. Go builder
  `WORKDIR /app/backend`, `COPY backend/go.mod backend/go.sum`, etc., then
  `COPY backend/ .`. Copy built `dist` into `backend/dist`.
- `docker-compose.yml`: `app` build context stays repo root; any volume/mount
  paths referencing `./migrations` or sessions update to `./backend/...` as
  needed. `entrypoint.sh` `-path ./migrations -database "$PG_DSN"` is unchanged
  because it runs from the backend working directory.
- `entrypoint.sh`: `migrate -path ./migrations ...` still resolves because the
  runtime `WORKDIR` is `backend`.

## `.gitignore` rewrite

Replace the current `*.md` block with explicit rules so source markdown is
tracked and only true junk is ignored:

```gitignore
# Build artifacts
/dist
/backend/dist
/frontend/node_modules
*.exe
whatsapp-service
whatsapp-api

# Secrets / env
.env
.env.local
*.pem

# Native/cache
.DS_Store
.idea/
.vscode/
*.log

# AI agent harnesses
.junie/
.claude/
.cursor/
.codex/
.aider/
.aider*
.cursorrules
.codex/
.windsurf/

# Storage / sessions (runtime data)
/storage
/sessions
/backend/sessions
```

Key correction: **remove** the `*.md` ignore. `README.md`, `AGENT.md`, and all
`openspec/**/*.md` must be tracked. Only `dist/`, `node_modules`, secrets,
caches, and runtime storage are ignored.

## `AGENT.md` rewrite

Reframe from a stale implementation checklist to a durable agent/contributor
guide:

- **What the project is** (multi-tenant WhatsApp service on whatsmeow; bot +
  operator dashboard).
- **Current architecture** (packages, event flow, auth: API key + dashboard
  session fallback).
- **Current status** (MVP implemented: signup/signin, QR onboarding, inbox
  text+media).
- **How to build / test / run** (backend `go build ./...`, `go test ./...`;
  frontend `npm run build`; `docker compose up`).
- **How it behaves during development** (embed expects `backend/dist`, Vite dev
  proxy, .env, migrations auto-run in Docker).
- **Known gaps** (webhooks, Cloud API, API-key lifecycle, streaming media,
  dead-letter queue) as a short roadmap.
- Remove dead links (`specs/01-operator-permissions/`, old `changes/` paths).

## Verification

After the move, confirm zero behavior change:

1. `cd backend && go build ./... && go test ./...`
2. `cd frontend && npm run build` (emits to `backend/dist`)
3. `docker compose build` (or a full `docker compose up --build`) succeeds
4. `git status` shows `openspec/**/*.md` now tracked (not ignored)
5. `git check-ignore .junie` and `git check-ignore dist` report ignored
6. `README.md` quick-start and `AGENT.md` reflect the new paths

## Out of scope

- Renaming the Go module path (deferred).
- Reorganizing files *within* packages.
- Removing currently-unused source or dependency cleanup beyond the layout.
