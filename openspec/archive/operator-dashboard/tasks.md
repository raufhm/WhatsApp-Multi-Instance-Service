# Tasks

## Phase 1 — Project setup and auth (MVP foundation)

- [x] Create `frontend/` directory with Vite + React + TypeScript + Tailwind + Shadcn/ui starter
- [x] Add `package.json` with dependencies and build scripts
- [x] Define TypeScript types for domain models (Conversation, Contact, Message, etc.)
- [x] Create HTTP client with auth headers and API key/session handling
- [x] Add React Query configuration with query keys and mutations
- [x] Create database migration for `operators` and `sessions` tables
- [x] Implement backend session auth (login/logout/me routes, session storage)
- [x] Serve dashboard static assets from Go on `/dashboard/*`
- [x] Add login page with form validation and session cookie handling
- [x] Add auth route guard (redirects unauthenticated users to `/login`)
- [x] Add logout endpoint and UI button

## Phase 2 — Inbox and conversation detail (core workflow)

- [x] Create reusable DataTable component (table + filters)
- [x] Implement inbox page with conversation list, filters, and ticket actions
- [x] Add activity queue section with acknowledge/dismiss mutations
- [x] Implement conversation detail page with message timeline
- [x] Style message bubbles (inbound vs outbound vs internal notes)
- [x] Add reply composer (text input, send button)
- [x] Add ticket controls: assign, handoff, close, reopen
- [x] Add error boundaries and retry logic
- [x] Add real-time polling (30s) for inbox updates

## Phase 3 — Additional pages and polish

- [x] Contact list page (detail/editing deferred)
- [x] Account management page with list and health indicators
- [x] Bot rules editor with draft/publish workflow
- [x] Upload jobs failure indicator page
- [x] Empty states (no tickets, no contacts, etc.)
- [x] Responsive layouts (mobile, tablet, desktop)
- [ ] Keyboard shortcuts (navigate, send, close)
- [ ] Theme and branding (light/dark mode optional)
- [ ] Accessibility audit (ARIA labels, focus management)

## Phase 4 — Testing and deployment

- [ ] Unit tests for components and hooks
- [x] UI test setup (Vitest + React Testing Library)
- [x] Build pipeline (Vite production build)
- [x] Deployment documentation (copy dist/, set env vars)
- [x] README updates (dashboard usage, local dev workflow)
