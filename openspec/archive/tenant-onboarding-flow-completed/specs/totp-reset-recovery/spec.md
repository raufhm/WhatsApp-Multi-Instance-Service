# TOTP Reset and Recovery

## Design Principle: No Passwords, Only TOTP Recovery

Since the system uses **TOTP-based authentication** (no passwords), recovery flows focus on:
- **Backup codes**: Primary self-service recovery method
- **TOTP reset by admin**: When backup codes are lost
- **WhatsApp notifications**: Alert operators of recovery events

---

## ADDED Requirements

### Requirement: Backup codes are the primary recovery method

Operators MUST use backup codes for self-service recovery when they lose authenticator access.

#### Scenario: Login with backup code

- **GIVEN** an operator has lost authenticator access
- **WHEN** they visit the login page
- **THEN** they MUST see:
  - Standard login fields (tenant ID, email/WhatsApp number)
  - TOTP code field (default)
  - "Lost access to authenticator?" link
- **AND** clicking the link:
  - Shows backup code input field
  - Hides TOTP code field
  - Displays brief instructions

#### Scenario: Backup code login succeeds

- **GIVEN** a valid, unused backup code
- **WHEN** the operator submits it
- **THEN** the system MUST:
  - Validate the code (bcrypt comparison)
  - Log in the operator
  - Invalidate that backup code (mark as used)
  - Show warning: "You have X backup codes remaining"
  - Strongly suggest regenerating backup codes
  - Redirect to account settings to regenerate codes

#### Scenario: Backup code login fails

- **WHEN** an invalid or used backup code is submitted
- **THEN** the system MUST:
  - Return generic error: "Invalid credentials"
  - NOT reveal whether the code was close or already used
  - Rate limit further attempts (3 per 15 minutes)
  - Log the failed attempt for security monitoring

#### Scenario: Regenerate backup codes

- **GIVEN** an operator logged in with backup code OR has few codes remaining
- **WHEN** they visit account settings
- **THEN** they MUST see:
  - "Backup Codes" section
  - Number of remaining codes
  - "Generate new backup codes" button
- **AND** clicking the button:
  - Generates 10 new backup codes
  - Invalidates all existing codes
  - Displays new codes (downloadable)
  - Requires acknowledgment: "I have saved these codes"

### Requirement: Admin can reset TOTP for operators

Tenant admins MUST be able to reset TOTP when operators lose all access (no backup codes).

#### Scenario: Admin initiates TOTP reset

- **GIVEN** a tenant admin
- **WHEN** they view an operator's profile
- **THEN** they MUST see a "Reset TOTP" action (in operator management)
- **AND** clicking it:
  - Shows confirmation dialog
  - Warns: "This will invalidate the operator's current TOTP and backup codes"
  - Requires admin to confirm
  - Optionally: require admin to provide a reason
  - On confirm:
    - Invalidate operator's current TOTP secret
    - Invalidate all backup codes
    - Generate new TOTP secret (unverified)
    - Send WhatsApp message to operator with reset link
    - Mark operator account as "TOTP setup required"
    - Log the reset event (admin ID, timestamp, reason)

#### Scenario: TOTP reset WhatsApp notification

- **WHEN** an admin resets an operator's TOTP
- **THEN** the system MUST send a WhatsApp message:
  - Example:
    ```
    TOTP Reset by Admin
    
    [Admin Name] has reset your two-factor authentication for WhatsApp Operator Dashboard.
    
    Complete TOTP setup here: https://your-domain.com/dashboard/totp/setup/:token
    
    You must complete setup within 24 hours to regain access.
    
    If this wasn't authorized, contact your admin immediately.
    ```

#### Scenario: Operator completes TOTP reset

- **GIVEN** an operator whose TOTP was reset by admin
- **WHEN** they click the reset link
- **THEN** the system MUST:
  - Validate the token (24-hour expiry)
  - Show TOTP setup page (same as invitation flow)
  - Display QR code for new TOTP secret
  - Require TOTP verification
  - Generate new backup codes
  - Require backup code acknowledgment
  - Clear "TOTP setup required" status
  - Log in the operator automatically
  - Redirect to dashboard

#### Scenario: TOTP reset token expires

