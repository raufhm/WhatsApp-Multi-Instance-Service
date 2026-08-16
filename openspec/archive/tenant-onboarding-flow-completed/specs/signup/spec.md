# Tenant and Operator Signup with TOTP

## Design Principle: TOTP-Based, Passwordless Onboarding

All operator accounts MUST use **TOTP (Time-based One-Time Password)** authentication:
- **No passwords**: Eliminate password storage, breaches, and reset flows
- **TOTP during signup**: QR code scan + verification before account activation
- **Backup codes**: 10 single-use codes generated and downloaded during setup
- **Email minimal**: Only tenant admin requires email verification; operators use WhatsApp

---

## ADDED Requirements

### Requirement: Tenant admin signup with TOTP

New tenant admins MUST sign up with email verification and TOTP setup.

#### Scenario: Tenant admin signup succeeds

- **GIVEN** a valid organization email and tenant details
- **WHEN** the tenant admin submits the signup form
- **THEN** the system MUST:
  - Create a tenant record (status: pending)
  - Create admin operator record (TOTP setup required)
  - Generate email verification token
  - Send verification email to admin
  - Redirect to "verify your email" page
  - After verification: require TOTP setup before first login

#### Scenario: Tenant admin TOTP setup after email verification

- **GIVEN** a tenant admin has verified their email
- **WHEN** they complete email verification
- **THEN** the system MUST:
  - Generate TOTP secret (encrypted with AES-256-GCM)
  - Display QR code for authenticator app
  - Provide manual entry option (base32 secret)
  - Require TOTP code verification
  - Generate 10 backup codes
  - Require backup codes acknowledgment
  - Mark TOTP as verified
  - Activate tenant account
  - Log admin in automatically

#### Scenario: Duplicate tenant signup is rejected

- **GIVEN** an email already associated with a tenant
- **WHEN** the same email attempts to sign up again
- **THEN** the system MUST:
  - Reject with "This email is already registered"
  - Suggest login or TOTP reset if access is lost
  - NOT reveal whether tenant exists (if email not found)

#### Scenario: Tenant ID is generated automatically

- **WHEN** a tenant is created
- **THEN** the system MUST:
  - Generate UUID v4 as tenant ID
  - Store UUID for future API calls
  - NOT expose UUID in public URLs (use slugs if needed)
  - Generate unique slug if vanity URLs are enabled

### Requirement: Operator signup via WhatsApp invitation (Primary)

Operators MUST sign up through WhatsApp invitations with TOTP setup.

#### Scenario: Operator accepts WhatsApp invitation

- **GIVEN** a valid, unexpired WhatsApp invitation
- **WHEN** the operator clicks the WhatsApp link
- **THEN** the system MUST:
  - Validate invitation token (7-day expiry)
  - Pre-fill signup form with:
    - WhatsApp number (from invitation)
    - Role (from invitation)
    - Tenant context (from invitation)
  - Require only: name, email (optional)
  - Skip email verification (WhatsApp proves identity)
  - **Generate TOTP secret** (encrypted with AES-256-GCM)
  - **Display QR code** for authenticator app setup
  - **Provide manual entry** option (formatted base32 secret)
  - **Require TOTP verification** (enter current 6-digit code)
  - **Generate 10 backup codes** (8 alphanumeric characters each)
  - **Display backup codes** in downloadable format
  - **Require acknowledgment**: "I have saved these backup codes"
  - Create operator account with verified TOTP
  - Mark invitation as accepted
  - Log operator in automatically
  - Redirect to dashboard or setup wizard

#### Scenario: WhatsApp invitation to existing number is rejected

- **GIVEN** a WhatsApp number already exists as operator in tenant
- **WHEN** admin tries to invite that number
- **THEN** the system MUST:
  - Reject with "This WhatsApp number is already registered"
  - Show existing operator's name and role
  - Offer to resend invitation or modify existing operator
  - NOT create duplicate account

#### Scenario: Invalid WhatsApp invitation

- **GIVEN** an expired or revoked invitation
- **WHEN** operator clicks the link
- **THEN** the system MUST:
  - Reject with "This invitation has expired or been revoked"
  - Instruct operator to contact their admin
  - Provide admin contact information
  - NOT allow signup without valid invitation

