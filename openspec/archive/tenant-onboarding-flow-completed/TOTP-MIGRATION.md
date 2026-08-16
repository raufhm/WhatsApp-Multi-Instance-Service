# TOTP Migration Summary

This document summarizes the key changes needed to migrate from password-based to TOTP-based authentication.

## What Changed

### ✅ Core Authentication
- **Passwords eliminated** → TOTP codes from authenticator apps
- **Password reset flows** → Backup codes + admin TOTP reset
- **Password hashes in DB** → Encrypted TOTP secrets (AES-256-GCM)
- **"Forgot password"** → "Lost authenticator access?" (backup codes)

### ✅ User Experience
- **Login**: Tenant ID + email/WhatsApp + **6-digit TOTP code** (not password)
- **Signup**: Scan QR code → Verify TOTP → Download backup codes
- **Recovery**: Use backup code OR contact admin for TOTP reset
- **No more**: Password strength requirements, password reset emails

### ✅ Security Improvements
- No password breach risk (no passwords stored)
- TOTP secrets encrypted at rest
- Backup codes bcrypt-hashed, single-use
- Constant-time TOTP comparison (timing attack prevention)
- Rate limiting on TOTP and backup code endpoints

## Files Status

### ✅ Completed (TOTP-ready)
1. `proposal.md` - Updated with TOTP focus
2. `README.md` - Complete rewrite with TOTP details
3. `design.md` - Database schema updated with TOTP tables
4. `specs/totp-authentication/spec.md` - NEW (21KB core TOTP spec)
5. `specs/totp-reset-recovery/spec.md` - NEW (13KB recovery spec)
6. `specs/whatsapp-invitations/spec.md` - Updated with TOTP setup

### ⚠️ Needs Updates
1. `design.md` - API endpoints section (partial, needs completion)
2. `tasks.md` - All tasks need TOTP reframing
3. `specs/signin/spec.md` - Update for TOTP login (not password)
4. `specs/signup/spec.md` - Remove password, add TOTP setup
5. `specs/password-reset/spec.md` - Delete or redirect to totp-reset-recovery

## Database Migration Priority

### Phase 1: Core TOTP Tables (Week 1)
```sql
-- 1. TOTP secrets (encrypted)
CREATE TABLE totp_secrets (...);

-- 2. Backup codes (bcrypt-hashed)
CREATE TABLE totp_backup_codes (...);

-- 3. Recovery tokens (replaces password_reset_tokens)
CREATE TABLE totp_recovery_tokens (...);

-- 4. Update operators table
ALTER TABLE operators 
  ADD COLUMN totp_secret_encrypted TEXT,
  ADD COLUMN totp_verified_at TIMESTAMPTZ,
  ADD COLUMN totp_setup_required BOOLEAN DEFAULT FALSE,
  ADD COLUMN whatsapp_number TEXT UNIQUE,
  ALTER COLUMN email DROP NOT NULL;
```

### Phase 2: Invitation Tables (Week 1-2)
```sql
-- 1. Invitations with channel support
ALTER TABLE invitations 
  ADD COLUMN whatsapp_number TEXT,
  ADD COLUMN email_address TEXT,
  ADD COLUMN channel TEXT DEFAULT 'whatsapp';

-- 2. Delivery tracking
CREATE TABLE whatsapp_invitation_delivery (...);
```

## API Endpoint Changes

### Authentication (Changed)
| Old | New |
|-----|-----|
| `POST /login` (password) | `POST /login` (TOTP code) |
| `POST /password-reset/request` | `POST /recovery/request` (WhatsApp) |
| `POST /password-reset/confirm` | `POST /login/backup-code` |
| - | `POST /totp/verify-setup` |
| - | `POST /account/totp/regenerate-backup-codes` |

### Admin Endpoints (New)
- `POST /operators/:id/totp-reset` - Reset operator TOTP
- `GET /operators/:id/totp-status` - View TOTP status

## Frontend Page Changes

### New Pages
- `/account/totp` - TOTP management, backup codes
- `/recovery` - Backup code login, recovery request

### Changed Pages
- `/login` - TOTP code input (not password)
- `/invitation/:token` - Full TOTP setup flow with QR code
- `/signup/tenant` - Add TOTP setup after email verification

### Removed Pages
- `/forgot-password` → Replaced by `/recovery`
- `/reset-password/:token` → Replaced by TOTP reset flow

## Implementation Checklist

