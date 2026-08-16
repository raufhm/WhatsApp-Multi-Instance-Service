# 🚀 TOTP Onboarding - Quick Start Guide

**For developers implementing the TOTP-based authentication system.**

---

## TL;DR: What Changed?

| Before (Passwords) | After (TOTP) |
|-------------------|--------------|
| Login: email + password | Login: email + **6-digit TOTP code** |
| Password hashes in DB | **Encrypted TOTP secrets** (AES-256-GCM) |
| "Forgot password" email | **Backup codes** (10 single-use) |
| Password strength rules | **No passwords** (TOTP is secure by default) |
| Password reset flow | **Admin TOTP reset** via WhatsApp |

---

## 5-Minute Overview

### 1. How Users Login

```
┌─────────────────────────────────────┐
│  Login                              │
│                                     │
│  Tenant ID: [550e8400-...]         │
│  Email/WhatsApp: [operator@...]    │
│  TOTP Code: [123 456] ← from app   │
│                                     │
│  [Lost access to authenticator?]   │
│  [Sign In]                          │
└─────────────────────────────────────┘
```

### 2. How Users Sign Up (via WhatsApp invitation)

```
1. Admin sends WhatsApp invitation
2. Operator clicks link
3. Enter name (email optional)
4. ┌──────────────────────────┐
   │ SCAN QR CODE            │
   │ with Google Authenticator│
   │ or Authy or 1Password   │
   └──────────────────────────┘
5. Enter 6-digit code to verify
6. Download 10 backup codes
7. Done! Auto-logged in.
```

### 3. How Recovery Works

**Scenario A: Has backup codes**
- Click "Lost access to authenticator?"
- Enter backup code (e.g., `A7B9-C2D4`)
- ✅ Logged in

**Scenario B: No backup codes**
- Contact admin
- Admin resets TOTP
- WhatsApp sent with setup link
- Complete TOTP setup again
- ✅ Access restored

---

## Database Changes (Critical)

### Add to `operators` table:
```sql
ALTER TABLE operators 
  ADD COLUMN whatsapp_number TEXT UNIQUE,
  ADD COLUMN totp_secret_encrypted TEXT,  -- AES-256-GCM encrypted base32
  ADD COLUMN totp_verified_at TIMESTAMPTZ,
  ADD COLUMN totp_setup_required BOOLEAN DEFAULT FALSE;

ALTER TABLE operators ALTER COLUMN email DROP NOT NULL;
```

### Create `totp_backup_codes` table:
```sql
CREATE TABLE totp_backup_codes (
    id          UUID PRIMARY KEY,
    operator_id UUID NOT NULL REFERENCES operators(id),
    code_hash   TEXT NOT NULL,  -- bcrypt hash
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_backup_codes_unused 
  ON totp_backup_codes(operator_id, used_at) WHERE used_at IS NULL;
```

---

## Code Examples (Go)

### Generate TOTP Secret
```go
import (
    "github.com/pquerna/otp"
    "github.com/pquerna/otp/totp"
)

// Generate secret
secret, err := otp.GenerateSecret(&totp.GenerateOpts{
    Issuer:      "WhatsApp Operator Dashboard",
    AccountName: "operator@example.com",
})

// Encrypt with AES-256-GCM before storing
encrypted := encryptAES256GCM(secret, encryptionKey)
```

### Verify TOTP Code
```go
import "github.com/pquerna/otp/totp"

// Decrypt secret from DB
secret := decryptAES256GCM(encryptedSecret, encryptionKey)

// Verify code (allows ±1 period for clock drift)
valid := totp.Validate(code, secret)

if !valid {
    return errors.New("invalid TOTP code")
}
```

### Generate QR Code
```go
import "github.com/skip2/go-qrcode"

// Generate otpauth:// URI
otpauth := fmt.Sprintf(
    "otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
    "WhatsApp%20Operator%20Dashboard",
    "operator@example.com",
    secret,
    "WhatsApp%20Operator%20Dashboard",
)

// Generate QR code PNG
png, err := qrcode.Encode(otpauth, qrcode.Medium, 256)
```

