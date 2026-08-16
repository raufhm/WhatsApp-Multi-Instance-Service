# Repository structure

## ADDED Requirements

### Requirement: Backend and frontend are separated into top-level directories

The repository MUST keep the Go service under `backend/` and the web UI under
`frontend/`, with build output clearly separated and ignored.

#### Scenario: Root is clean

- **WHEN** a contributor lists the repo root
- **THEN** they MUST see `backend/`, `frontend/`, and top-level config/docs,
  not loose Go source files

### Requirement: Source documentation is version-controlled

Markdown source (`README.md`, `AGENT.md`, and all `openspec/**/*.md`) MUST be
tracked by git and MUST NOT be covered by a blanket ignore rule.

#### Scenario: openspec is committed

- **WHEN** a change is added under `openspec/`
- **THEN** `git status` MUST show those files as untracked/trackable, not ignored

### Requirement: AI harness and build artifacts are ignored

AI-harness directories, secrets, native caches, and build outputs MUST be
ignored by git.

#### Scenario: Clean working tree

- **WHEN** a contributor uses an AI harness (e.g. `.junie/`, `.claude/`) and
  builds the project
- **THEN** those directories and `dist/`, `node_modules`, and `.env` MUST be
  listed as ignored

### Requirement: The project remains buildable and testable after restructuring

The backend and frontend MUST build and test green after the directory move,
with Docker still producing a working image.

#### Scenario: Build and test

- **WHEN** a contributor runs the backend build/test and the frontend build
- **THEN** all MUST pass
- **AND** `docker compose` MUST build successfully

### Requirement: AGENT.md reflects the project and its development workflow

`AGENT.md` MUST describe the project's purpose, current state, architecture, and
how it is built, tested, and run during development, without stale references.

#### Scenario: Accurate guidance

- **WHEN** a new agent or contributor reads `AGENT.md`
- **THEN** the documented build/test/run commands and architecture MUST match the
  current repository layout and behavior
