# Design: Operator dashboard architecture and workflow

## Stack

- **Framework**: React 18 + TypeScript (Vite for dev/build)
- **Styling**: Tailwind CSS with Shadcn/ui or Radix components
- **State**: TanStack Query for server state, context/atoms for UI state
- **Routing**: TanStack Router (code-based routes) with client-side navigation
- **Tables**: TanStack Table (headless) for sortable/filterable/paginated data grids
- **Auth**: Session cookie (HttpOnly, Secure, SameSite=Lax) + CSRF token
- **Build**: Vite (dev server + production build to `dist/`)
- **Backend**: New `/dashboard` route on Server serving static assets from `dist/`

## Authentication flow

1. POST `/dashboard/login` with `{email, password}` → validates against `operators` table, returns session cookie
2. Session stored in `sessions(id, operator_id, expires_at)` with HttpOnly cookie `sid`
3. Server middleware verifies session on all `/dashboard/*` routes, redirects to `/login` if invalid
4. Dashboard GET `/dashboard` serves `index.html` with JS/CSS; TanStack Router handles client routes
5. Logout: POST `/dashboard/logout` clears session

## Directory structure

```
frontend/
  src/
    components/  # Reusable UI (DataTable, MessageBubble, etc.)
    pages/       # Route components (Inbox, ConversationDetail, etc.)
    hooks/       # Custom hooks (useInbox, useConversation, etc.)
    lib/         # HTTP client, query keys, utils
    types/       # TypeScript types
  vite.config.ts
  package.json
```

## Core pages

### Inbox (`/dashboard/inbox`)
- DataTable with columns: ticket#, contact, account, status, priority, assignee, last activity
- Filters: status (multi-select), account, assignee, priority, date range, unread
- Actions: assign, handoff, close, reopen (via POST to `/api/v1/operator/*`)
- Activity queue section with acknowledge/dismiss
- Real-time poll (30s interval) or SSE for updates

### Conversation detail (`/dashboard/conversations/:id`)
- Header with contact, account, ticket#, status badges
- Message timeline (scrollable) with media previews, internal notes styled differently
- Reply composer (text + media upload) → POST to `/api/v1/accounts/:account/messages`
- Ticket controls: assign, handoff, close, reopen, merge, split
- Bot session state display

### Contact detail (`/dashboard/contacts/:id`)
- Contact metadata (name, number, tags)
- Message history (link to conversations)
- Editable fields

### Account management (`/dashboard/accounts`)
- List of paired WhatsApp accounts with health indicators
- QR pairing flow (open new window, poll status, show success/error)

### Bot rules (`/dashboard/bot-rules`)
- List of versions with publish status
- Draft editor (form or JSON) with validation
- Publish/draft/save workflow
- Version rollback

## Error and loading states

- Loading: skeleton screens or spinners
- Empty: illustration + CTA ("No tickets" → "Create one")
- Error: error boundary with retry, "Something went wrong"
- Permission denied: 403 page
- Not found: 404 page

## Build and deployment

- `npm run build` in `frontend/` → `dist/` with hashed assets
- Copy `dist/` into project root or serve from Go's embedded filesystem
- `/dashboard` routes serve `index.html`; all other assets use hashed filenames for cache busting

## Testing

- Unit: React Testing Library for components
- Integration: Cypress or Playwright for E2E flows (login → inbox → conversation → reply)
- Mock API responses or run against local backend
