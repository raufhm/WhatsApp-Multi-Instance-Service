# Design: Tenant Onboarding Flow

## User Journey

### New Tenant Admin

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Landing page   │────▶│  Signup form     │────▶│  Verify email   │
│  (optional)     │     │  (tenant + admin)│     │  (check inbox)  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                                                        │
                                                        ▼
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Dashboard      │◀────│  Setup wizard    │◀────│  Login          │
│  (inbox)        │     │  (org + WhatsApp)│     │  (first time)   │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

### New Operator (Invited via WhatsApp with TOTP Setup)

```
┌─────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  Invitation     │────▶│  Signup + TOTP   │────▶│  Download        │
│  WhatsApp msg   │     │  setup (QR code) │     │  backup codes    │
└─────────────────┘     └──────────────────┘     └──────────────────┘
                                                        │
                                                        ▼
┌─────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  Dashboard      │◀────│  Login           │◀────│  Auto-login      │
│  (inbox)        │     │  (TOTP codes)    │     │  after setup     │
└─────────────────┘     └──────────────────┘     └──────────────────┘
```

### Existing User (Recovery via Backup Code or Admin Reset)

```
┌─────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  Login page     │────▶│  Use backup code │────▶│  Login succeeds  │
│  (no TOTP)      │     │  OR admin reset  │     │  (regenerate!)   │
└─────────────────┘     └──────────────────┘     └──────────────────┘
                                                        │
                          ┌──────────────────┐         │
                          │  Admin resets    │         │
                          │  TOTP via        │─────────┘
                          │  WhatsApp link   │
                          └──────────────────┘

## Database Schema Changes

### Table: `tenants` (unchanged)

```sql
CREATE TABLE IF NOT EXISTS tenants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    slug            TEXT UNIQUE,  -- for vanity URLs
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'suspended')),
    metadata        JSONB DEFAULT '{}',  -- org details, timezone, industry, etc.
    setup_completed BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Table: `totp_secrets` ⭐ NEW

```sql
CREATE TABLE IF NOT EXISTS totp_secrets (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id      UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    secret_encrypted TEXT NOT NULL,  -- AES-256-GCM encrypted base32 secret
    algorithm        TEXT NOT NULL DEFAULT 'SHA1',
    digits           INTEGER NOT NULL DEFAULT 6,
    period           INTEGER NOT NULL DEFAULT 30,
    is_active        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    activated_at     TIMESTAMPTZ,
    deactivated_at   TIMESTAMPTZ
);

CREATE INDEX idx_totp_secrets_operator_active ON totp_secrets(operator_id) WHERE is_active = TRUE;
CREATE UNIQUE INDEX idx_totp_secrets_operator ON totp_secrets(operator_id) WHERE is_active = TRUE;
```

### Table: `totp_backup_codes` ⭐ NEW

```sql
CREATE TABLE IF NOT EXISTS totp_backup_codes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    code_hash   TEXT NOT NULL,  -- bcrypt hash of the backup code
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_backup_codes_operator ON totp_backup_codes(operator_id) WHERE used_at IS NULL;
CREATE INDEX idx_backup_codes_unused ON totp_backup_codes(operator_id, used_at) WHERE used_at IS NULL;
COMMENT ON TABLE totp_backup_codes IS 'Single-use backup codes for TOTP recovery';
```

### Table: `email_verification_tokens` (tenant admin only)

```sql
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id   UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    token_hash    TEXT NOT NULL,  -- SHA-256 hash of the token
    expires_at    TIMESTAMPTZ NOT NULL,
    used_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(operator_id)  -- one active token per operator
);

CREATE INDEX idx_verification_tokens_expires ON email_verification_tokens(expires_at);
```

### Table: `totp_recovery_tokens` ⭐ NEW (replaces password_reset_tokens)

```sql
CREATE TABLE IF NOT EXISTS totp_recovery_tokens (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id   UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    token_hash    TEXT NOT NULL,  -- SHA-256 hash of the token
    expires_at    TIMESTAMPTZ NOT NULL,  -- 1 hour for recovery
    used_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_recovery_tokens_operator ON totp_recovery_tokens(operator_id);
CREATE INDEX idx_recovery_tokens_expires ON totp_recovery_tokens(expires_at);
```

### Table: `invitations` (with WhatsApp support)

