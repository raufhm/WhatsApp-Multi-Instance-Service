# Change: Repository tidy-up and structure

## Why

The repository has outgrown its initial flat layout and its tooling config has
drifted out of step with how the code is actually maintained:

1. The Go backend is at the repo root (`main.go`, `dashboard_embed.go`,
   `config/`, `domain/`, `handler/`, `internal/`, `whatsapp/`, `migrations/`)
   while the frontend lives in `frontend/`. This makes the backend/frontend
   boundary hard to see and the root noisy.
2. `.gitignore` has `*.md` (with only `README.md` and `AGENT.md` re-allowed),
   which silently **ignores every `openspec/**/*.md`** — the spec changes and
   archive are not version-controlled. It also fails to ignore AI-harness
   directories (`.junie/`) and build output details.
3. `AGENT.md` is stale: it still claims the dashboard is "not implemented yet",
   points at a `specs/01-operator-permissions/` path that no longer exists, and
   references the old `changes/` layout which has since been archived.
4. `dist/` and `/frontend/node_modules` are build artifacts that belong only in
   `.gitignore` and `.dockerignore`.

## What Changes

- Move all Go backend code under `backend/` and all frontend code under
  `frontend/`, with the build/embed/entrypoint/paths updated to match.
- Fix `.gitignore` so source `.md` files (README/AGENT/openspec) are tracked but
  AI-harness dirs, secrets, and build artifacts are ignored.
- Rewrite `AGENT.md` to describe the project's actual purpose, current
  architecture, and the dev workflow (how to build, test, and run) rather than a
  stale checklist.
- Update `Dockerfile`, `docker-compose.yml`, `entrypoint.sh`, and supporting
  config for the new layout.
- Add `.gitignore` entries for AI harnesses (`.junie/`, `.claude/`, `.cursor/`,
  `.codex/`, `.aider*`, etc.).
- Verify the whole thing still builds, tests, and runs after the move.

## Impact

- Affected specs: `repo-structure` (this change is structural; no runtime
  feature behavior changes)
- Affected code: package paths/`go.mod` module, `Dockerfile`,
  `docker-compose.yml`, `entrypoint.sh`, `vite.config.ts`, embed directive,
  `README.md`, `AGENT.md`, `.gitignore`, `.dockerignore`
- Compatibility: no API or data-model changes; same HTTP surface and behavior.
