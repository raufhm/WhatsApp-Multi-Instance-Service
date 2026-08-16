# Tenant Onboarding Flow - Implementation Plan Summary

## Overview

This spec-driven plan implements a **WhatsApp-first, TOTP-based** self-service onboarding flow. Key innovations:
- **No passwords**: TOTP authentication (Google Authenticator, Authy, 1Password, Bitwarden, etc.)
- **WhatsApp invitations**: Operator invites via WhatsApp with TOTP setup link
- **Zero email dependency**: No email provider lock-in for operator onboarding

### Core Principles

- ✅ **TOTP authentication**: All operators use authenticator apps (no passwords)
- ✅ **Operator invitations**: Sent via WhatsApp with TOTP setup link
- ✅ **Backup codes**: Recovery when operators lose authenticator access
- ✅ **Admin TOTP reset**: Recovery path without email dependency
- ✅ **WhatsApp notifications**: TOTP reset, recovery, security alerts
- ⚠️ **Email remains only for**: Tenant admin signup (one-time verification)

## What's Included

### 📋 Specification Documents

Located in `openspec/changes/tenant-onboarding-flow/`:

1. **proposal.md** - High-level overview, why, what changes, impact, success metrics
2. **design.md** - Detailed technical design, database schema, API endpoints, security considerations
3. **tasks.md** - Comprehensive implementation tasks organized in 7 phases

### 📐 Detailed Specs

Located in `openspec/changes/tenant-onboarding-flow/specs/`:

1. **signup/spec.md** - Tenant and operator signup requirements
   - Self-service tenant registration (email verification for admin only)
   - Operator signup via WhatsApp invitation with TOTP setup
   - TOTP QR code generation and verification
   - Backup codes generation
   - Form validation and rate limiting
   - Setup wizard for new tenants

2. **signin/spec.md** - TOTP authentication requirements
   - Login with tenant ID + email/WhatsApp number + TOTP code
   - Backup code login fallback
   - Tenant ID handling and remember me
   - Session management
   - Rate limiting (TOTP and backup codes)
   - UX requirements for TOTP input

3. **totp-authentication/spec.md** ⭐ **NEW** - Core TOTP authentication
   - TOTP secret generation and encrypted storage (AES-256-GCM)
   - QR code generation for authenticator apps
   - TOTP verification during login (6-digit, 30s period)
   - Backup codes (10 single-use codes)
   - Admin TOTP reset workflow
   - Multi-device TOTP support
   - Rate limiting and timing attack prevention

4. **totp-reset-recovery/spec.md** ⭐ **NEW** - Recovery without passwords
   - Backup code login (primary recovery)
   - Admin TOTP reset (when no backup codes)
   - Account recovery via WhatsApp
   - Recovery audit logging
   - Rate limiting for recovery endpoints

5. **whatsapp-invitations/spec.md** - WhatsApp-first invitations
   - Operator invitations via WhatsApp with TOTP setup link
   - WhatsApp number as primary identifier
   - TOTP setup during invitation acceptance
   - Backup codes generation
   - WhatsApp message templates
   - Delivery tracking and analytics
   - Rate limiting and privacy compliance

6. **email-verification/spec.md** - Email verification (tenant admin only)
   - One-time verification for tenant admin signup
   - Optional for operators who provide email
   - Minimal email dependency

## Implementation Phases

### Phase 1: Foundation (Week 1-2)
- Database migrations for TOTP secrets, backup codes, invitations, delivery tracking
- WhatsApp message template service
- WhatsApp message templates (invitation, TOTP reset, recovery)
- TOTP secret encryption/decryption utilities
- QR code generation library integration
- Token storage implementations (TOTP, backup codes, recovery)
- Minimal email service for tenant admin fallback only

### Phase 2: Signup Flow with TOTP (Week 2-3)
- Tenant signup endpoint (email verification for admin)
- WhatsApp invitation endpoint with TOTP secret generation
- Operator signup via WhatsApp link with TOTP setup
- QR code display and manual entry fallback
- TOTP verification during signup
- Backup codes generation and display
- Rate limiting middleware
- Frontend signup pages with TOTP setup UI

### Phase 3: TOTP Login and Recovery (Week 3-4)
- TOTP login endpoint (tenant ID + email/number + TOTP code)
- Backup code login endpoint
- TOTP reset by admin endpoint
- Account recovery request via WhatsApp
- WhatsApp message delivery for TOTP reset/recovery
- Frontend login page with TOTP input
- Backup code login UI
- Recovery request UI

### Phase 4: WhatsApp Invitation Management (Week 4-5)
- Create WhatsApp invitation endpoint (with TOTP secret)
- Create email invitation endpoint (fallback, rare)
- List/revoke invitations
- Accept invitation page with TOTP setup flow
- Frontend invitation management UI
- Delivery tracking and analytics
- TOTP reset notifications via WhatsApp

### Phase 5: Tenant Setup Wizard (Week 5-6)
- Setup status endpoints
- Update setup step endpoint
- Complete setup endpoint
- Multi-step wizard UI (org details + WhatsApp pairing)
- First-login redirect to wizard

