# menu-contacts-error-handling Specification

## Purpose

Stop the left-nav **Contacts** menu from rendering a broken or erroring page.
Operators report that clicking Contacts shows a generic failure
("Something went wrong!") instead of the customer directory. Two concrete
defects exist today:

1. The served UI is the Go-embedded build (`backend/dist`), which predates the
   current source and contains a generic "Something went wrong!" error surface.
   Rebuilding the frontend embeds the current page, but even the current page
   has no actionable error handling.
2. The Contacts page cannot render real rows: `Contacts.tsx`
   (`frontend/src/pages/Contacts.tsx:11-14`) reads `c.name`, `c.number`, and
   `c.tags`, but the backend serializes `domain.Contact`
   (`backend/domain/models.go:51-56`) without JSON field tags, emitting
   `DisplayName`, `NormalizedAddress`, `ProviderAddress`, and `Metadata`
   instead. Every contact card therefore renders blank.

## Requirements

### Requirement: The backend serializes contacts using the frontend's expected schema

The contact payload returned by `GET /api/v1/contacts` and
`GET /api/v1/contacts/{id}` MUST include `id`, `tenant_id`, `name`, `number`,
`email`, `tags`, `created_at`, and `updated_at` fields that the frontend
`Contact` type consumes (`frontend/src/types/index.ts:51-60`).

#### Scenario: Contacts payload contains display-facing fields

- **WHEN** a tenant with at least one stored contact calls
  `GET /api/v1/contacts`
- **THEN** each returned object MUST contain a populated `name`
- **AND** `number` MUST be the contact's provider address and `email`/`tags`
  MUST be read from contact metadata when present
- **AND** no field key MUST use the camel-cased Go struct names
  (`DisplayName`, `NormalizedAddress`, `ProviderAddress`)

### Requirement: The Contacts page failure state shows an actionable error with Retry

When the `useContacts()` query fails, the page MUST render a clear message with
a Retry control that calls `refetch()`, matching the pattern already used by
the Inbox page (`frontend/src/pages/Inbox.tsx:151-158`).

#### Scenario: Contacts list request fails

- **WHEN** `GET /api/v1/contacts` returns a non-2xx response or network error
- **THEN** the page MUST NOT render generic text such as "Something went wrong!"
- **AND** it MUST show a specific message and a Retry button that triggers
  `refetch()`

### Requirement: The Contacts page preserves loading, empty, and data states

The page MUST keep a loading spinner, a "No contacts found" empty state, and the
searchable contact-card grid, and these states MUST render the real data
produced by the fixed schema.

#### Scenario: Contacts data loads successfully

- **WHEN** `GET /api/v1/contacts` succeeds with an empty list
- **THEN** the page MUST render the "No contacts found" empty state
- **AND** when the list is non-empty, cards MUST show the contact's name and
  number (non-blank)

### Requirement: The fixed Contacts page is what operators actually run

The "Something went wrong!" text observed by operators is served from the stale
embedded bundle. Freshly built frontend assets (`npm run build`, output to
`backend/dist`) MUST be embedded so the dashboard serves this corrected page.

#### Scenario: Embedded bundle is rebuilt and served

- **WHEN** the frontend is rebuilt and the backend restarted
- **THEN** navigating to the Contacts menu MUST render the corrected contacts
  page with real data and retry handling
- **AND** the served bundle MUST NOT contain the generic "Something went wrong!"
  error surface for this page

### Requirement: The fix is covered by regression tests

Contact schema mapping and the page's retry behavior MUST be covered by tests.

#### Scenario: Regression tests guard schema and retry

- **WHEN** backend contact serialization tests run against a contact with
  metadata
- **THEN** the JSON MUST expose `name`, `number`, `email`, and `tags`
- **AND** a frontend test rendering `Contacts` against a failing query MUST
  find a Retry button that refetches

## Notes

- The `Contact` struct lacks JSON tags because no DTO was defined when the
  dashboard was introduced; the fix should introduce an explicit DTO rather
  than tagging the domain struct.
- The observed "Something went wrong!" is the stale `backend/dist` bundle; the
  same rebuild note applies to all menu specs in this series.