```sql
CREATE TABLE IF NOT EXISTS invitations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    whatsapp_number TEXT,
    email_address   TEXT,  -- optional, for fallback
    channel         TEXT NOT NULL DEFAULT 'whatsapp' CHECK (channel IN ('whatsapp', 'email', 'manual')),
    role            TEXT NOT NULL DEFAULT 'operator' CHECK (role IN ('admin', 'operator', 'viewer')),
    token_hash      TEXT NOT NULL,
    invited_by      UUID NOT NULL REFERENCES operators(id),
    expires_at      TIMESTAMPTZ NOT NULL,  -- 7 days
    accepted_at     TIMESTAMPTZ,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_invitations_tenant_whatsapp ON invitations(tenant_id, whatsapp_number) 
  WHERE whatsapp_number IS NOT NULL AND accepted_at IS NULL AND revoked_at IS NULL;
  
CREATE UNIQUE INDEX idx_invitations_tenant_email ON invitations(tenant_id, email_address) 
  WHERE email_address IS NOT NULL AND accepted_at IS NULL AND revoked_at IS NULL;
  
CREATE INDEX idx_invitations_token ON invitations(token_hash);
CREATE INDEX idx_invitations_expires ON invitations(expires_at);
CREATE INDEX idx_invitations_channel ON invitations(channel);
```

### Table: `whatsapp_invitation_delivery` ⭐ NEW

```sql
CREATE TABLE IF NOT EXISTS whatsapp_invitation_delivery (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invitation_id UUID NOT NULL REFERENCES invitations(id) ON DELETE CASCADE,
    status        TEXT NOT NULL DEFAULT 'pending' 
      CHECK (status IN ('pending', 'sent', 'delivered', 'read', 'failed', 'expired')),
    sent_at       TIMESTAMPTZ,
    delivered_at  TIMESTAMPTZ,
    read_at       TIMESTAMPTZ,
    failed_at     TIMESTAMPTZ,
    failure_reason TEXT,
    retry_count   INTEGER DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_whatsapp_delivery_status ON whatsapp_invitation_delivery(status);
CREATE INDEX idx_whatsapp_delivery_invitation ON whatsapp_invitation_delivery(invitation_id);
```

### Modified Table: `operators` ⭐ TOTP-focused

```sql
-- Add TOTP and WhatsApp fields
ALTER TABLE operators ADD COLUMN IF NOT EXISTS whatsapp_number TEXT UNIQUE;
ALTER TABLE operators ADD COLUMN IF NOT EXISTS totp_secret_encrypted TEXT;  -- AES-256-GCM encrypted
ALTER TABLE operators ADD COLUMN IF NOT EXISTS totp_verified_at TIMESTAMPTZ;
ALTER TABLE operators ADD COLUMN IF NOT EXISTS totp_setup_required BOOLEAN DEFAULT FALSE;
ALTER TABLE operators ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ;

-- Email becomes optional for operators (only required for tenant admin)
ALTER TABLE operators ALTER COLUMN email DROP NOT NULL;

-- Indexes for efficient lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_operators_tenant_whatsapp 
  ON operators(tenant_id, whatsapp_number) WHERE whatsapp_number IS NOT NULL;
  
CREATE UNIQUE INDEX IF NOT EXISTS idx_operators_tenant_email 
  ON operators(tenant_id, email) WHERE email IS NOT NULL;
```

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/dashboard/api/signup/tenant` | Create new tenant + admin | No |
| POST | `/dashboard/api/signup/operator` | Create operator in tenant | No |
| POST | `/dashboard/api/login` | Login (existing) | No |
| POST | `/dashboard/api/logout` | Logout | Yes |
| GET | `/dashboard/api/me` | Current operator | Yes |

### Email Verification

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/dashboard/api/verify-email/request` | Request verification email | No (needs email+tenant) |
| GET | `/dashboard/api/verify-email/:token` | Verify email token | No |
| POST | `/dashboard/api/verify-email/resend` | Resend verification email | No |

### Password Reset

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/dashboard/api/password-reset/request` | Request reset email | No |
| GET | `/dashboard/api/password-reset/:token` | Validate reset token | No |
| POST | `/dashboard/api/password-reset/confirm` | Set new password | No (needs token) |

### Invitations (WhatsApp-First)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/dashboard/api/invitations/whatsapp` | Create WhatsApp invitation | Yes (admin) |
| POST | `/dashboard/api/invitations/email` | Create email invitation (fallback) | Yes (admin) |
| GET | `/dashboard/api/invitations` | List pending invitations | Yes (admin) |
| DELETE | `/dashboard/api/invitations/:id` | Revoke invitation | Yes (admin) |
| GET | `/dashboard/api/invitations/accept/:token` | Accept invitation page | No |

