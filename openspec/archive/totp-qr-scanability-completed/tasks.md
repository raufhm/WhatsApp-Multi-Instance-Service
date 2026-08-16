# Tasks

## Phase 1 — Reproduce and confirm

- [x] Reproduce: complete a signup, open TOTP setup, attempt Google Authenticator scan; record exact failure.
- [x] Capture the exact `otpauth_url` from the API response and decode it; record against expected format.
- [x] Control test: scan the same QR in Authy/1Password; determine if failure is Google-specific.
- [x] Variant sweep: matched issuer, `%20` vs `+`, per-component escaping; identify which makes Google Authenticator succeed.
- [x] Update `proposal.md`/`design.md` with confirmed root cause and final encoding decisions.

## Phase 2 — Backend fix

- [x] Introduce a single issuer constant used for both label prefix and `issuer` param.
- [x] Rewrite `GenerateOtpauthURL` to emit a spec-compliant URI per confirmed design.
- [x] `totp_test.go`: assert exact full URI (golden string) instead of prefix-only.

## Phase 3 — Frontend parity

- [x] Export shared issuer + URI builder in `frontend/src/lib/qrCode.ts`.
- [x] `TotpQrCode.tsx` fallback builds the identical URI from the shared builder.
- [x] `VerifyEmail.tsx`: remove/hard-align hardcoded demo URIs (fake secret, third issuer spelling).

## Phase 4 — Tests and verification

- [x] Backend round-trip test: `GenerateQRCodeDataURL` → decode → parse `otpauth://` → verify scheme/label/params.
- [x] Frontend golden-string test: `qrCode.ts` output identical to Go output for fixed secret/account.
- [x] `go test ./internal/totp/...` passes.
- [x] Frontend test suite passes (`npm test` / `vitest`).
- [x] Manual QA checklist: real scan succeeds in Google Authenticator; issuer label is correct; control app still works.