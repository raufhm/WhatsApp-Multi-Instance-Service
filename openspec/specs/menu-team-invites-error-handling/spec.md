# menu-team-invites-error-handling Specification

## Purpose

Fix the left-nav **Team & Invites** menu so failures are never masked. Today
`Team.tsx` (`frontend/src/pages/Team.tsx`) swallows errors in nested `try/catch`
blocks and substitutes **fake data**, which is worse than an error banner:

- `fetchOperators` (`Team.tsx:63-101`) falls back to fabricated operators
  ("Sarah Connor", "sarah@example.com", …) when
  `GET /dashboard/api/operators` fails.
- `fetchInvitations` (`Team.tsx:103-114`) silently falls back to an empty list.
- `handleSendInvite` (`Team.tsx:122-167`) fabricates a mock invitation on
  failure and reports success.
- `handleRevokeInvitation` and `handleResetTotp` also swallow errors.

Operators therefore cannot tell whether real data loaded, and the page displays
people/invitations that do not exist.

## Requirements

### Requirement: API failures render real error states, never fabricated data

When `GET /dashboard/api/operators` or `GET /dashboard/api/invitations` fails,
the affected tab MUST render a specific error message with a Retry action.
Fabricated operator and invitation records MUST be removed.

#### Scenario: Operators list request fails

- **WHEN** `GET /dashboard/api/operators` returns a non-2xx response or a
  network error
- **THEN** the Operators tab MUST show an error message with a Retry control
- **AND** the page MUST NOT display the mock operator records such as
  "Sarah Connor"

### Requirement: Mutation failures surface as errors instead of mock success

Creating an invitation, revoking an invitation, and resetting TOTP MUST report
their real failure to the operator; the mock invitation and optimistic
success fallbacks MUST be removed.

#### Scenario: Dispatch invitation fails

- **WHEN** `POST /dashboard/api/invitations/{channel}` returns a non-2xx
  response
- **THEN** the invite modal MUST show the server error message
- **AND** it MUST NOT create or display a fabricated invitation

### Requirement: Loading and empty states remain intact

While loading, spinners MUST render; genuinely empty results MUST render the
existing "No operators registered yet." / "No active or pending invitations."
empty states.

#### Scenario: Team page loads with no data

- **WHEN** both list endpoints succeed with empty arrays
- **THEN** the Operators tab MUST render "No operators registered yet."
- **AND** the Invitations tab MUST render the no-invitations empty state

### Requirement: The fix is covered by regression tests

#### Scenario: Regression tests guard against fabricated data

- **WHEN** a frontend test renders `Team` against a failing operators request
- **THEN** the test MUST find an error state with Retry
- **AND** the test MUST assert the mock operator records are absent

## Notes

- Mock fallbacks were introduced to support test/mock environments; replace
  them with explicit error states and use the real API in tests via mocks.
- After source changes, rebuild the embedded frontend (`npm run build`) so the
  served bundle no longer contains the generic "Something went wrong!" surface.