### Phase 6: Polish & Testing (Week 6-7)
- Unit, integration, and E2E tests
- Security review (TOTP handling, encryption, rate limiting)
- Accessibility audit (TOTP input, QR code alternatives)
- Documentation (TOTP setup guides for users)

### Phase 7: Deployment & Monitoring (Week 7-8)
- Environment configuration (WhatsApp API, TOTP encryption key)
- Database migration deployment
- WhatsApp message delivery monitoring
- Metrics and alerts (TOTP setup success, backup code usage)
- Launch

## Key Features

### For New Tenant Admins
1. Sign up with organization details
2. Verify email (one-time, required for admin)
3. Complete setup wizard (org info + WhatsApp pairing)
4. **Set up TOTP with QR code** ⭐ NEW
5. **Download backup codes** ⭐ NEW
6. Access dashboard

### For New Operators
**Option A: WhatsApp Invitation (Primary, Recommended)**
1. Admin sends invitation via WhatsApp message
2. Click WhatsApp link (pre-filled: number, role, tenant)
3. Enter name (email optional)
4. **Scan QR code with authenticator app** ⭐ NEW
5. **Enter TOTP code to verify setup** ⭐ NEW
6. **Download 10 backup codes** ⭐ NEW
7. Logged in automatically

**Option B: Email Invitation (Fallback)**
1. Admin sends invitation via email (if WhatsApp unavailable)
2. Click email link (pre-filled form)
3. **Set up TOTP with QR code** ⭐ NEW
4. **Download backup codes** ⭐ NEW
5. Email verification (required for email invitations)
6. Login

### For Existing Users (Daily Use)
- **Login with**: Tenant ID + email/WhatsApp number + **TOTP code** ⭐ CHANGED
- Remember me option (extends session to 30 days)
- **Recovery**: Use backup code OR contact admin for TOTP reset ⭐ CHANGED
- Session management (8-hour default expiry)

## Database Changes

### New Tables
- `totp_secrets` - Store encrypted TOTP secrets per operator
- `totp_backup_codes` - Store hashed backup codes (single-use)
- `totp_recovery_tokens` - Recovery tokens for lost access
- `invitations` - Store operator invitations (whatsapp/email/manual channels)
- `whatsapp_invitation_delivery` - Track WhatsApp message delivery status
- `recovery_audit_log` - Audit all recovery events

### Modified Tables
- `operators`:
  - Add `whatsapp_number` (unique, indexed)
  - Add `totp_secret_encrypted` (AES-256-GCM encrypted)
  - Add `totp_verified_at` (when TOTP was verified)
  - Add `totp_setup_required` (after admin reset)
  - Add `email_verified_at` (optional for operators)
  - Make `email` optional (nullable)
- `tenants` - Add setup tracking fields

## New API Endpoints

### Authentication (TOTP-Based)
- `POST /dashboard/api/signup/tenant` (email verification for admin, then TOTP setup)
- `POST /dashboard/api/signup/operator` (via WhatsApp invitation with TOTP)
- `POST /dashboard/api/login` (tenant ID + email/number + TOTP code)
- `POST /dashboard/api/login/backup-code` (fallback recovery)
- `POST /dashboard/api/logout`

### TOTP Setup and Management
- `GET /dashboard/api/totp/setup/:token` (get QR code for invitation/signup)
- `POST /dashboard/api/totp/verify-setup` (verify TOTP during setup)
- `GET /dashboard/api/account/totp` (get TOTP status, backup codes remaining)
- `POST /dashboard/api/account/totp/regenerate-backup-codes` (generate new codes)

### TOTP Reset and Recovery
- `POST /dashboard/api/recovery/request` (request recovery via WhatsApp)
- `GET /dashboard/api/recovery/:token` (recovery instructions page)
- `POST /dashboard/api/operators/:id/totp-reset` (admin resets operator TOTP)
- `GET /dashboard/api/operators/:id/totp-status` (admin views TOTP status)

### WhatsApp Invitations
- `POST /dashboard/api/invitations/whatsapp` (primary, includes TOTP secret)
- `POST /dashboard/api/invitations/email` (fallback, rare)
- `GET /dashboard/api/invitations` (list all invitations)
- `DELETE /dashboard/api/invitations/:id` (revoke invitation)
- `GET /dashboard/api/invitations/accept/:token` (accept with TOTP setup)

### Tenant Setup
- `GET /dashboard/api/tenant/setup-status`
- `PUT /dashboard/api/tenant/setup`
- `POST /dashboard/api/tenant/complete-setup`

## New Frontend Pages

- `/signup` - Signup landing (choose tenant vs operator)
- `/signup/tenant` - Tenant signup form + **TOTP setup** ⭐ NEW
- `/signup/operator` - Operator signup via invitation + **TOTP setup** ⭐ NEW
- `/login` - Login with **TOTP code input** ⭐ CHANGED
- `/verify-email` - Email verification page (admin only)
- `/invitation/:token` - Accept invitation with **full TOTP setup flow** ⭐ CHANGED
- `/setup` - Tenant setup wizard (multi-step)
- `/account/totp` - **TOTP management** (regenerate backup codes, view status) ⭐ NEW
- `/recovery` - **Account recovery** (backup code login, request admin reset) ⭐ NEW