- **GIVEN** a TOTP reset token older than 24 hours
- **WHEN** the operator attempts to use it
- **THEN** the system MUST:
  - Reject with "Reset link expired"
  - Instruct operator to contact their admin
  - NOT allow login without TOTP setup
  - Admin must initiate a new reset

### Requirement: Recovery via WhatsApp (no backup codes, no admin access)

Operators who lose both authenticator and backup codes MUST have a recovery path.

#### Scenario: Request account recovery via WhatsApp

- **GIVEN** an operator with no backup codes and no authenticator access
- **WHEN** they visit the login page
- **THEN** they MUST see:
  - "Lost access to authenticator and backup codes?" link
  - Clicking it shows:
    - Tenant ID field
    - WhatsApp number field
    - "Send recovery instructions" button

#### Scenario: Recovery WhatsApp message sent

- **GIVEN** a valid WhatsApp number and tenant ID
- **WHEN** recovery is requested
- **THEN** the system MUST:
  - Look up operator by WhatsApp number + tenant
  - Check if operator exists and is active
  - Generate a recovery token (UUID, 1-hour expiry)
  - Send WhatsApp message with recovery link
  - Return generic success (no enumeration)
  - Example message:
    ```
    Account Recovery Request
    
    You requested to recover access to your WhatsApp Operator Dashboard account.
    
    Recover here: https://your-domain.com/dashboard/recovery/:token
    
    This link expires in 1 hour.
    
    If you didn't request this, contact your admin immediately.
    ```

#### Scenario: Recovery link leads to admin contact

- **GIVEN** a valid recovery token
- **WHEN** the operator clicks the recovery link
- **THEN** the system MUST:
  - Validate the token
  - Show message: "Account recovery requires admin approval"
  - Display:
    - Admin contact information (email, WhatsApp)
    - Operator's account details (masked for security)
    - Instructions: "Contact your admin to reset TOTP"
  - Optionally: send notification to all tenant admins
    - "Operator [name] requested account recovery"

### Requirement: Recovery rate limiting and security

Recovery endpoints MUST be heavily rate-limited to prevent abuse.

#### Scenario: Backup code rate limiting

- **WHEN** backup code attempts exceed 3 failed attempts in 15 minutes
- **THEN** the system MUST:
  - Block further backup code attempts for 60 minutes
  - Return generic error: "Too many failed attempts. Try again later."
  - Log the event for security review
  - Optionally notify operator via WhatsApp (if configured)

#### Scenario: TOTP reset by admin rate limiting

- **WHEN** an admin resets TOTP more than 10 times in 1 hour
- **THEN** the system MUST:
  - Block further resets for 1 hour
  - Notify super-admin or system owner
  - Log for security review (potential abuse)

#### Scenario: Recovery request rate limiting

- **WHEN** recovery requests exceed 3 per hour for the same WhatsApp number
- **THEN** the system MUST:
  - Block further requests for 24 hours
  - Return generic success (no enumeration)
  - Log for security review

### Requirement: Recovery audit logging

All recovery events MUST be logged for security and compliance.

#### Scenario: Backup code usage audit log

- **WHEN** a backup code is used
- **THEN** the system MUST log:
  - Operator ID (masked in logs)
  - Timestamp
  - IP address
  - User agent / device info
  - Which backup code was used (index, not the code itself)
  - Remaining backup codes count

#### Scenario: Admin TOTP reset audit log

- **WHEN** an admin resets an operator's TOTP
- **THEN** the system MUST log:
  - Admin ID and email
  - Operator ID and email/WhatsApp number
  - Timestamp
  - IP address of admin
  - Reason provided (if any)
  - Session ID of admin

#### Scenario: Recovery request audit log

- **WHEN** account recovery is requested
- **THEN** the system MUST log:
  - WhatsApp number (masked)
  - Tenant ID
  - Timestamp
  - IP address
  - Whether operator was found
  - Whether WhatsApp message was sent

### Requirement: Recovery UI/UX

Recovery flows MUST be clear, reassuring, and actionable.

#### Scenario: Backup code login UI

- **WHEN** operator clicks "Lost access to authenticator?"
- **THEN** they MUST see:
  - Clear heading: "Use a backup code"
  - Instructions: "Enter one of your 10 backup codes"
  - Input field (8 characters, alphanumeric)
  - Format-agnostic input (accepts with or without hyphens)
  - "Back to TOTP login" link
  - Help text: "Backup codes were provided when you set up two-factor authentication"

