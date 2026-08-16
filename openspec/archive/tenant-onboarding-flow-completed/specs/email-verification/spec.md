# Email Verification

## ADDED Requirements

### Requirement: Email verification is required for new accounts

All new tenant admins and operators MUST verify their email before accessing the dashboard.

#### Scenario: Verification email sent on signup

- **GIVEN** a user completes signup
- **WHEN** the signup succeeds
- **THEN** the system MUST generate a unique verification token (UUID)
- **AND** store it with expiry (24 hours)
- **AND** send a verification email with a unique link

#### Scenario: Verification email content

- **WHEN** a verification email is sent
- **THEN** it MUST include:
  - Recipient name
  - Organization name (for operators)
  - Clear call-to-action button/link
  - Verification link with token
  - Expiry information (24 hours)
  - Instructions for manual copy-paste if link doesn't work

#### Scenario: Email verification succeeds

- **GIVEN** a valid, unexpired verification token
- **WHEN** the user clicks the verification link
- **THEN** the system MUST mark the email as verified
- **AND** set verified_at timestamp
- **AND** delete the verification token
- **AND** redirect to login with success message

#### Scenario: Expired verification token

- **GIVEN** a verification token older than 24 hours
- **WHEN** the user attempts to verify
- **THEN** the system MUST reject with "Verification link expired"
- **AND** offer to resend verification email

#### Scenario: Already verified account

- **GIVEN** an already verified account
- **WHEN** the user clicks a verification link (old or new)
- **THEN** the system MUST show "Email already verified"
- **AND** redirect to login

#### Scenario: Invalid verification token

- **WHEN** an invalid or malformed token is provided
- **THEN** the system MUST reject with "Invalid verification link"
- **AND** NOT modify any account data

### Requirement: Resend verification email

Users MUST be able to request a new verification email.

#### Scenario: Resend from verification page

- **GIVEN** a user on the "verify your email" page
- **WHEN** they click "Resend verification email"
- **THEN** the system MUST:
  - Generate a new verification token
  - Invalidate previous tokens
  - Send a new verification email
  - Show success confirmation
  - Rate limit to 3 requests per hour

#### Scenario: Resend from login page

- **GIVEN** an unverified user attempting to log in
- **WHEN** they see the "verify email" error
- **THEN** they MUST see a "Resend verification email" link
- **AND** clicking it sends a new verification email

### Requirement: Verification token security

Verification tokens MUST be secure and single-use.

#### Scenario: Token is single-use

- **GIVEN** a verification token has been used
- **WHEN** the same token is presented again
- **THEN** the system MUST reject with "Invalid or expired token"
- **AND** NOT allow re-verification

#### Scenario: Token format

- **WHEN** a verification token is generated
- **THEN** it MUST be a cryptographically secure random UUID v4
- **AND** NOT be sequential or predictable

#### Scenario: Token storage

- **WHEN** a verification token is created
- **THEN** it MUST be stored hashed (SHA-256) in the database
- **AND** the plaintext token is ONLY included in the email link
- **AND** the token MUST have an expiry timestamp

### Requirement: Admin verification bypass

Tenant admins MUST be able to manually verify operators in special cases.

#### Scenario: Admin verifies operator manually

- **GIVEN** a tenant admin
- **WHEN** they view pending operators
- **THEN** they MUST see unverified operators
- **AND** have a "Verify manually" action
- **AND** the operator's email is marked verified immediately

#### Scenario: Manual verification audit log

- **WHEN** an admin manually verifies an operator
- **THEN** an audit log entry MUST be created
- **AND** include admin ID, operator email, timestamp, reason

### Requirement: Verification UI states

The verification flow MUST provide clear feedback at each stage.

#### Scenario: Verification page loading

- **WHEN** a user clicks a verification link
- **THEN** the page MUST show a loading spinner
- **AND** display "Verifying your email..." message
- **AND** automatically process the token

#### Scenario: Verification success

- **WHEN** verification succeeds
- **THEN** the page MUST show:
  - Success icon/animation
  - "Email verified successfully!" message
  - "Continue to login" button
  - Auto-redirect to login after 3 seconds

#### Scenario: Verification failure

- **WHEN** verification fails (expired, invalid)
- **THEN** the page MUST show:
  - Error icon
  - Clear error message with reason
  - "Resend verification email" button
  - "Back to login" link

#### Scenario: Email sent confirmation

- **WHEN** a verification email is sent or resent
- **THEN** the UI MUST show:
  - Confirmation message
  - "Check your inbox" instruction
  - "Didn't receive email? Check spam folder" hint
  - Countdown timer until resend is available

### Requirement: Verification email template

Verification emails MUST follow branding and accessibility standards.

#### Scenario: Email branding

- **WHEN** a verification email is sent
- **THEN** it MUST include:
  - Organization logo/header
  - Consistent color scheme
  - Professional, clear language
  - Contact information for support

#### Scenario: Email accessibility

- **WHEN** a verification email is sent
- **THEN** it MUST:
  - Have proper alt text for images
  - Use semantic HTML
  - Be readable in plain text fallback
  - Work in major email clients (Gmail, Outlook, Apple Mail)

#### Scenario: Email localization

- **GIVEN** a tenant with a preferred language
- **WHEN** a verification email is sent
- **THEN** it SHOULD be in the tenant's language (if i18n is implemented)
- **OR** default to English

### Requirement: Bulk verification for imported operators

When operators are imported via CSV/bulk upload, verification MUST be handled appropriately.

#### Scenario: Imported operators require verification

- **WHEN** operators are imported by admin
- **THEN** they MUST receive verification emails
- **AND** follow the standard verification flow
- **EXCEPT** if admin explicitly skips verification (trusted import)

#### Scenario: Imported operators with verification skip

- **GIVEN** admin marks import as "trusted"
- **WHEN** operators are imported
- **THEN** their emails MAY be marked verified immediately
- **AND** a welcome email is sent with login instructions
- **AND** audit log records the trusted import
