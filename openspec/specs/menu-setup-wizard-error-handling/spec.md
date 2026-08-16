# menu-setup-wizard-error-handling Specification

## Purpose

Fix the left-nav **Setup Wizard** menu so it never silently proceeds on failure.
Today `SetupWizard.tsx` (`frontend/src/pages/SetupWizard.tsx`) masks errors:

- `fetchSetupStatus` (`SetupWizard.tsx:69-86`) swallows failures and falls back
  to "My Organization".
- `handleStep1Submit` (`SetupWizard.tsx:88-107`) advances to Step 2 on failure
  regardless ("Allow step progression in test/mock environment").
- `handleSendInvite` and `handleCompleteSetup` also proceed on error.

Operators can complete a wizard run that never actually saved anything.

## Requirements

### Requirement: Setup status load failures show a real error with Retry

When `GET /dashboard/api/tenant/setup-status` fails, the wizard MUST render an
error banner with a Retry control instead of silently substituting default
business details.

#### Scenario: Setup status request fails

- **WHEN** `GET /dashboard/api/tenant/setup-status` returns a non-2xx response
  or a network error
- **THEN** the wizard MUST show an error message with a Retry control
- **AND** it MUST NOT prefill the organization name with a fabricated value

### Requirement: Step submissions do not advance on failure

`PUT /dashboard/api/tenant/setup` must be treated as authoritative: if it
fails, the operator MUST remain on the current step with the error displayed,
so progress is never claimed without persistence.

#### Scenario: Saving Step 1 fails

- **WHEN** `PUT /dashboard/api/tenant/setup` returns a non-2xx response
- **THEN** the wizard MUST NOT advance to Step 2
- **AND** the current step MUST show the server error message and remain
  editable

### Requirement: Invite and completion actions surface real errors

Sending an invite and completing setup MUST surface real failures rather than
recording fake "dispatched" entries or force-navigating to the inbox.

#### Scenario: Completing setup fails

- **WHEN** `POST /dashboard/api/tenant/complete-setup` returns a non-2xx
  response
- **THEN** the wizard MUST NOT navigate away automatically
- **AND** it MUST show the server error with a retry action

### Requirement: The fix is covered by regression tests

#### Scenario: Regression test guards step progression on error

- **WHEN** a frontend test renders `SetupWizard` and the setup PUT request is
  rejected
- **THEN** the current step MUST remain Step 1 with the error visible
- **AND** the wizard MUST NOT show a fabricated organization name on a failed
  status load

## Notes

- Mock fallbacks were introduced to support test/mock environments; replace
  them with explicit error states.
- After source changes, rebuild the embedded frontend (`npm run build`) so the
  served bundle no longer contains the generic "Something went wrong!" surface.