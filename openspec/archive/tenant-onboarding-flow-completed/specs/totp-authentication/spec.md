# TOTP-Based Authentication

## Design Principle: Passwordless with TOTP

The system MUST use **TOTP (Time-based One-Time Password)** as the primary authentication method for all operators:
- **No passwords**: Eliminates password storage, breaches, and reset flows
- **Authenticator app agnostic**: Works with Google Authenticator, Authy, 1Password, Bitwarden, etc.
- **TOTP during onboarding**: Setup TOTP when accepting invitation
- **Backup codes**: Recovery mechanism for lost authenticator access
- **TOTP for tenant admin**: Required during initial tenant setup

---

## ADDED Requirements

### Requirement: TOTP secret generation and storage

The system MUST generate and securely store TOTP secrets for each operator.

#### Scenario: Generate TOTP secret for new operator

- **GIVEN** an operator account is being created (via invitation)
- **WHEN** the invitation is created
- **THEN** the system MUST:
  - Generate a cryptographically secure random TOTP secret (20 bytes, base32 encoded)
  - Encrypt the secret before storage (AES-256-GCM)
  - Store the encrypted secret in the database
  - Associate the secret with the operator account
  - Mark the secret as "not yet verified" (pending setup)

#### Scenario: TOTP secret storage encryption

- **WHEN** a TOTP secret is stored
- **THEN** it MUST be encrypted using:
  - Algorithm: AES-256-GCM
  - Key: Derived from a master key (environment variable)
  - Unique nonce/IV per secret
  - Authentication tag for integrity
- **AND** the encryption key MUST be stored separately from the database
  - Environment variable: `TOTP_ENCRYPTION_KEY` (32 bytes)
  - Or cloud KMS (AWS KMS, GCP KMS, Azure Key Vault)

#### Scenario: TOTP configuration parameters

- **WHEN** TOTP secrets are generated
- **THEN** they MUST use standard parameters:
  - Algorithm: SHA-1 (RFC 6238 standard)
  - Digits: 6 digits (standard) or 8 digits (optional, more secure)
  - Period: 30 seconds (standard)
  - Skew: ±1 period (allow for clock drift)
  - Issuer: "WhatsApp Operator Dashboard"
  - Label format: `issuer:operator-email` or `issuer:whatsapp-number`

### Requirement: QR code generation for TOTP setup

The system MUST generate QR codes for easy TOTP setup in authenticator apps.

#### Scenario: Generate TOTP setup QR code

- **GIVEN** an operator accepting an invitation
- **WHEN** they visit the invitation acceptance page
- **THEN** the system MUST:
  - Generate an `otpauth://` URI with:
    - Type: `totp`
    - Issuer: Organization name (tenant)
    - Account: Operator email or WhatsApp number
    - Secret: Base32-encoded TOTP secret
    - Algorithm: `SHA1`
    - Digits: `6`
    - Period: `30`
  - Example URI:
    ```
    otpauth://totp/WhatsApp%20Operator%20Dashboard:operator@example.com?secret=JBSWY3DPEHPK3PXP&issuer=WhatsApp%20Operator%20Dashboard&algorithm=SHA1&digits=6&period=30
    ```
  - Generate a QR code encoding the URI
  - Display the QR code prominently
  - Provide manual entry instructions (secret key, parameters)

#### Scenario: QR code display requirements

- **WHEN** the QR code is displayed
- **THEN** it MUST:
  - Be at least 200x200 pixels (scannable size)
  - Have high contrast (black on white)
  - Include the issuer logo in the center (optional, for branding)
  - Be accompanied by:
    - Step-by-step setup instructions
    - Manual entry key (base32 secret, formatted in groups)
    - Link to download authenticator apps (Google Authenticator, Authy, etc.)
  - Expire after first successful TOTP verification (one-time setup)

#### Scenario: Manual TOTP entry fallback

- **GIVEN** an operator cannot scan the QR code
- **WHEN** they choose manual entry
- **THEN** the system MUST display:
  - Secret key in base32, formatted in groups of 4 characters
  - Example: `JBSW Y3DP EHPK 3PXP`
  - Account name/email
  - Issuer name
  - Algorithm: SHA1
  - Digits: 6
  - Period: 30 seconds

