# Archived: Repository tidy-up and structure

**Archived**: August 16, 2026
**Status**: ✅ FULLY IMPLEMENTED

This change has been fully implemented, built, and tested. The specification is
preserved here for historical reference.

## What Was Implemented

Reorganized the repo into a clean `backend/` + `frontend/` split, fixed broken
git hygiene, and rewrote `AGENT.md` as an accurate developer/agent guide.

### Layout
- Moved all Go code into `backend/`: `main.go`, `dashboard_embed.go`, `config/`,
  `domain/`, `handler/`, `internal/`, `whatsapp/`, `test/`, `migrations/`,
  `go.mod`, `go.sum`.
- Kept the Go module path unchanged (`github.com/raufhm/whatsapp-testing`), so
  imports stay valid across the move.
- `frontend/` (Vite + React + TS) stays; `vite.config.ts` `build.outDir` →
  `../backend/dist`.
- `backend/dist/` is the frontend build output, embedded by
  `//go:embed all:dist`; `.gitkeep` placeholder documented.

### Git hygiene
- Rewrote `.gitignore`: removed the `*.md` blanket rule so `README.md`,
  `AGENT.md`, and `openspec/**/*.md` are tracked.
- Added AI-harness ignores: `.junie/`, `.claude/`, `.cursor/`, `.codex/`,
  `.aider/`, `.aider*`, `.cursorrules`, `.windsurf/`.
- Ignored build artifacts (`backend/dist/*`, `node_modules/`), secrets (`.env`,
  `*.pem`), caches, and runtime storage.
- Updated `.dockerignore` to match.

### Docker
- `Dockerfile` builds the Go module from `WORKDIR /app/backend`.
- `COPY --from=frontend-builder /app/backend/dist` correctly captures the build
  (frontend `outDir` is `../backend/dist` within its stage).

### Documentation
- Rewrote `AGENT.md` as "Developer & AI Agent Guide": purpose, architecture,
  dev workflow (build/test/run), authentication model, and known roadmap gaps.
- Removed stale links (`specs/01-operator-permissions/`, old `changes/` paths).
- Updated `README.md` for the new paths.

## Acceptance Criteria Met

- `cd backend && go build ./...` and `go test ./...` pass ✅
- `cd frontend && npm run build` emits to `backend/dist` ✅
- `docker build` succeeds ✅
- `git check-ignore` confirms `.junie/`, `backend/dist/*`, `.env` ignored and
  openspec `.md` tracked ✅
- `AGENT.md` reflects the current layout and workflow ✅

## Notes

- `backend/dist/.gitkeep` is documented in `AGENT.md` but was not present in the
  working tree at archive time (the build currently populates `dist/` with real
  assets). Recreate it (`touch backend/dist/.gitkeep`) if `dist/` is ever
  emptied so `//go:embed all:dist` keeps compiling.
- The `docker-compose.yml` `version` key still emits an "obsolete" warning;
  cosmetic only.

## Related

- `../agent-implementation-plan/` — remaining feature roadmap
- `../dashboard-inbox-access-completed/` — inbox auth bridge (archived)
- `../qr-pairing-completed/` — onboarding (archived)
