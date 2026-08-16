# menu-security-totp-error-handling Specification

## Purpose

Fix the left-nav **Security & TOTP** menu so it never reports fake security
state. Today `TotpSettings.tsx` (`frontend/src/pages/TotpSettings.tsx`)
fabricates real-looking data on failure:

- `fetchTotpStatus` (`TotpSettings.tsx:36-52`) catches failures and sets a mock
  status (`enabled: true`, current `verified_at`, `backup_codes_remaining: 8`).
- `handleRegenerateCodes` (`TotpSettings.tsx:54-87`) falls back to hard-coded
  backup codes on failure.

Operators can be misled into believing TOTP is enabled and protected when the
call never succeeded.

## Requirements

### Requirement: TOTP status load failures show a real error with Retry

When `GET /dashboard/api/account/totp` fails, the page MUST render an error
banner with a Retry control and MUST NOT display a fabricated enabled status.

#### Scenario: TOTP status request fails

- **WHEN** `GET /dashboard/api/account/totp` returns a non-2xx response or a
  network error
- **THEN** the page MUST show a specific error message with a Retry control
- **AND** it MUST NOT render a fabricated "enabled" badge or backup-code count

### Requirement: Backup code regeneration never returns fabricated codes

When `POST /dashboard/api/account/totp/regenerate-backup-codes` fails, the
generate/regenerate flow MUST show the real error and MUST NOT display the
hard-coded fallback codes.

#### Scenario: Regenerating backup codes fails

- **WHEN** the regeneration endpoint returns a non-2xx response
- **THEN** the regeneration error MUST be shown (`TotpSettings.tsx:83`
  behavior preserved)
- **AND** no hard-coded fallback codes MUST be displayed or stored

### Requirement: The TOTP settings page renders accurately in all states

Loading, enabled, disabled/not configured, and reset states MUST reflect the
real API response.

#### Scenario: TOTP is not configured

- **WHEN** `GET /dashboard/api/account/totp` succeeds and reports TOTP disabled
- **THEN** the page MUST render the not-configured state accurately without
  implying the authenticator is active

### Requirement: The fix is covered by regression tests

#### Scenario: Regression test guards against mock status and codes

- **WHEN** a frontend test renders `TotpSettings` against a failing status
  request
- **THEN** the test MUST find an error state with Retry
- **AND** the test MUST assert no fabricated backup codes are shown on a
  failed regeneration

## Notes

- Mock fallbacks were introduced to support test/mock environments; replace
  them with explicit error states.
- After source changes, rebuild the embedded frontend (`npm run build`) so the
  served bundle no longer contains the generic "Something went wrong!" surface.