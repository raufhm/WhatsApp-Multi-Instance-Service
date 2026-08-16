# Operator dashboard deployment and operations

## ADDED Requirements

### Requirement: Frontend builds reproducibly

The build process MUST produce deterministic, cache-busted assets.

#### Scenario: Production build

- **WHEN** `npm run build` is executed
- **THEN** it MUST generate hashed filenames (e.g., `app.a1b2c3.js`)
- **AND** produce a manifest or index.html that references them

### Requirement: Static assets are served efficiently

The Go server MUST serve dashboard assets with appropriate cache headers.

#### Scenario: Asset caching

- **WHEN** a browser requests a hashed asset
- **THEN** the server MUST set long cache TTL (e.g., 1 year)
- **AND** rely on filename hashing for cache busting

#### Scenario: HTML is not cached

- **WHEN** a browser requests `/dashboard` or `/dashboard/*` HTML routes
- **THEN** the server MUST set no-cache headers to avoid stale SPA routes

### Requirement: Local development is productive

Developers MUST be able to work on frontend and backend simultaneously.

#### Scenario: Vite dev server proxy

- **WHEN** running `npm run dev` in frontend/
- **THEN** Vite MUST proxy API requests to the Go backend (e.g., http://localhost:3000)
- **AND** enable hot module replacement for fast iteration

#### Scenario: Go serves dev assets

- **WHEN** running the Go server in development mode
- **THEN** it MAY proxy to the Vite dev server for assets
- **OR** serve a development build from `dist/`

### Requirement: Environment configuration is documented

Deployment instructions MUST cover environment variables, build steps, and asset serving.

#### Scenario: Production deployment

- **WHEN** deploying to production
- **THEN** the README MUST document: build frontend, copy dist/, set API base URL, configure auth