### Backend (Go)
- [ ] Add TOTP library dependency (`pquerna/otp` or similar)
- [ ] Add QR code library (`skip2/go-qrcode`)
- [ ] Implement TOTP secret generation (crypto/rand, base32)
- [ ] Implement AES-256-GCM encryption for TOTP secrets
- [ ] Create encryption key management (env var or KMS)
- [ ] Implement TOTP verification (constant-time comparison)
- [ ] Implement backup codes generation (10 codes, alphanumeric)
- [ ] Implement backup codes bcrypt hashing
- [ ] Update login handler to accept TOTP code
- [ ] Add backup code login endpoint
- [ ] Add TOTP reset by admin endpoint
- [ ] Add recovery request endpoint
- [ ] Update invitation endpoint to generate TOTP secret
- [ ] Add WhatsApp message templates (TOTP setup, reset, recovery)

### Frontend (React/TypeScript)
- [ ] Add TOTP input component (6 digits, countdown timer)
- [ ] Add QR code display component
- [ ] Add backup codes display/download component
- [ ] Update login page (TOTP instead of password)
- [ ] Create TOTP setup page (QR code + verification)
- [ ] Create backup codes page (download + acknowledgment)
- [ ] Create recovery page (backup code login + request reset)
- [ ] Add "Lost authenticator?" link on login
- [ ] Update invitation acceptance flow with TOTP setup

### Testing
- [ ] Unit tests: TOTP generation and verification
- [ ] Unit tests: Backup codes generation and validation
- [ ] Unit tests: TOTP encryption/decryption
- [ ] Integration tests: Full TOTP signup flow
- [ ] Integration tests: TOTP login flow
- [ ] Integration tests: Backup code login
- [ ] Integration tests: Admin TOTP reset
- [ ] E2E tests: Complete onboarding with TOTP
- [ ] Security tests: Rate limiting, timing attacks

### Documentation
- [ ] User guide: Setting up TOTP
- [ ] User guide: Using backup codes
- [ ] Admin guide: Resetting operator TOTP
- [ ] FAQ: Common TOTP issues
- [ ] Security documentation: TOTP implementation details

## Security Checklist

- [ ] TOTP secrets encrypted with AES-256-GCM
- [ ] Encryption key stored separately from database
- [ ] Constant-time comparison for TOTP verification
- [ ] Backup codes bcrypt-hashed before storage
- [ ] Rate limiting on TOTP endpoints (5 per 15 min)
- [ ] Rate limiting on backup codes (3 per 15 min)
- [ ] Recovery tokens expire after 1 hour
- [ ] Invitation tokens expire after 7 days
- [ ] Audit logging for all recovery events
- [ ] WhatsApp number masking in UI
- [ ] No sensitive data in logs

## Migration from Existing System (If Applicable)

If you have existing password-based operators:

### Phase 1: Parallel Operation (2-4 weeks)
- [ ] Support both password and TOTP login
- [ ] Add "Enable TOTP" option in account settings
- [ ] Encourage operators to enable TOTP
- [ ] Track TOTP adoption rate

### Phase 2: Mandatory TOTP (2 weeks)
- [ ] Require TOTP setup for all operators
- [ ] Set deadline for password deprecation
- [ ] Send reminders to operators without TOTP
- [ ] Provide admin TOTP reset for operators who lose access

### Phase 3: Password Removal (1 week)
- [ ] Disable password login entirely
- [ ] Remove password_hash column from operators table
- [ ] Remove password-related endpoints
- [ ] Update documentation (TOTP-only)

## Success Metrics

- **TOTP setup success rate** > 90%
- **Backup code generation rate** = 100% (all operators)
- **Backup code usage success rate** > 95%
- **Zero password-related security incidents**
- **Reduction in support requests** (no password resets)
- **Operator onboarding time** < 2 minutes

## Recommended Libraries (Go)

### TOTP Generation/Verification
```go
github.com/pquerna/otp  // RFC 6238 compliant
// or
github.com/xlzd/gotp   // Simple TOTP library
```

### QR Code Generation
```go
github.com/skip2/go-qrcode  // Pure Go QR code generator
```

### Encryption
```go
crypto/cipher  // Standard library (AES-GCM)
github.com/aws/aws-sdk-go-v2/service/kms  // If using AWS KMS
```

### Backup Codes
```go
crypto/rand  // Cryptographically secure random
golang.org/x/crypto/bcrypt  // Hashing backup codes
```

## Next Actions

1. **Review specs** - Read `totp-authentication/spec.md` and `totp-reset-recovery/spec.md`
2. **Choose libraries** - Select TOTP, QR code, encryption libraries
3. **Setup encryption key** - Generate 32-byte key for AES-256-GCM
4. **Run Phase 1 migrations** - Create TOTP tables
5. **Implement core TOTP** - Secret generation, encryption, verification
6. **Update login flow** - TOTP code input instead of password
7. **Test thoroughly** - Security, UX, edge cases
8. **Deploy incrementally** - Phase rollout if migrating existing users

---

**Questions?** Review the detailed specs in `specs/totp-authentication/` and `specs/totp-reset-recovery/`.