### Requirement: TOTP verification during login

Operators MUST authenticate using TOTP codes from their authenticator app.

#### Scenario: Login with TOTP

- **GIVEN** an operator with a verified TOTP secret
- **WHEN** they visit the login page
- **THEN** they MUST:
  - Enter tenant ID (or select from remembered tenants)
  - Enter email or WhatsApp number
  - Enter current 6-digit TOTP code from authenticator app
  - Submit the form

#### Scenario: TOTP code validation

- **GIVEN** a login attempt with TOTP code
- **WHEN** the system validates the code
- **THEN** it MUST:
  - Retrieve the operator's encrypted TOTP secret
  - Decrypt the secret
  - Generate expected TOTP code for current time period
  - Check current period code
  - Check previous period code (skew +1, for clock drift)
  - Check next period code (skew +1, for clock drift)
  - Use constant-time comparison (prevent timing attacks)
  - Accept if any of the 3 codes match
  - Reject if no match

#### Scenario: TOTP login succeeds

- **WHEN** a valid TOTP code is provided
- **THEN** the system MUST:
  - Create a session record
  - Set session cookie (HttpOnly, Secure, SameSite)
  - Redirect to dashboard
  - Log the successful login (timestamp, IP, device info)

#### Scenario: TOTP login fails

- **WHEN** an invalid TOTP code is provided
- **THEN** the system MUST:
  - Return generic error: "Invalid credentials" (no enumeration)
  - NOT reveal whether email/number exists
  - NOT reveal whether the code was close to valid
  - Rate limit further attempts (see rate limiting section)
  - Log the failed attempt for security monitoring

### Requirement: TOTP setup during invitation acceptance

Operators MUST set up TOTP when accepting an invitation.

#### Scenario: Invitation acceptance flow with TOTP

- **GIVEN** an operator clicking a WhatsApp invitation link
- **WHEN** they visit the invitation acceptance page
- **THEN** the system MUST:
  - Validate the invitation token
  - Pre-fill operator details (WhatsApp number, role, tenant)
  - Require: name, email (optional)
  - Display TOTP setup QR code
  - Require TOTP verification before account activation
  - Flow:
    1. Operator enters name
    2. System generates TOTP secret and displays QR code
    3. Operator scans QR code with authenticator app
    4. Operator enters current TOTP code to verify setup
    5. System generates backup codes
    6. Operator downloads/saves backup codes
    7. Account is activated
    8. Operator is logged in automatically

#### Scenario: TOTP setup verification

- **GIVEN** an operator has scanned the QR code
- **WHEN** they submit a TOTP code
- **THEN** the system MUST:
  - Validate the code against the generated secret
  - If valid: mark TOTP as "verified" and proceed
  - If invalid: show error, allow retry
  - If 5 failed attempts: invalidate the secret, require restart

#### Scenario: Backup codes generation

- **GIVEN** a successful TOTP setup verification
- **WHEN** the operator completes setup
- **THEN** the system MUST:
  - Generate 10 single-use backup codes
  - Each code: 8 alphanumeric characters (e.g., `A7B9-C2D4`)
  - Store hashed backup codes (bcrypt, like passwords)
  - Display codes to operator in a downloadable format
  - Require operator to acknowledge: "I have saved these backup codes"
  - Show warning: "These codes can only be used once. Store them safely."
  - Example display:
    ```
    Backup Codes for WhatsApp Operator Dashboard
    Account: operator@example.com
    
    A7B9-C2D4    E8F1-G3H5    I9J2-K4L6
    M7N9-P1Q3    R5S7-T9U1    V3W5-X7Y9
    Z1A3-B5C7    D9E1-F3G5    H7I9-J1K3
    L5M7-N9P1
    
    Save these codes in a safe place. Each code can only be used once.
    If you lose access to your authenticator app, these codes are your only way to log in.
    ```

### Requirement: Backup codes for recovery

Backup codes MUST provide recovery when operators lose authenticator access.

#### Scenario: Login with backup code

- **GIVEN** an operator has lost authenticator access
- **WHEN** they visit the login page
- **THEN** they MUST:
  - Click "Use backup code instead" link
  - Enter tenant ID
  - Enter email or WhatsApp number
  - Enter a backup code (8 characters, with or without hyphen)
  - System validates the code
  - If valid: log in, invalidate that backup code, prompt to regenerate