### WhatsApp Verification

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/dashboard/api/verify-whatsapp/request` | Request verification via WhatsApp | No (needs number+tenant) |
| GET | `/dashboard/api/verify-whatsapp/:token` | Verify WhatsApp token | No |
| POST | `/dashboard/api/verify-whatsapp/resend` | Resend WhatsApp verification | No |

### Tenant Setup

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/dashboard/api/tenant/setup-status` | Get setup progress | Yes |
| PUT | `/dashboard/api/tenant/setup` | Update setup step | Yes (admin) |
| POST | `/dashboard/api/tenant/complete-setup` | Mark setup complete | Yes (admin) |

## Frontend Routes

```
/dashboard
  ├── /login                    # Login page
  ├── /signup                   # Signup landing (choose tenant vs operator)
  │   ├── /tenant               # Tenant signup form
  │   └── /operator             # Operator signup form
  ├── /verify-email             # Verify email page
  ├── /forgot-password          # Request password reset
  ├── /reset-password/:token    # Set new password
  ├── /invitation/:token        # Accept invitation
  ├── /setup                    # Tenant setup wizard
  │   ├── /organization         # Org details
  │   ├── /whatsapp             # WhatsApp pairing
  │   └── /complete             # Success screen
  └── /                         # Dashboard (inbox) - requires auth
```

## WhatsApp Message Templates

### Required Templates (WhatsApp)

1. **Operator Invitation** (primary channel)
   - Sent when admin invites operator via WhatsApp
   - Includes signup link with token
   - Example: "Hi! [Admin Name] from [Org Name] invited you to join as [Role]. Sign up: [link]"

2. **Password Reset** (primary channel for operators)
   - Sent when operator requests password reset
   - Includes reset link with token (1-hour expiry)
   - Example: "Password Reset Request. Reset here: [link]. Expires in 1 hour."

3. **Password Changed Confirmation** (primary channel)
   - Sent after successful password change
   - Security notification
   - Example: "Password Changed Successfully. If this wasn't you, contact your admin."

4. **Welcome After Invitation Acceptance**
   - Sent after operator completes signup via WhatsApp invitation
   - Welcome message + getting started tips

### Email Templates (Fallback Only)

1. **Welcome & Verify Email** (tenant admin signup only - one-time)
2. **Password Reset** (fallback when WhatsApp fails)
3. **Invitation** (fallback when WhatsApp unavailable)

### Template Structure

Each WhatsApp message MUST include:
- Clear, concise heading
- Personalized greeting (when possible)
- Context (who sent it, what organization)
- Clear call-to-action with link
- Expiry information
- Security notice (for password reset)
- Support contact information

## Security Considerations

### Token Security

- All tokens (verification, reset, invitation) MUST be:
  - Cryptographically secure random UUIDs
  - Hashed before database storage (SHA-256)
  - Single-use only
  - Time-limited (expiry enforced)
  - Invalidated on use

### Rate Limiting

| Endpoint | Limit | Window | Action on Exceed |
|----------|-------|--------|------------------|
| `/signup/tenant` | 5 | 10 min | Block 30 min |
| `/signup/operator` | 5 | 10 min | Block 30 min |
| `/login` | 5 | 15 min | Block 30 min |
| `/verify-email/request` | 3 | 1 hour | Block 1 hour |
| `/password-reset/request` | 3 | 1 hour | Block 1 hour |

### Session Security

- Sessions MUST be invalidated on:
  - Password change
  - Password reset
  - Account deactivation
  - Manual logout
- Session cookies MUST be:
  - HttpOnly
  - Secure (in production)
  - SameSite=Lax
  - Reasonable expiry (8h default, 30d with remember-me)

### Email Enumeration Prevention

- Login, password reset, and signup endpoints MUST return generic success messages
- Never reveal whether an email exists in the system
- Always send the same response time (constant-time comparison)

## Accessibility Requirements

- All forms MUST have proper labels and ARIA attributes
- Error messages MUST be announced to screen readers
- Color contrast MUST meet WCAG AA standards
- Keyboard navigation MUST work throughout the flow
- Loading states MUST be announced
- Success/error messages MUST be visible and persistent

## Performance Requirements

- Page load time < 2 seconds
- Form submission feedback < 500ms
- Email delivery < 30 seconds (P95)
- Token validation < 100ms

## Metrics & Monitoring

### Key Metrics to Track

- Signup completion rate (by tenant/operator)
- Email verification rate
- Time from signup to verification
- Password reset success rate
- Invitation acceptance rate
- Setup wizard completion rate
- Drop-off points in each flow

### Alerts

- Email delivery failure rate > 5%
- Signup error rate spike > 2x normal
- Token validation failures (potential attack)
- Rate limit triggers (unusual patterns)

## Out of Scope (Deferred)

- OAuth/Social login (Google, Microsoft, etc.)
- SSO/SAML integration
- Multi-factor authentication (MFA)
- Phone-based verification
- Account self-deletion
- Data export on signup