### Generate Backup Codes
```go
import (
    "crypto/rand"
    "golang.org/x/crypto/bcrypt"
)

func generateBackupCodes() ([]string, []string, error) {
    var plaintext []string
    var hashed []string
    
    for i := 0; i < 10; i++ {
        // Generate 8 alphanumeric characters
        code := generateRandomCode(8)
        plaintext = append(plaintext, code)
        
        // Hash with bcrypt
        hash, err := bcrypt.GenerateFromPassword([]byte(code), 12)
        if err != nil {
            return nil, nil, err
        }
        hashed = append(hashed, string(hash))
    }
    
    return plaintext, hashed, nil
}
```

---

## Frontend Changes (React/TypeScript)

### Login Page Component
```tsx
// Before: Password input
<Input type="password" value={password} />

// After: TOTP code input
<Input 
  type="text" 
  inputMode="numeric" 
  pattern="\d{6}"
  maxLength={6}
  value={totpCode}
  placeholder="123 456"
/>
<CountdownTimer period={30} />
<Link to="/recovery">Lost access to authenticator?</Link>
```

### TOTP Setup Component
```tsx
// New component for signup flow
<TOTPSetup
  qrCode={qrCodePNG}
  secret="JBSW Y3DP EHPK 3PXP"
  onVerify={(code) => verifyTOTP(code)}
  onComplete={(backupCodes) => downloadBackupCodes(backupCodes)}
/>
```

---

## Security Checklist (Must-Have)

- [ ] TOTP secrets encrypted with AES-256-GCM
- [ ] Encryption key in env var or KMS (NOT in code)
- [ ] Constant-time comparison for TOTP verification
- [ ] Backup codes bcrypt-hashed before storage
- [ ] Rate limiting: 5 TOTP attempts per 15 min
- [ ] Rate limiting: 3 backup code attempts per 15 min
- [ ] No sensitive data in logs (no TOTP codes, no secrets)
- [ ] Session cookies: HttpOnly, Secure, SameSite=Lax

---

## Common Pitfalls to Avoid

❌ **Don't** store TOTP secrets in plaintext  
✅ **Do** encrypt with AES-256-GCM before storage

❌ **Don't** use `time.Now()` directly for TOTP (timezone issues)  
✅ **Do** use UTC consistently

❌ **Don't** log TOTP codes or secrets  
✅ **Do** log only success/failure with masked identifiers

❌ **Don't** allow unlimited TOTP attempts  
✅ **Do** rate limit (5 per 15 min)

❌ **Don't** store backup codes in plaintext  
✅ **Do** bcrypt-hash before storage

❌ **Don't** use sequential random numbers  
✅ **Do** use `crypto/rand` for cryptographic randomness

---

## Testing Checklist

```bash
# Unit tests
go test ./totp/...      # Secret generation, encryption, verification
go test ./backup/...    # Backup codes generation, hashing

# Integration tests
go test ./handler/...   # Login endpoint, signup endpoint
go test ./storage/...   # DB operations for TOTP secrets, backup codes

# E2E tests (Cypress/Playwright)
cy.run('totp-signup.cy.js')    # Full signup with TOTP setup
cy.run('totp-login.cy.js')     # Login with TOTP code
cy.run('backup-code.cy.js')    # Recovery with backup code
```

---

## Environment Variables

```bash
# TOTP Configuration
TOTP_ENCRYPTION_KEY=<32-byte-hex-key>  # Required
TOTP_ISSUER="WhatsApp Operator Dashboard"
TOTP_DIGITS=6
TOTP_PERIOD=30

# Optional
TOTP_BACKUP_CODES_COUNT=10
TOTP_BACKUP_CODES_LENGTH=8
```

Generate encryption key:
```bash
openssl rand -hex 32
# Output: 64 hex characters (32 bytes)
```

---

## Files Reference

| File | Purpose |
|------|---------|
| `README.md` | Overview and implementation plan |
| `TOTP-MIGRATION.md` | Detailed migration checklist |
| `IMPLEMENTATION-STATUS.md` | Complete status and next steps |
| `design.md` | Database schema and architecture |
| `specs/totp-authentication/spec.md` | Core TOTP requirements |
| `specs/totp-reset-recovery/spec.md` | Recovery flows |
| `specs/signin/spec.md` | Login requirements |
| `specs/signup/spec.md` | Signup with TOTP setup |

---

## Need Help?

1. **Spec questions** → Read `specs/totp-authentication/spec.md`
2. **Implementation** → Follow `TOTP-MIGRATION.md` checklist
3. **Database** → See `design.md` for schema
4. **Code examples** → Check this guide + specs

---

**You're ready to implement! Start with Phase 1 in `TOTP-MIGRATION.md`.** 🚀