#### Scenario: Backup code validation

- **WHEN** a backup code is submitted
- **THEN** the system MUST:
  - Retrieve operator's hashed backup codes
  - Hash the submitted code (bcrypt)
  - Compare against stored hashes
  - If match: accept, invalidate that code (delete from database)
  - If no match: reject with generic error
  - Rate limit backup code attempts (stricter than TOTP)

#### Scenario: Regenerate backup codes

- **GIVEN** an operator logged in with backup code OR has low backup codes remaining
- **WHEN** they visit account settings
- **THEN** the system MUST:
  - Show remaining backup codes count
  - Provide "Generate new backup codes" button
  - When clicked:
    - Generate 10 new backup codes
    - Invalidate all existing backup codes
    - Display new codes (same format as initial setup)
    - Require acknowledgment: "I have saved these codes"

#### Scenario: Out of backup codes warning

- **GIVEN** an operator with 0 backup codes remaining
- **WHEN** they log in successfully with TOTP
- **THEN** the system MUST:
  - Show a prominent warning: "You have no backup codes remaining"
  - Provide link to regenerate codes immediately
  - Optionally require regeneration before continuing (configurable)

### Requirement: TOTP reset by admin

Tenant admins MUST be able to reset TOTP for operators who lose access.

#### Scenario: Admin resets operator TOTP

- **GIVEN** a tenant admin
- **WHEN** they view an operator's profile
- **THEN** they MUST see a "Reset TOTP" action
- **AND** clicking it:
  - Invalidates the current TOTP secret
  - Generates a new TOTP secret
  - Sends a WhatsApp message to the operator:
    - "Your TOTP has been reset by [Admin Name]"
    - New setup link with QR code
    - "Complete setup within 24 hours or contact your admin"
  - Marks the operator account as "TOTP pending setup"
  - Operator must complete TOTP setup on next login

#### Scenario: Operator completes TOTP reset

- **GIVEN** an operator whose TOTP was reset by admin
- **WHEN** they log in
- **THEN** the system MUST:
  - Detect "TOTP pending setup" status
  - Redirect to TOTP setup page (same as invitation flow)
  - Require QR code scan and verification
  - Generate new backup codes
  - Clear "TOTP pending setup" status

#### Scenario: TOTP reset audit log

- **WHEN** an admin resets an operator's TOTP
- **THEN** the system MUST log:
  - Admin ID who performed the reset
  - Operator ID affected
  - Timestamp
  - IP address of admin
  - Reason (optional, admin can provide)

### Requirement: TOTP rate limiting and security

TOTP endpoints MUST be rate-limited to prevent brute force attacks.

#### Scenario: TOTP login rate limiting

- **WHEN** an email/number exceeds 5 failed TOTP attempts in 15 minutes
- **THEN** the system MUST:
  - Block further TOTP attempts for 30 minutes
  - Return generic error: "Too many failed attempts. Try again later."
  - Log the event for security monitoring
  - Optionally notify the operator via WhatsApp (if configured)

#### Scenario: Backup code rate limiting

- **WHEN** backup code attempts exceed 3 failed attempts in 15 minutes
- **THEN** the system MUST:
  - Block further backup code attempts for 60 minutes
  - Return generic error: "Too many failed attempts. Try again later."
  - Log the event for security monitoring
  - Notify the operator via WhatsApp: "Suspicious backup code attempts detected"

#### Scenario: TOTP secret enumeration prevention

- **WHEN** a login attempt fails
- **THEN** the system MUST:
  - Return the same generic error regardless of failure reason:
    - Invalid email/number
    - Invalid TOTP code
    - Account not found
  - Use constant-time comparison for all cryptographic operations
  - Maintain consistent response time (prevent timing attacks)

### Requirement: Multi-device TOTP support

Operators MAY set up TOTP on multiple devices.

#### Scenario: Enable multi-device TOTP

- **GIVEN** an operator with verified TOTP
- **WHEN** they visit account settings
- **THEN** they MUST see an option: "Add another authenticator device"
- **AND** clicking it:
  - Displays the same QR code (same secret)
  - Allows scanning on multiple devices
  - All devices generate the same codes
  - No need to regenerate secret

