# Tasks

## Phase 1 — Create backend directory and move Go code

- [x] Create `backend/` and move `main.go`, `dashboard_embed.go`, `config/`,
  `domain/`, `handler/`, `internal/`, `whatsapp/`, `test/`, `migrations/`,
  `go.mod`, `go.sum` into it.
- [x] Keep the Go module path unchanged (imports stay module-path-relative).
- [x] Update `dashboard_embed.go` comment if it references paths.

## Phase 2 — Frontend output path

- [x] Update `frontend/vite.config.ts` `build.outDir` from `../dist` to
  `../backend/dist`.
- [x] Add `backend/dist/.gitkeep` so `//go:embed all:dist` compiles when the
  frontend is unbuilt.

## Phase 3 — Docker & entrypoint

- [x] Update `Dockerfile` to build the Go module from `backend/`
  (`WORKDIR /app/backend`, copy `backend/go.mod sum`, `COPY backend/ .`).
- [x] Copy the frontend-built `dist` into `backend/dist`.
- [x] Update `docker-compose.yml` volumes/context for the new layout.
- [x] Confirm `entrypoint.sh` migrations `-path ./migrations` resolves from the
  backend working dir.

## Phase 4 — Git hygiene

- [x] Rewrite `.gitignore`: remove `*.md`; ignore `dist`, `node_modules`,
  secrets, caches, runtime storage, and AI-harness dirs (`.junie/`, `.claude/`,
  `.cursor/`, `.codex/`, `.aider*`, `.windsurf/`, etc.).
- [x] Update `.dockerignore` to match (backend/dist, node_modules, AI harnesses).
- [x] Ensure `openspec/**/*.md`, `README.md`, `AGENT.md` are tracked again.

## Phase 5 — AGENT.md and README

- [x] Rewrite `AGENT.md`: purpose, architecture, status, build/test/run, dev
  workflow, known gaps; remove dead links.
- [x] Update `README.md` paths (module under `backend/`, frontend build output).
- [x] Update any other docs referencing root-level Go paths.

## Phase 6 — Verification

- [x] `cd backend && go build ./...` and `go test ./...`.
- [x] `cd frontend && npm run build` emits to `backend/dist`.
- [x] `docker compose build` (or full up) succeeds.
- [x] `git check-ignore` confirms dist/node_modules/.junie ignored and openspec md tracked.
- [x] Manual smoke: sign up → onboarding → inbox send/receive still works.