### Requirement: Operator signup via email invitation (Fallback)

Email-based operator signup MUST remain available as fallback.

#### Scenario: Operator accepts email invitation

- **GIVEN** a valid email invitation
- **WHEN** the operator clicks the email link
- **THEN** the system MUST:
  - Validate invitation token (7-day expiry)
  - Pre-fill signup form with:
    - Email address (from invitation)
    - Role (from invitation)
    - Tenant context (from invitation)
  - Require: name, WhatsApp number
  - **Require email verification** (send verification email)
  - After email verification: proceed to TOTP setup
  - Generate TOTP secret and display QR code
  - Require TOTP verification
  - Generate and display backup codes
  - Require backup codes acknowledgment
  - Create operator account with verified TOTP
  - Mark invitation as accepted
  - Log operator in automatically

#### Scenario: Email invitation requires email verification

- **GIVEN** an operator signing up via email invitation
- **WHEN** they submit the signup form
- **THEN** the system MUST:
  - Generate email verification token
  - Send verification email
  - Display "Check your email" page
  - After verification: proceed to TOTP setup
  - NOT create operator account until email is verified

### Requirement: TOTP setup during signup

TOTP setup MUST be completed during signup before account activation.

#### Scenario: TOTP secret generation

- **WHEN** operator signup reaches TOTP setup step
- **THEN** the system MUST:
  - Generate cryptographically secure random secret (20 bytes)
  - Encode secret in base32 format
  - Encrypt secret with AES-256-GCM
  - Store encrypted secret in database
  - Mark secret as "pending verification"
  - Use standard TOTP parameters:
    - Algorithm: SHA-1
    - Digits: 6
    - Period: 30 seconds
    - Issuer: "WhatsApp Operator Dashboard"
    - Label: operator email or WhatsApp number

#### Scenario: QR code display for TOTP setup

- **WHEN** TOTP secret is generated
- **THEN** the system MUST:
  - Generate `otpauth://` URI:
    ```
    otpauth://totp/WhatsApp%20Operator%20Dashboard:operator@example.com?secret=JBSWY3DPEHPK3PXP&issuer=WhatsApp%20Operator%20Dashboard&algorithm=SHA1&digits=6&period=30
    ```
  - Generate QR code encoding the URI
  - Display QR code prominently (minimum 200x200 pixels)
  - Ensure high contrast (black on white)
  - Include issuer logo in center (optional, for branding)
  - Provide step-by-step scanning instructions
  - Auto-expire QR code after first successful verification

#### Scenario: Manual TOTP entry fallback

- **GIVEN** operator cannot scan QR code
- **WHEN** they choose manual entry option
- **THEN** the system MUST display:
  - Secret key in base32, formatted in groups of 4
  - Example: `JBSW Y3DP EHPK 3PXP`
  - Account name/email
  - Issuer name: "WhatsApp Operator Dashboard"
  - Algorithm: SHA1
  - Digits: 6
  - Period: 30 seconds
  - Copy-to-clipboard button for secret

#### Scenario: TOTP verification during setup

- **GIVEN** operator has scanned QR code or entered manually
- **WHEN** they submit a TOTP code
- **THEN** the system MUST:
  - Decrypt the TOTP secret
  - Generate expected code for current time period
  - Check current period ±1 (clock drift tolerance)
  - Use constant-time comparison
  - If valid: mark TOTP as "verified", proceed to backup codes
  - If invalid: show error, allow retry
  - If 5 failed attempts: invalidate secret, require restart
  - Log failed attempts for security monitoring

#### Scenario: TOTP setup failure handling

- **WHEN** TOTP setup fails (5 invalid attempts)
- **THEN** the system MUST:
  - Invalidate the TOTP secret
  - Display error: "Too many failed attempts. Please restart setup."
  - Offer to generate new TOTP secret
  - Invalidate the old secret completely
  - Log the security event
  - Notify admin if pattern suggests attack (optional)

### Requirement: Backup codes generation and download

Backup codes MUST be generated during signup for recovery.

