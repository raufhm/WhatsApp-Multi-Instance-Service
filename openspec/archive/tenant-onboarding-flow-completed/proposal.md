# Change: Add tenant onboarding flow with self-service signup

## Why

Currently, operators must be manually created in the database by an admin. There is no self-service onboarding flow for:
- New tenants to sign up and create their organization
- Operators to create their own accounts
- Password recovery or email verification

This creates friction for adoption and requires manual intervention for every new user. A proper onboarding flow with signup, signin, password reset, and email verification will enable self-service tenant and operator management.

## What Changes

- **Self-service tenant signup**: Organizations can create their tenant account (email verification required for tenant admin)
- **TOTP-based authentication**: Operators authenticate via TOTP codes from authenticator apps (Google Authenticator, Authy, 1Password, etc.)
- **WhatsApp-first invitations**: Operator invitations sent via WhatsApp with TOTP setup link
- **No passwords**: Eliminates password storage, breaches, and reset flows
- **Backup codes**: Recovery codes for when operators lose authenticator access
- **Improved signin flow**: Tenant ID + email/WhatsApp number + TOTP code
- **Onboarding wizard**: First-time tenant admin guidance and WhatsApp account pairing

## Impact

- **Affected specs**: `operator-auth`, `security`, `tenant-management`, `whatsapp-invitations`, `totp-authentication`
- **Affected code**: 
  - New backend endpoints for TOTP setup, verification, backup codes
  - TOTP secret generation and storage (encrypted)
  - QR code generation for authenticator apps
  - WhatsApp invitations with TOTP setup links
  - Database migrations for TOTP secrets, backup codes
  - Email service remains optional (only for tenant admin signup, recovery)
- **Security improvements**: No password hashes, no password reset flows, reduced attack surface
- **API compatibility**: New endpoints added; no breaking changes to existing APIs

## Out of scope (deferred)

- Password-based authentication (deprecated in favor of TOTP)
- OAuth/Social login (Google, Microsoft)
- SSO/SAML for enterprise tenants
- SMS-based 2FA (TOTP via authenticator apps only)
- Account deactivation/self-delete flows
- Audit logging for onboarding events (deferred to monitoring spec)

## Success Metrics

- Operator onboarding completion in < 2 minutes
- WhatsApp invitation delivery rate > 95%
- Invitation acceptance rate > 70%
- TOTP setup success rate > 90%
- Reduction in support requests (no password resets)
- Zero password-related security incidents
- Zero dependency on external email providers for operator onboarding
