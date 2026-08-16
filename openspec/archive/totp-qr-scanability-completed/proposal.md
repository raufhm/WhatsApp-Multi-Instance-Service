# Change: Investigate and fix TOTP QR code scanability with Google Authenticator

## Why

A tenant admin reported that the TOTP enrollment QR code cannot be scanned with
Google Authenticator, blocking signup/TOTP setup. The duplicate-whatsapp
constraint was cleaned up, but if the QR itself is malformed or unreadable, the
flow remains broken for every enrollee, not just one.

The QR is generated from an `otpauth://` URI. Google Authenticator enforces the
IETF `otpauth` conventions (RFC 6238 + the common Google extension rules) far
more strictly than other apps, so even a small URI defect makes the scan fail
while Authy/1Password may still import it.

## Investigation findings (so far)

Initial code review suggests two concrete suspects in the URI construction:

1. **Issuer / label mismatch.** `backend/internal/totp/totp.go`
   `GenerateOtpauthURL` builds the label as
   `AccountPrefix + ":" + account` (`WhatsApp Service:<account>`) but sets
   `issuer=Issuer` which is **`WhatsApp Multi-Instance Service`**. Many
   authenticator apps require the label provider prefix to match the `issuer`
   parameter, else they reject the entire URI or mis-name the account.
2. **Encoding divergence between backend and frontend.** The Go backend emits
   `otpauth://totp/WhatsApp%20Service:<account>?issuer=WhatsApp+Multi-Instance+Service...`
   while the React fallback (`TotpQrCode.tsx`) builds a different URI
   `otpauth://totp/WhatsApp%20Service:<account>?issuer=WhatsApp%20Service`
   from a different `Issuer` constant, and `VerifyEmail.tsx` hardcodes yet
   another URI. Tests only assert a prefix, so none of this divergence is
   caught.

This change is deliberately scoped as **investigate-then-fix**: reproduce and
confirm the exact failure with Google Authenticator, then apply the minimal URI
correction(s).

## What Changes

- Confirm the exact failure mode of the current `otpauth://` URI in Google
  Authenticator (and at least one control app such as Authy/1Password).
- Normalize the TOTP URI construction to a single source of truth shared by
  backend and frontend:
  - a single issuer/label description with matching `issuer` parameter,
  - standard `otpauth://totp/{issuer}:{account}` labeling,
  - correct percent-encoding of both label components,
  - explicit `algorithm=SHA1&digits=6&period=30` parameters.
- Fix the Go `GenerateOtpauthURL` and the frontend fallback to emit identical,
  spec-compliant URIs; remove hardcoded test/demo URIs where feasible.
- Add regression tests asserting the full URI (not just prefix) and,
  where possible, decode-and-reparse the QR payload.

## Impact

- **Affected specs**: `totp-qr-scanability` (new)
- **Affected code**:
  - `backend/internal/totp/totp.go` (`GenerateOtpauthURL`, constants)
  - `backend/internal/totp/totp_test.go`
  - `frontend/src/components/ui/TotpQrCode.tsx`
  - `frontend/src/lib/qrCode.ts`
  - `frontend/src/pages/VerifyEmail.tsx` (fallback demo URIs)
- **API compatibility**: Internal `otpauth://` URI format changes, but it is
  only ever consumed by authenticator-app scans; no wire/API break.
- **Security**: No secrets change; enrollment secrets remain encrypted at rest.

## Out of scope (deferred)

- Switching to WebAuthn / passkeys as an alternative to TOTP.
- Adding QR download/print affordances.
- Persisting the otpauth URI (it is regenerated per enrollment by design).

## Success Metrics

- A fresh TOTP QR scans and imports successfully in Google Authenticator.
- Google Authenticator shows the expected issuer (not "undefined" or the
  account URL muddled with query parameters).
- Backend and frontend URIs are byte-identical for the same secret/account.
- `go test ./internal/totp/...` and frontend test suite pass.