#### Scenario: Backup codes generation

- **GIVEN** successful TOTP verification during signup
- **WHEN** operator proceeds to next step
- **THEN** the system MUST:
  - Generate 10 single-use backup codes
  - Each code: 8 alphanumeric characters (uppercase + digits)
  - Format: groups of 4 with hyphen (e.g., `A7B9-C2D4`)
  - Cryptographically secure random generation
  - Bcrypt-hash each code before storage
  - Store hashed codes in `totp_backup_codes` table
  - Display codes to operator in clear text (one-time only)

#### Scenario: Backup codes display

- **WHEN** backup codes are generated
- **THEN** the system MUST:
  - Display codes in a clear, scannable format
  - Organize in grid (e.g., 5 columns x 2 rows)
  - Include example: `A7B9-C2D4`
  - Provide "Download" button (TXT or PDF)
  - Provide "Print" option (optional)
  - Include account info (email/WhatsApp number)
  - Include instructions: "Save these codes in a safe place"
  - Include warning: "Each code can only be used once"
  - Include tenant name and support contact

#### Scenario: Backup codes acknowledgment

- **GIVEN** backup codes are displayed
- **WHEN** operator views the codes
- **THEN** the system MUST:
  - Require checkbox: "I have saved these backup codes"
  - Disable "Continue" button until checked
  - Show warning: "These codes cannot be recovered if lost"
  - Show consequence: "Contact your admin if you lose access"
  - Allow re-display of codes before continuing (one more time)
  - After continuing: NEVER show codes again

#### Scenario: Backup codes storage

- **WHEN** backup codes are stored in database
- **THEN** the system MUST:
  - Bcrypt-hash each code (like passwords)
  - Store only hashes, never plaintext
  - Associate codes with operator account
  - Mark all codes as "unused"
  - Track which code index was used (for audit)
  - Delete code or set `used_at` when consumed

### Requirement: Signup form validation

Signup forms MUST provide real-time validation and error handling.

#### Scenario: Name field validation

- **WHEN** operator enters name
- **THEN** the system MUST:
  - Require minimum 2 characters
  - Allow Unicode characters (international names)
  - Trim leading/trailing whitespace
  - Show inline error if too short
  - Disable submit until valid

#### Scenario: Email field validation (optional for WhatsApp signup)

- **WHEN** operator enters email (if provided)
- **THEN** the system MUST:
  - Validate email format (RFC 5322)
  - Show inline error for invalid format
  - Check for duplicate email in tenant (if provided)
  - Mark email as optional for WhatsApp invitations
  - Required only for email-based invitations

#### Scenario: WhatsApp number validation

- **WHEN** WhatsApp number is entered or pre-filled
- **THEN** the system MUST:
  - Validate international format (E.164)
  - Require country code (e.g., +1, +91, +44)
  - Normalize to standard format before storage
  - Check for duplicate number in tenant
  - Show inline error for invalid format
  - Provide format hint: "+1234567890"

#### Scenario: Invitation token validation

- **WHEN** operator accesses invitation link
- **THEN** the system MUST:
  - Validate token format (UUID)
  - Check token exists in database
  - Verify token has not expired (7 days)
  - Verify token has not been revoked
  - Verify token has not been accepted
  - Extract pre-filled data from token
  - Show error if any validation fails

### Requirement: Signup rate limiting

Signup endpoints MUST be rate-limited to prevent abuse.

#### Scenario: Tenant signup rate limiting

- **WHEN** an IP exceeds 5 tenant signup attempts in 10 minutes
- **THEN** the system MUST:
  - Block further attempts for 30 minutes
  - Return generic error: "Too many attempts. Try again later."
  - Log the event for security monitoring
  - NOT reveal whether email exists

#### Scenario: Invitation acceptance rate limiting

- **WHEN** an invitation token is accessed >20 times in 1 hour
- **THEN** the system MUST:
  - Temporarily block access to that token
  - Notify the inviting admin
  - Log for security review (potential token sharing)
  - Require admin to resend invitation if legitimate

#### Scenario: TOTP setup rate limiting

