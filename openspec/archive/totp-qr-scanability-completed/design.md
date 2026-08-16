# Design: TOTP QR scanability investigation and fix

## Current behavior (confirmed by code)

Backend URI produced by `GenerateOtpauthURL(account, secret)`:

```text
otpauth://totp/WhatsApp%20Service:user@example.com?algorithm=SHA1&digits=6&issuer=WhatsApp+Multi-Instance+Service&period=30&secret=ABCD...
```

Issuer is applied **three** different ways across the stack:

| Location | Label prefix | `issuer` param |
| --- | --- | --- |
| Go `totp.go` (`AccountPrefix`/`Issuer`) | `WhatsApp Service` | `WhatsApp Multi-Instance Service` |
| Frontend `TotpQrCode.tsx` fallback (default prop) | `WhatsApp Service` | `WhatsApp Service` |
| Frontend `VerifyEmail.tsx` hardcoded fallback | `WhatsApp` | `WhatsApp` |

Only the prefix is asserted in `totp_test.go:159`, so none of these mismatches
are detected by tests.

## Root-cause hypothesis

Google Authenticator parses `otpauth://totp/{label}?...` and, per the widely
implemented convention, takes everything before the first `:` inside the label
as the issuer. When an explicit `issuer` query parameter exists it must agree
with that prefix; a mismatch commonly causes Authenticator to fall back to a
bare account name, show an odd/mismatched entry, or fail the scan outright.
Here the prefix (`WhatsApp Service`) and the `issuer` parameter
(`WhatsApp Multi-Instance Service`) disagree, which is the primary suspect.

Secondary suspects to validate during investigation:

- `url.PathEscape` on the whole `Prefix:account` label vs. escaping each
  component separately (the `:` must be preserved as the label separator).
- `url.Values.Encode()` encoding spaces as `+` inside the query (`issuer=WhatsApp+Multi-Instance+Service`). Most apps accept this, but some strict parsers prefer `%20`.
- Display size / quiet-zone: `TotpQrCode.tsx` renders at `w-48 h-48` (192px) from a 256px PNG — likely fine, but the SVG path (`generateQrSvg`) is a separate code path to compare.

## Investigation protocol (before any code change)

1. Reproduce current behavior with real backend data: run the app, complete a
   signup, open the TOTP setup page, and attempt a Google Authenticator scan.
2. Capture the actual `otpauth://` URI string (from the API response
   `otpauth_url`) and decode it; verify against the standards.
3. Try scanning the same QR in Authy/1Password as a control to isolate whether
   the failure is Google-specific or universal.
4. Test re-encoded variants (matched issuer; `%20` vs `+`; label components
   escaped separately) to find which change makes the scan succeed.
5. Record findings in the change; proceed to a fix only for the confirmed
   difference(s).

## Fix design (pending confirmation)

### Single issuer constant, shared semantics

Introduce one issuer string used as **both** the label prefix and the `issuer`
parameter, everywhere:

```go
const Issuer = "WhatsApp Multi-Instance Service"
```

`GenerateOtpauthURL` becomes:

```go
label := fmt.Sprintf("%s:%s", Issuer, account)
labelEscaped := url.QueryEscape(label) // otpauth label uses query-param escaping semantics
v := url.Values{}
v.Set("secret", cleaned)
v.Set("issuer", Issuer)
v.Set("algorithm", "SHA1")
v.Set("digits", "6")
v.Set("period", "30")
return fmt.Sprintf("otpauth://totp/%s?%s", labelEscaped, v.Encode())
```

Acknowledge that the exact default (e.g. `WhatsApp Multi-Instance Service`) is
data, not code: the key property is that label prefix and `issuer` match and
are encoded consistently.

### Frontend parity

- `TotpQrCode.tsx` must not default to a divergent issuer; either it receives
  the backend URI (`otpauthUrl`) or it constructs the identical URI from the
  shared issuer constant exported from `frontend/src/lib/qrCode.ts`.
- Remove/adjust the hardcoded demo URIs in `VerifyEmail.tsx` (they use a fake
  secret `JBSWY3DPEHPK3PXP` and a third issuer spelling) so fallback rendering
  never shows an obviously-wrong URI.

### Encoding decisions (finalized during investigation)

- Preserve `:` between issuer and account; escape each component.
- Use `%20` (path-style) or `+` for spaces consistently — decided by the
  reproduction step, defaulting to whatever Google Authenticator accepts.

## Test strategy

- `totp_test.go`: assert the **full** otpauth URL matches an expected exact
  string (issuer prefix + `issuer` param + encoding), not just a prefix.
- Add a round-trip test: `GenerateQRCodeDataURL` → decode data URL → decode QR
  payload → parse `otpauth://` → confirm scheme, label, and params.
- Frontend: unit test `qrCode.ts` URI builder parity with the Go output for a
  fixed secret/account (golden string test).
- Manual QA checklist (in `tasks.md`) for Google Authenticator + one control app.

## Out of scope

- WebAuthn/passkeys, download/print QR affordances, URI persistence.