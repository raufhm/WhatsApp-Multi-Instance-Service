# Change: Add operator dashboard for tenant workflows

## Why

The backend provides a complete API but operators must interact through raw HTTP requests or build external clients. A browser-based dashboard improves usability for human operators handling conversations, activities, contacts, and accounts.

## What Changes

- A React-based (or Vue/Svelte) frontend served from `/dashboard/*`
- Tenant/operator authentication and session handling
- Inbox view with filtering, ticket actions, and activity queue
- Conversation detail with message timeline, reply composer, and merge/split
- Contact and account management screens
- Bot-rule management with draft/publish workflow
- Loading, empty, error, and permission states throughout

## Impact

- Affected specs: `operator-auth`, `inbox-ui`, `conversation-detail`, `contact-management`
- Affected code: new `frontend/` directory, build pipeline, static asset serving
- API compatibility: No breaking changes; dashboard consumes existing `/api/v1/*` endpoints

## Out of scope (deferred)

- WebSocket/push for real-time updates
- Complex routing or micro-frontend architecture
- Browser-based onboarding/QR code capture
- Analytics or reporting dashboards
