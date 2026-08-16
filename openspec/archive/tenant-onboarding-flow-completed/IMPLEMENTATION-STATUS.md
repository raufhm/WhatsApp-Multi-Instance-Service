# ✅ TOTP-Based Onboarding Specification - COMPLETE

## Status: READY FOR IMPLEMENTATION

All specification documents have been updated to reflect **TOTP-based authentication** instead of passwords. The onboarding flow is now:
- **Passwordless**: No password storage, resets, or strength requirements
- **WhatsApp-first**: Operator invitations via WhatsApp (no email dependency)
- **TOTP-based**: Authentication via authenticator apps (Google Authenticator, Authy, etc.)
- **Secure recovery**: Backup codes + admin TOTP reset

---

## 📁 Complete File Inventory

### ✅ Core Documents (Updated for TOTP)

| File | Status | Size | Description |
|------|--------|------|-------------|
| `proposal.md` | ✅ Complete | 2KB | High-level overview with TOTP benefits |
| `README.md` | ✅ Complete | 14KB | Implementation summary with TOTP details |
| `design.md` | ✅ Complete | ~12KB | DB schema with TOTP tables |
| `TOTP-MIGRATION.md` | ✅ NEW | 8KB | Migration guide and implementation checklist |

### ✅ Specification Files (TOTP-Ready)

| File | Status | Size | Description |
|------|--------|------|-------------|
| `specs/signin/spec.md` | ✅ **Rewritten** | 17KB | TOTP login, backup codes, session management |
| `specs/signup/spec.md` | ✅ **Rewritten** | 17KB | TOTP setup with QR codes, backup codes generation |
| `specs/totp-authentication/spec.md` | ✅ NEW | 21KB | Core TOTP implementation details |
| `specs/totp-reset-recovery/spec.md` | ✅ NEW | 13KB | Recovery flows (backup codes, admin reset) |
| `specs/whatsapp-invitations/spec.md` | ✅ Updated | 17KB | WhatsApp invitations with TOTP setup |
| `specs/email-verification/spec.md` | ⚠️ Legacy | 7KB | Admin-only (can be removed or minimized) |

### 🗑️ Files to Remove

| File | Reason |
|------|--------|
| `specs/password-reset/spec.md` | Replaced by `totp-reset-recovery/spec.md` |

---

## 🎯 Key Features Summary

### Authentication Flow (TOTP-Based)

**Login:**
1. Enter tenant ID (UUID)
2. Enter email OR WhatsApp number
3. **Enter 6-digit TOTP code** from authenticator app
4. ✅ Logged in (session created)

**Recovery (Lost Authenticator):**
1. Click "Lost access to authenticator?"
2. Enter backup code (8 characters)
3. ✅ Logged in (regenerate codes required)

**Recovery (No Backup Codes):**
1. Contact admin
2. Admin resets TOTP
3. WhatsApp sent with setup link
4. Complete TOTP setup again

### Operator Onboarding (WhatsApp Invitation)

1. Admin sends WhatsApp invitation
2. Operator clicks link
3. Enter name (email optional)
4. **Scan QR code with authenticator app**
5. **Enter TOTP code to verify**
6. **Download 10 backup codes**
7. ✅ Account activated, auto-logged in

### Tenant Admin Onboarding

1. Sign up with email + tenant details
2. Verify email (one-time)
3. **Set up TOTP with QR code**
4. **Download backup codes**
5. Complete setup wizard
6. ✅ Dashboard access

---

## 🔐 Security Features

✅ **No password storage** (eliminates breach risk)  
✅ **TOTP secrets encrypted** (AES-256-GCM)  
✅ **Backup codes bcrypt-hashed** (single-use)  
✅ **Constant-time comparison** (timing attack prevention)  
✅ **Rate limiting** on all auth endpoints  
✅ **Session management** (8h default, 30d remember-me)  
✅ **Audit logging** for all recovery events  
✅ **Email enumeration prevention**  
✅ **WhatsApp number masking** in UI  

---

## 📊 Database Schema (TOTP Tables)

### New Tables
- ✅ `totp_secrets` - Encrypted TOTP secrets per operator
- ✅ `totp_backup_codes` - Hashed single-use backup codes
- ✅ `totp_recovery_tokens` - Recovery tokens (1-hour expiry)
- ✅ `whatsapp_invitation_delivery` - WhatsApp delivery tracking
- ✅ `recovery_audit_log` - All recovery events audit

### Modified Tables
- ✅ `operators` - Add `whatsapp_number`, `totp_secret_encrypted`, `totp_verified_at`
- ✅ `invitations` - Add `whatsapp_number`, `channel` (whatsapp/email/manual)
- ✅ `tenants` - Add setup tracking fields

---

## 🚀 Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [x] Specs complete
- [ ] DB migrations (TOTP tables, backup codes)
- [ ] TOTP library integration (`pquerna/otp`)
- [ ] QR code library (`skip2/go-qrcode`)
- [ ] Encryption utilities (AES-256-GCM)
- [ ] WhatsApp message templates (TOTP setup, reset)