#### Scenario: Multi-device considerations

- **WHEN** an operator uses multiple devices
- **THEN** the system MUST:
  - Use the same TOTP secret for all devices
  - All devices generate identical codes
  - Operator is responsible for managing their own devices
  - If secret is reset, all devices must be reconfigured

### Requirement: TOTP setup UI/UX

The TOTP setup and login experience MUST be intuitive and accessible.

#### Scenario: TOTP setup instructions

- **WHEN** an operator sees the TOTP setup page
- **THEN** they MUST see:
  - Clear heading: "Set up two-factor authentication"
  - Step-by-step instructions:
    1. "Install an authenticator app (if you don't have one)"
       - Links to: Google Authenticator, Authy, 1Password, Bitwarden
    2. "Scan the QR code with your app"
       - QR code displayed prominently
    3. "Or enter the code manually"
       - Manual secret key displayed
    4. "Enter the 6-digit code from your app"
       - Input field for verification code
    5. "Save your backup codes"
       - Downloadable backup codes
  - Visual aids (icons, diagrams)
  - Link to help documentation

#### Scenario: TOTP login UI

- **WHEN** an operator visits the login page
- **THEN** they MUST see:
  - Tenant ID field (or tenant selector if remembered)
  - Email or WhatsApp number field
  - TOTP code field (6 digits, numeric keyboard on mobile)
  - "Use backup code instead" link
  - Clear error messages
  - Loading state during verification
  - Auto-focus on appropriate field

#### Scenario: TOTP code input UX

- **WHEN** entering a TOTP code
- **THEN** the input MUST:
  - Accept only numeric input (0-9)
  - Auto-format as 6 digits (or 8 if configured)
  - Show a countdown timer (seconds until code expires)
  - Auto-submit when 6 digits are entered (optional)
  - Clear on invalid code, allow immediate retry

#### Scenario: Accessibility requirements

- **WHEN** TOTP setup/login is used with assistive technology
- **THEN** the system MUST:
  - Provide screen reader announcements for QR code
  - Describe manual entry key clearly
  - Announce success/error states
  - Support keyboard-only navigation
  - Meet WCAG AA contrast requirements
  - Provide text alternatives for all visual elements

### Requirement: TOTP migration and rollover

The system MUST support TOTP secret rotation and algorithm upgrades.

#### Scenario: TOTP secret rotation

- **GIVEN** security best practices
- **WHEN** an operator requests TOTP reset (or admin initiates)
- **THEN** the system MUST:
  - Generate a new secret
  - Invalidate the old secret immediately
  - Require operator to scan new QR code
  - Generate new backup codes
  - Log the rotation event

#### Scenario: Algorithm upgrade path

- **GIVEN** future security requirements (e.g., SHA-256 instead of SHA-1)
- **WHEN** the system needs to upgrade TOTP algorithm
- **THEN** the system MUST:
  - Support multiple algorithms simultaneously
  - Mark secrets with their algorithm version
  - Gradually migrate operators on next TOTP reset
  - Maintain backward compatibility during transition

---

## Database Schema Changes

### Modified Table: `operators`

```sql
ALTER TABLE operators ADD COLUMN totp_secret_encrypted TEXT;  -- Encrypted TOTP secret (base64)
ALTER TABLE operators ADD COLUMN totp_verified_at TIMESTAMPTZ;  -- When TOTP was verified
ALTER TABLE operators ADD COLUMN totp_required BOOLEAN DEFAULT TRUE;  -- Whether TOTP is required
```

### New Table: `totp_backup_codes`

```sql
CREATE TABLE IF NOT EXISTS totp_backup_codes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id     UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    code_hash       TEXT NOT NULL,  -- Bcrypt hash of the backup code
    used_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_backup_codes_operator ON totp_backup_codes(operator_id) WHERE used_at IS NULL;
CREATE INDEX idx_backup_codes_unused ON totp_backup_codes(operator_id, used_at) WHERE used_at IS NULL;
```

### New Table: `totp_secrets` (optional, for audit/rotation history)

