# menu-accounts-error-handling Specification

## Purpose

Fix the left-nav **Accounts** menu so operators never see an unhandled failure
("Something went wrong!") when opening the page. Today `Accounts.tsx`
(`frontend/src/pages/Accounts.tsx:18`) loads via `useDashboardAccounts()`
→ `GET /dashboard/api/accounts`; on failure it shows only
"Failed to load accounts." with no Retry control. Because the backend enforces
the `view-accounts` permission (`backend/handler/dashboard_api.go:57-59`),
permission denials surface as the same undifferentiated error as a server
failure.

## Requirements

### Requirement: The Accounts failure state shows an actionable error with Retry

When `useDashboardAccounts()` fails, the page MUST render a specific message
with a Retry button that calls `refetch()`.

#### Scenario: Accounts list request fails

- **WHEN** `GET /dashboard/api/accounts` returns a non-2xx response or a
  network error
- **THEN** the page MUST NOT show a generic "Something went wrong!" message
- **AND** it MUST show a specific error message with a Retry button that
  triggers `refetch()`

### Requirement: Permission failures are distinguished from server failures

A `403` from `GET /dashboard/api/accounts` MUST render a dedicated
"You don't have permission to view accounts" state (with guidance to contact an
admin) instead of the generic failure, since the dashboard route already
returns `FORBIDDEN` for unauthorized roles
(`backend/handler/dashboard_api.go:135`).

#### Scenario: Viewer role opens Accounts

- **WHEN** an operator without `view-accounts` permission opens the Accounts
  menu and the API returns `403 FORBIDDEN`
- **THEN** the page MUST render the dedicated insufficient-permission message
- **AND** it MUST NOT render the generic error or empty "Link device" CTA

### Requirement: Loading and empty states remain intact

The loading spinner and the "No WhatsApp accounts linked" empty state with the
"Link Your First Device" CTA MUST continue to render correctly.

#### Scenario: Page renders with zero accounts

- **WHEN** `GET /dashboard/api/accounts` succeeds with an empty array
- **THEN** the page MUST render the "No WhatsApp accounts linked" card with the
  pairing CTA

### Requirement: The fix is covered by regression tests

#### Scenario: Regression test guards retry and permission states

- **WHEN** a frontend test renders `Accounts` against a failing query
- **THEN** the test MUST find a Retry button
- **AND** a second test simulating a `403` MUST find the permission-denied
  message

## Notes

- `useDashboardAccounts` polls every 10s (`frontend/src/hooks/useDashboard.ts:18`);
  the error and permission states must coexist with polling without flicker.
- After source changes, rebuild the embedded frontend (`npm run build`) so the
  served bundle no longer contains the generic "Something went wrong!" surface.