- **WHEN** TOTP verification fails 5 times in 10 minutes
- **THEN** the system MUST:
  - Invalidate the TOTP secret
  - Block further attempts for 30 minutes
  - Require restart of signup flow
  - Log the security event

### Requirement: Signup UI/UX

Signup flow MUST provide clear, intuitive guidance.

#### Scenario: Multi-step signup flow

- **WHEN** operator accepts invitation
- **THEN** they MUST see:
  - Progress indicator (Step 1 of 3, etc.)
  - Clear step labels:
    1. "Your details"
    2. "Set up authentication"
    3. "Save backup codes"
  - Back/Next navigation
  - Ability to go back (except after completing TOTP verification)
  - Estimated time to complete (< 2 minutes)

#### Scenario: TOTP setup instructions

- **WHEN** operator reaches TOTP setup step
- **THEN** they MUST see:
  - Clear heading: "Set up two-factor authentication"
  - Step-by-step instructions:
    1. "Install an authenticator app"
       - Links: Google Authenticator, Authy, 1Password, Bitwarden
    2. "Scan the QR code with your app"
       - QR code displayed
    3. "Or enter code manually"
       - Manual secret displayed
    4. "Enter the 6-digit code from your app"
       - Input field with countdown timer
  - Visual aids (icons, diagrams)
  - Link to detailed help documentation
  - Video tutorial (optional)

#### Scenario: Loading states

- **WHEN** signup form is submitted
- **THEN** the system MUST:
  - Show loading spinner on submit button
  - Disable submit (prevent double-submission)
  - Display progress message: "Creating your account..."
  - Maintain form values during processing
  - Re-enable on error with clear error message

#### Scenario: Error handling

- **WHEN** signup fails
- **THEN** the system MUST:
  - Display clear error message at top of form
  - Highlight specific fields with errors
  - Preserve all valid form data
  - Provide specific guidance to fix errors
  - Offer help link for common issues
  - Log error for debugging (without PII)

#### Scenario: Success and auto-login

- **WHEN** signup completes successfully
- **THEN** the system MUST:
  - Display success message: "Account created successfully!"
  - Auto-log the operator in
  - Create session cookie
  - Redirect to dashboard or setup wizard
  - Send welcome WhatsApp message (optional)
  - Track signup completion in analytics

### Requirement: Accessibility for signup

Signup flow MUST be accessible to all users.

#### Scenario: Screen reader support

- **WHEN** signup is accessed with screen reader
- **THEN** the system MUST:
  - Provide proper ARIA labels for all fields
  - Announce step changes in multi-step flow
  - Announce form errors clearly
  - Describe QR code purpose and alternatives
  - Provide text description of visual elements

#### Scenario: Keyboard navigation

- **WHEN** operator uses keyboard only
- **THEN** the system MUST:
  - Support tab navigation through all fields
  - Show visible focus indicators
  - Support Enter key to submit form
  - Support keyboard shortcuts for common actions
  - Maintain logical tab order

#### Scenario: Mobile responsiveness

- **WHEN** signup is accessed on mobile device
- **THEN** the system MUST:
  - Display legible text without zooming
  - Support touch targets minimum 44x44 pixels
  - Prevent auto-zoom on input focus (iOS)
  - Optimize QR code size for mobile scanning
  - Support landscape and portrait orientations
  - Load quickly on mobile networks (< 3 seconds)

---

## Database Schema (Reference)

See `design.md` for complete schema. Key tables:

- `tenants` - Tenant records
- `operators` - Operator accounts with `totp_secret_encrypted`
- `totp_secrets` - Encrypted TOTP secrets (audit history)
- `totp_backup_codes` - Hashed backup codes
- `invitations` - Operator invitations (WhatsApp/email/manual)
- `email_verification_tokens` - Email verification (admin only)

---

## Out of Scope (Deferred)

- Password-based signup (deprecated)
- Social login (Google, Microsoft)
- SSO/SAML integration
- SMS-based verification
- Self-service operator signup without invitation (open registration)
- Bulk operator import with pre-generated TOTP secrets
- Custom TOTP parameters per operator (8 digits, non-standard periods)