```sql
CREATE TABLE IF NOT EXISTS totp_secrets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operator_id     UUID NOT NULL REFERENCES operators(id) ON DELETE CASCADE,
    secret_encrypted TEXT NOT NULL,  -- Encrypted TOTP secret
    algorithm       TEXT NOT NULL DEFAULT 'SHA1',
    digits          INTEGER NOT NULL DEFAULT 6,
    period          INTEGER NOT NULL DEFAULT 30,
    is_active       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    activated_at    TIMESTAMPTZ,
    deactivated_at  TIMESTAMPTZ,
    deactivated_reason TEXT
);

CREATE INDEX idx_totp_secrets_operator_active ON totp_secrets(operator_id) WHERE is_active = TRUE;
```

## API Endpoints

### TOTP Setup and Verification

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/dashboard/api/totp/setup/:token` | Get TOTP setup info (QR code, secret) | No (invitation token) |
| POST | `/dashboard/api/totp/verify-setup` | Verify TOTP setup code | No (needs token + code) |
| POST | `/dashboard/api/totp/generate-backup-codes` | Generate new backup codes | Yes |

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/dashboard/api/login` | Login with TOTP code | No |
| POST | `/dashboard/api/login/backup-code` | Login with backup code | No |
| POST | `/dashboard/api/logout` | Logout | Yes |
| GET | `/dashboard/api/me` | Current operator | Yes |

### Account Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/dashboard/api/account/totp` | Get TOTP status (setup, backup codes remaining) | Yes |
| POST | `/dashboard/api/account/totp/reset` | Reset own TOTP (requires current code) | Yes |
| GET | `/dashboard/api/account/backup-codes` | View remaining backup codes count | Yes |
| POST | `/dashboard/api/account/totp/regenerate-backup-codes` | Regenerate backup codes | Yes |

### Admin TOTP Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/dashboard/api/operators/:id/totp-reset` | Reset operator TOTP | Yes (admin) |
| GET | `/dashboard/api/operators/:id/totp-status` | Get operator TOTP status | Yes (admin) |

## Security Considerations

### TOTP Secret Encryption

- **Encryption algorithm**: AES-256-GCM
- **Key management**: 
  - Master key via environment variable: `TOTP_ENCRYPTION_KEY`
  - Or cloud KMS (recommended for production)
  - Key rotation support
- **Storage**: Encrypted secrets in database, key separate

### Rate Limiting

| Endpoint | Limit | Window | Action on Exceed |
|----------|-------|--------|------------------|
| `/login` (TOTP) | 5 | 15 min | Block 30 min |
| `/login/backup-code` | 3 | 15 min | Block 60 min |
| `/totp/verify-setup` | 5 | 10 min | Invalidate setup, require restart |
| `/operators/:id/totp-reset` | 10 | 1 hour | Require 2FA for admin |

### Timing Attack Prevention

- All TOTP comparisons MUST use constant-time algorithms
- Response times MUST be consistent regardless of failure point
- Network-level padding to obscure processing time

### Backup Code Security

- Backup codes MUST be bcrypt-hashed before storage
- Each code is single-use (deleted after use)
- Codes are 8 alphanumeric characters (52 bits of entropy)
- 10 codes per operator provides sufficient recovery options

### Session Security

- TOTP verification creates a session (cookie-based)
- Sessions expire after 8 hours (configurable)
- "Remember me" extends to 30 days
- Sessions invalidated on TOTP reset

## Migration from Password-Based Auth

If migrating from existing password-based system:

### Phase 1: Parallel Operation

- Support both password and TOTP login
- Encourage operators to enable TOTP
- Show "Enable TOTP" prompt in dashboard

### Phase 2: Mandatory TOTP

- Require TOTP setup for all operators
- Set deadline for password deprecation
- Provide admin TOTP reset for operators who lose access

### Phase 3: Password Deprecation

- Disable password login entirely
- Remove password hashes from database
- TOTP-only authentication

---

## Out of Scope (Deferred)

- WebAuthn/FIDO2 support (hardware keys, biometrics)
- Push-based authentication (Duo Push, Microsoft Authenticator push)
- SMS-based 2FA fallback
- Hardware token support (YubiKey OTP mode)
- TOTP code length customization per operator (8 digits)
- Custom TOTP periods (non-standard 30s)