#### Scenario: Out of backup codes warning

- **GIVEN** an operator with 0-2 backup codes remaining
- **WHEN** they log in successfully
- **THEN** the system MUST show:
  - Prominent warning banner: "You have X backup codes remaining"
  - "Generate new codes now" button (prominent)
  - "Remind me later" option (but show on every login until resolved)
  - Optional: require regeneration if 0 codes remain

#### Scenario: Admin TOTP reset confirmation

- **WHEN** an admin resets an operator's TOTP
- **THEN** they MUST see:
  - Success confirmation
  - "Operator has been sent recovery instructions via WhatsApp"
  - "Operator must complete setup within 24 hours"
  - "You will be notified when they complete setup" (optional)
  - Link to view operator status

---

## Database Schema

### Modified Table: `operators`

```sql
-- Already has totp_secret_encrypted from TOTP spec
-- Add fields for recovery tracking

ALTER TABLE operators ADD COLUMN totp_setup_required BOOLEAN DEFAULT FALSE;  -- After admin reset
ALTER operators ADD COLUMN last_backup_code_used_at TIMESTAMPTZ;
```

### New Table: `totp_recovery_tokens`

```sql
CREATE TABLE IF NOT EXISTS totp_recovery_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id     UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL,  -- SHA-256 hash
    expires_at      TIMESTAMPTZ NOT NULL,  -- 1 hour from creation
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_recovery_tokens_expires ON totp_recovery_tokens(expires_at);
CREATE INDEX idx_recovery_tokens_operator ON totp_recovery_tokens(operator_id);
```

### New Table: `recovery_audit_log`

```sql
CREATE TABLE IF NOT EXISTS recovery_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id     UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    event_type      TEXT NOT NULL CHECK (event_type IN ('backup_code_used', 'totp_reset_by_admin', 'recovery_requested')),
    actor_id        UUID,  -- Admin ID if applicable
    details         JSONB,  -- event-specific details
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_recovery_audit_operator ON recovery_audit_log(operator_id);
CREATE INDEX idx_recovery_audit_type ON recovery_audit_log(event_type);
CREATE INDEX idx_recovery_audit_created ON recovery_audit_log(created_at);
```

## API Endpoints

### Backup Codes

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/dashboard/api/login/backup-code` | Login with backup code | No |
| GET | `/dashboard/api/account/backup-codes` | Get remaining codes count | Yes |
| POST | `/dashboard/api/account/totp/regenerate-backup-codes` | Generate new backup codes | Yes |

### TOTP Reset (Admin)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/dashboard/api/operators/:id/totp-reset` | Reset operator TOTP | Yes (admin) |
| GET | `/dashboard/api/operators/:id/totp-status` | Get TOTP status | Yes (admin) |

### Account Recovery

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/dashboard/api/recovery/request` | Request recovery (WhatsApp) | No |
| GET | `/dashboard/api/recovery/:token` | Recovery instructions page | No |

---

## Security Considerations

### Backup Code Entropy

- 8 alphanumeric characters = 52 bits of entropy
- 10 codes per operator = sufficient recovery options
- Single-use only (deleted after use)
- Bcrypt-hashed before storage

### Rate Limiting

| Endpoint | Limit | Window | Action on Exceed |
|----------|-------|--------|------------------|
| `/login/backup-code` | 3 | 15 min | Block 60 min |
| `/operators/:id/totp-reset` | 10 | 1 hour | Block 1 hour, notify super-admin |
| `/recovery/request` | 3 | 1 hour | Block 24 hours |

### Admin Permissions

- Only admins can reset operator TOTP
- Super-admin can reset admin TOTP
- All TOTP resets require audit logging
- Optional: require 2FA for admin performing TOTP reset

### No Self-Service TOTP Reset Without Backup Codes

- Operators cannot reset their own TOTP without current code
- Lost authenticator + no backup codes = admin reset required
- This prevents account takeover attacks

---

## Out of Scope (Deferred)

- SMS-based recovery codes
- Email-based recovery (except tenant admin)
- Automated identity verification for recovery
- Recovery contact persons (trusted contacts)
- Time-delayed recovery (security waiting period)