## Security Features

✅ **No password storage** (eliminates password breach risk)
✅ **TOTP secrets encrypted at rest** (AES-256-GCM)
✅ **Backup codes bcrypt-hashed** before storage
✅ Cryptographically secure tokens (UUID v4)
✅ Token hashing before storage (SHA-256)
✅ Single-use tokens and backup codes
✅ Time-limited tokens with expiry
✅ Rate limiting on all sensitive endpoints (TOTP, backup codes, recovery)
✅ **Constant-time TOTP comparison** (prevent timing attacks)
✅ Email/WhatsApp number enumeration prevention
✅ Secure session cookies (HttpOnly, Secure, SameSite)
✅ Session invalidation on TOTP reset
✅ **No password strength requirements** (TOTP is inherently strong)
✅ WhatsApp number masking in UI (privacy)
✅ WhatsApp Business Policy compliance
✅ Comprehensive audit logging for all recovery events

## WhatsApp Message Templates Required

1. **Operator Invitation** (with TOTP setup link)
2. **TOTP Reset by Admin** (with setup link)
3. **Account Recovery Instructions** (admin contact info)
4. **Backup Code Usage Alert** (optional, security notification)
5. **Welcome After Invitation Acceptance**

## Email Templates Required (Fallback Only)

1. Welcome & Verify Email (tenant admin signup - one-time)
2. Invitation (fallback when WhatsApp unavailable)
3. Recovery notifications (optional fallback)

## Success Metrics

- Operator onboarding completion in **< 2 minutes** ⚡ (was 3+ minutes)
- WhatsApp invitation delivery rate > 95%
- WhatsApp invitation acceptance rate > 70%
- **TOTP setup success rate > 90%** ⭐ NEW
- **Backup code usage success rate > 95%** ⭐ NEW
- **Zero password-related security incidents** ⭐ NEW
- **Reduction in support requests** (no password resets!)
- **Zero dependency on external email providers for operator onboarding**

## Next Steps

1. **Review the specs** - Read through proposal.md, design.md, and all spec files
2. **Prioritize** - Decide which phases to implement first (recommendation: start with Phase 1-3)
3. **Setup TOTP infrastructure** - Choose encryption key management (env var or KMS)
4. **Start implementation** - Follow tasks.md in order
5. **Test thoroughly** - Security (TOTP, encryption) and UX (QR code, backup codes) are critical

## Questions to Consider

Before starting implementation:

1. **TOTP encryption key**: Environment variable or cloud KMS (AWS KMS, GCP KMS, Azure Key Vault)?
2. **QR code library**: Which library for QR code generation? (recommended: `skip2/go-qrcode` or similar)
3. **Authenticator app recommendations**: Which apps to recommend? (Google Authenticator, Authy, 1Password, Bitwarden)
4. **Backup code format**: 8 alphanumeric characters in groups of 4? (e.g., `A7B9-C2D4`)
5. **TOTP parameters**: 6 digits, 30 seconds (standard) or 8 digits for higher security?
6. **Tenant ID UX**: Should we allow vanity slugs instead of UUIDs for user-facing use?
7. **Session policy**: Concurrent sessions or single session per user?
8. **Invitation expiry**: 7 days (default) or configurable per tenant?

## Files Created

```
openspec/changes/tenant-onboarding-flow/
├── proposal.md                    # High-level overview (TOTP-based)
├── design.md                      # Technical design & architecture (needs TOTP update)
├── tasks.md                       # Implementation tasks (7 phases, needs TOTP update)
├── README.md                      # This file (TOTP-based)
└── specs/
    ├── signup/
    │   └── spec.md                # Signup requirements (needs TOTP update)
    ├── signin/
    │   └── spec.md                # Authentication requirements (needs TOTP update)
    ├── email-verification/
    │   └── spec.md                # Email verification (admin only)
    ├── whatsapp-invitations/
    │   └── spec.md                # WhatsApp invitations with TOTP setup
    ├── totp-authentication/
    │   └── spec.md                # ⭐ NEW - Core TOTP authentication
    └── totp-reset-recovery/
        └── spec.md                # ⭐ NEW - Recovery without passwords
```

## Ready to Implement?

Start with:
1. Review all spec documents (especially new TOTP specs)
2. **Set up TOTP encryption key** (32-byte key for AES-256-GCM)
3. **Choose QR code generation library** for Go
4. Run Phase 1 database migrations (TOTP tables, backup codes)
5. Implement TOTP secret generation and encryption utilities
6. Implement WhatsApp message templates (invitation with TOTP setup)
7. Begin Phase 2 (signup flow with TOTP setup)

The specs are designed to be implementation-ready with clear acceptance criteria, scenarios, and technical details. **TOTP eliminates password complexity while providing superior security!** 🎉
