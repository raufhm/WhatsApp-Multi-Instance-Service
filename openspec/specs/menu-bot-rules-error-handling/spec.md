# menu-bot-rules-error-handling Specification

## Purpose

Fix the left-nav **Bot Rules** menu so operators never see an unhandled failure
("Something went wrong!") when opening or interacting with the page. Today
`BotRules.tsx` (`frontend/src/pages/BotRules.tsx:18`) loads rules via
`useDashboardBotRules()` → `GET /dashboard/api/bot-rules`; when that query fails
the page shows only a red "Failed to load bot rules." banner with no Retry
control (header Refresh exists but is not discoverable inside the error state).

## Requirements

### Requirement: The Bot Rules failure state shows an actionable error with Retry

When the `useDashboardBotRules()` query fails, the page MUST render a clear
message with a Retry button that calls `refetch()`, mirroring the Inbox error
state (`frontend/src/pages/Inbox.tsx:151-158`).

#### Scenario: Bot rules list request fails

- **WHEN** `GET /dashboard/api/bot-rules` returns a non-2xx response or a
  network error
- **THEN** the page MUST NOT show a generic "Something went wrong!" message
- **AND** it MUST show a specific error message with a Retry button that
  triggers `refetch()`

### Requirement: Rule set actions surface real mutation errors

The "Save Draft as New Version" and "Activate" actions MUST surface backend or
validation failures in the existing inline banner
(`frontend/src/pages/BotRules.tsx:56`) rather than silently failing or
requiring a full page reload.

#### Scenario: Saving an invalid or rejected rule set

- **WHEN** a draft is saved and the server rejects it
- **THEN** the inline error banner MUST display the server-provided message
- **AND** the drafts MUST remain editable so the user can correct and retry

### Requirement: Loading and empty states remain intact

While loading, a spinner MUST render; when there are no rule sets, a
"No rulesets yet." state MUST render, and the draft editor MUST still be usable.

#### Scenario: Page renders with zero rule sets

- **WHEN** `GET /dashboard/api/bot-rules` succeeds with an empty array
- **THEN** the Versions table area MUST show "No rulesets yet."
- **AND** the "Draft New Version" editor MUST remain rendered and interactive

### Requirement: The fix is covered by regression tests

#### Scenario: Regression test guards the retry path

- **WHEN** a frontend test renders `BotRules` against a failing query
- **THEN** the test MUST find a Retry button that refetches the rules

## Notes

- `useDashboardBotRules` polls every 30s (`frontend/src/hooks/useDashboard.ts:30`);
  the error state must coexist with polled refetch without flicker.
- After source changes, rebuild the embedded frontend (`npm run build`) so the
  served bundle no longer contains the generic "Something went wrong!" surface.