### Phase 2: TOTP Signup (Week 2-3)
- [ ] TOTP secret generation endpoint
- [ ] QR code generation endpoint
- [ ] TOTP verification endpoint
- [ ] Backup codes generation
- [ ] WhatsApp invitation with TOTP
- [ ] Frontend TOTP setup UI

### Phase 3: TOTP Login (Week 3-4)
- [ ] TOTP login endpoint
- [ ] Backup code login endpoint
- [ ] Session management
- [ ] Rate limiting
- [ ] Frontend login UI (TOTP input)
- [ ] "Lost authenticator?" flow

### Phase 4: Recovery (Week 4-5)
- [ ] Admin TOTP reset endpoint
- [ ] Recovery request endpoint
- [ ] WhatsApp notifications
- [ ] Audit logging
- [ ] Frontend recovery UI

### Phase 5-7: Polish & Launch (Week 5-8)
- [ ] Testing (unit, integration, E2E)
- [ ] Security audit
- [ ] Documentation
- [ ] Deployment
- [ ] Monitoring

---

## 📚 Recommended Go Libraries

### TOTP Generation/Verification
```go
github.com/pquerna/otp  // RFC 6238 compliant, well-maintained
// or
github.com/xlzd/gotp   // Simpler, lighter
```

### QR Code Generation
```go
github.com/skip2/go-qrcode  // Pure Go, no dependencies
```

### Encryption (AES-256-GCM)
```go
crypto/cipher              // Standard library
crypto/aes                 // Standard library
github.com/aws/aws-sdk-go-v2/service/kms  // If using AWS KMS
```

### Backup Codes
```go
crypto/rand                // Cryptographically secure random
golang.org/x/crypto/bcrypt // Hashing backup codes
```

### WhatsApp Integration
```go
// Already using: github.com/tulir/whatsmeow
// Add template system for invitation/reset messages
```

---

## 🎯 Success Metrics

| Metric | Target |
|--------|--------|
| Operator onboarding time | < 2 minutes |
| TOTP setup success rate | > 90% |
| WhatsApp invitation delivery | > 95% |
| Invitation acceptance rate | > 70% |
| Backup code generation | 100% (all operators) |
| Backup code usage success | > 95% |
| Password-related incidents | 0 (eliminated!) |
| Support requests (auth) | -50% (no password resets) |

---

## 📖 Next Steps for Implementation

### 1. Review Specs (Day 1)
- [ ] Read `README.md` for overview
- [ ] Read `TOTP-MIGRATION.md` for checklist
- [ ] Review `specs/totp-authentication/spec.md` (core TOTP)
- [ ] Review `specs/signin/spec.md` (login flow)
- [ ] Review `specs/signup/spec.md` (signup with TOTP)

### 2. Setup Infrastructure (Day 2-3)
- [ ] Choose TOTP library (`pquerna/otp` recommended)
- [ ] Choose QR code library (`skip2/go-qrcode`)
- [ ] Generate TOTP encryption key (32 bytes)
- [ ] Setup key management (env var or KMS)
- [ ] Add dependencies to `go.mod`

### 3. Run Phase 1 Migrations (Day 4)
- [ ] Create `totp_secrets` table
- [ ] Create `totp_backup_codes` table
- [ ] Create `totp_recovery_tokens` table
- [ ] Update `operators` table (add TOTP fields)
- [ ] Create `invitations` table (with WhatsApp support)
- [ ] Create `whatsapp_invitation_delivery` table

### 4. Implement Core TOTP (Day 5-7)
- [ ] TOTP secret generation function
- [ ] TOTP encryption/decryption utilities
- [ ] TOTP verification function (constant-time)
- [ ] QR code generation endpoint
- [ ] Backup codes generation function
- [ ] Backup codes bcrypt hashing

### 5. Continue with Phase 2-7
Follow the detailed tasks in `tasks.md` (needs TOTP update) or `TOTP-MIGRATION.md`.

---

## ❓ Questions or Need Help?

Refer to:
- **Spec details**: `specs/totp-authentication/spec.md`
- **Implementation guide**: `TOTP-MIGRATION.md`
- **User flows**: `README.md` (Key Features section)
- **DB schema**: `design.md` (Database Schema Changes)

---

## 🎉 Summary

**You now have a complete, production-ready specification for TOTP-based authentication!**

✅ No passwords (more secure, less complexity)  
✅ WhatsApp-first invitations (no email vendor lock-in)  
✅ TOTP via authenticator apps (industry standard)  
✅ Backup codes for recovery (simple, secure)  
✅ Admin TOTP reset (recovery without email)  
✅ Comprehensive specs (ready for implementation)  

**Start with Phase 1 and you'll have a modern, secure authentication system in 6-8 weeks!** 🚀
