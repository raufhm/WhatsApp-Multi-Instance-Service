# Sign In (TOTP Authentication)

## Design Principle: Passwordless with TOTP

Authentication MUST use **TOTP (Time-based One-Time Password)** codes from authenticator apps:
- **No passwords**: Operators authenticate with tenant ID + email/WhatsApp number + TOTP code
- **Authenticator app agnostic**: Works with Google Authenticator, Authy, 1Password, Bitwarden, etc.
- **Backup codes**: Recovery when operators lose authenticator access
- **Session-based**: Successful TOTP login creates a session cookie

---

## ADDED Requirements

### Requirement: TOTP-based authentication

Operators MUST authenticate using a 6-digit TOTP code from their authenticator app.

#### Scenario: Login succeeds with valid TOTP code

- **GIVEN** a verified operator account with TOTP setup
- **WHEN** they enter correct tenant ID, email/WhatsApp number, and current TOTP code
- **THEN** the system MUST:
  - Validate tenant ID format (UUID)
  - Look up operator by email/WhatsApp number + tenant ID
  - Decrypt the operator's TOTP secret
  - Verify the TOTP code (current period ±1 for clock drift)
  - Use constant-time comparison (prevent timing attacks)
  - Create a session record
  - Set HttpOnly, Secure, SameSite=Lax session cookie
  - Log successful login (timestamp, IP, device info)
  - Redirect to dashboard inbox

#### Scenario: Login fails with invalid TOTP code

- **WHEN** incorrect TOTP code is provided
- **THEN** the system MUST:
  - Return 401 Unauthorized
  - Display generic error: "Invalid credentials"
  - NOT reveal whether email/number exists
  - NOT reveal whether the code was close to valid
  - Log failed attempt for security monitoring
  - Rate limit further attempts (5 per 15 minutes)

#### Scenario: Login fails for TOTP not setup

- **GIVEN** an operator whose TOTP was reset by admin
- **WHEN** they attempt to log in
- **THEN** the system MUST:
  - Reject with "Two-factor authentication setup required"
  - Display: "Contact your admin to complete TOTP setup"
  - NOT allow login without completing TOTP setup

#### Scenario: Login fails for inactive account

- **GIVEN** a deactivated operator account
- **WHEN** they attempt to log in
- **THEN** the system MUST:
  - Reject with "Account deactivated"
  - Instruct them to contact their admin
  - NOT attempt TOTP verification

### Requirement: Tenant ID handling

The tenant ID MUST be provided during login for multi-tenant isolation.

#### Scenario: Tenant ID is required

- **WHEN** no tenant ID is provided
- **THEN** the system MUST:
  - Reject with "Tenant ID is required"
  - Highlight the missing field in the UI
  - NOT attempt authentication

#### Scenario: Invalid tenant ID format

- **WHEN** a non-UUID tenant ID is provided
- **THEN** the system MUST:
  - Reject with "Invalid tenant ID format"
  - Show format hint (e.g., "e.g., 550e8400-e29b-41d4-a716-446655440000")
  - NOT attempt authentication

#### Scenario: Remember tenant ID

- **WHEN** an operator logs in successfully
- **THEN** the system MAY:
  - Store tenant ID in localStorage
  - Pre-fill tenant ID on subsequent login attempts
  - Provide "Forget tenant" action to clear stored ID
  - Only store tenant ID, NEVER store credentials

### Requirement: Backup code login (Recovery)

Operators MUST be able to log in using backup codes when they lose authenticator access.

#### Scenario: Login with backup code succeeds

- **GIVEN** an operator has lost authenticator access
- **WHEN** they click "Lost access to authenticator?" link
- **AND** enter valid tenant ID, email/WhatsApp number, and unused backup code
- **THEN** the system MUST:
  - Validate the backup code (bcrypt comparison)
  - Mark that specific backup code as used (delete or set used_at)
  - Create a session record
  - Set session cookie
  - Show warning: "You have X backup codes remaining"
  - Strongly suggest: "Regenerate backup codes immediately"
  - Redirect to account settings page
  - Log backup code usage for security audit

#### Scenario: Login with backup code fails

- **WHEN** invalid or used backup code is provided
- **THEN** the system MUST:
  - Return 401 Unauthorized
  - Display generic error: "Invalid credentials"
  - NOT reveal whether code was already used
  - Rate limit further attempts (3 per 15 minutes)
  - Log failed attempt for security monitoring

#### Scenario: Out of backup codes warning

- **GIVEN** an operator with 0-2 backup codes remaining
- **WHEN** they log in successfully (via TOTP or backup code)
- **THEN** the system MUST:
  - Show prominent warning banner: "You have X backup codes remaining"
  - Display "Regenerate backup codes now" button (prominent)
  - Provide "Remind me later" option (but show on every login until resolved)
  - Optionally: require regeneration if 0 codes remain (configurable per tenant policy)

### Requirement: TOTP code input UX

The TOTP code input MUST provide an intuitive, accessible user experience.

#### Scenario: TOTP code input field

- **WHEN** operator visits login page
- **THEN** they MUST see:
  - Clear label: "Authentication code" or "TOTP code"
  - 6-digit input field (numeric only)
  - Numeric keyboard on mobile devices
  - Auto-format as 6 digits
  - Auto-focus on the field (after tenant/email are filled)
  - Countdown timer showing seconds until code expires (30s cycle)
  - Visual indicator of code expiry (color change, progress bar)

#### Scenario: TOTP code auto-submit

- **WHEN** operator enters 6 digits
- **THEN** the form MAY:
  - Auto-submit immediately (optional, configurable)
  - OR require manual submit button click
  - Show loading state during verification
  - Disable submit button during verification (prevent double-submit)

#### Scenario: Invalid TOTP code handling

- **WHEN** invalid TOTP code is submitted
- **THEN** the system MUST:
  - Clear the input field
  - Re-focus on the input field
  - Show error: "Invalid code. Please try again."
  - Display countdown timer for next code
  - Allow immediate retry (within rate limits)
  - NOT reveal whether code was close to valid

#### Scenario: "Lost authenticator?" link

- **WHEN** operator clicks "Lost access to authenticator?"
- **THEN** the system MUST:
  - Hide TOTP code input field
  - Show backup code input field
  - Display instructions: "Enter one of your 10 backup codes"
  - Show format hint: "e.g., A7B9-C2D4 (8 characters, with or without hyphen)"
  - Provide "Back to TOTP code" link
  - Log the event for security monitoring (optional)

### Requirement: Login form UX

The login form MUST provide a smooth, accessible user experience.

#### Scenario: Form validation

- **WHEN** required fields are empty on submit
- **THEN** the form MUST:
  - Show inline errors for each missing field
  - Prevent submission until all required fields are filled
  - Highlight missing fields visually
  - Announce errors to screen readers

#### Scenario: Loading state during login

- **WHEN** the login form is submitted
- **THEN** the system MUST:
  - Show loading spinner on submit button
  - Disable submit button (prevent double-submission)
  - Display "Signing in..." or similar message
  - Re-enable on success or error
  - Maintain tenant ID and email/number values on error (don't clear)
  - Clear TOTP code field on error (security)

#### Scenario: Error display

- **WHEN** login fails
- **THEN** the system MUST:
  - Display error message clearly above the form
  - Use generic error: "Invalid credentials" (no enumeration)
  - Provide helpful recovery options:
    - "Lost access to authenticator?" link
    - "Contact your admin" link
  - NOT expose sensitive information (e.g., whether email exists)

#### Scenario: Navigate to login when unauthenticated

- **WHEN** an unauthenticated user visits a protected route
- **THEN** the system MUST:
  - Redirect to /login
  - Optionally preserve intended destination for post-login redirect
  - Display message: "Please log in to continue"

#### Scenario: Redirect authenticated users away from login

- **GIVEN** an already authenticated operator
- **WHEN** they visit /login
- **THEN** the system MUST:
  - Redirect to dashboard inbox immediately
  - NOT show the login form
  - Preserve any post-login redirect destination

### Requirement: Session management

Sessions MUST be securely managed with proper expiry and revocation.

#### Scenario: Session creation on login

- **WHEN** login succeeds (TOTP or backup code)
- **THEN** the system MUST create a session record with:
  - Unique session ID (UUID v4)
  - Operator ID reference
  - Tenant ID reference
  - Expiry timestamp (8 hours from creation by default)
  - Created timestamp
  - IP address (for audit)
  - User agent / device info (for audit)
  - Login method: "totp" or "backup_code"

#### Scenario: Session validation on each request

- **WHEN** a request includes a session cookie
- **THEN** the system MUST validate:
  - Session exists in database
  - Session has not expired
  - Session belongs to active operator
  - Session tenant matches request tenant context
  - Session has not been revoked
- **AND** reject invalid sessions with 401 and redirect to /login

#### Scenario: Session extension on activity

- **WHEN** a valid session is used
- **THEN** the system MAY:
  - Update session's last_used_at timestamp
  - Extend expiry using sliding window (max 24 hours)
  - NOT extend "remember me" sessions beyond 30 days

#### Scenario: Logout invalidates session

- **WHEN** an operator logs out
- **THEN** the system MUST:
  - Delete session from database
  - Clear session cookie client-side
  - Redirect to /login
  - Display: "You have been logged out successfully"
  - NOT allow session reuse

#### Scenario: Session expiry redirect

- **GIVEN** an expired session
- **WHEN** operator attempts to access protected route
- **THEN** the system MUST:
  - Redirect to /login
  - Display: "Your session has expired. Please log in again."
  - Optionally preserve intended destination for post-login redirect

### Requirement: Remember me functionality

Operators MUST be able to opt for extended sessions.

#### Scenario: Remember me extends session

- **GIVEN** the "Remember me" checkbox is selected during login
- **WHEN** login succeeds
- **THEN** the system MUST:
  - Extend session expiry to 30 days
  - Store a flag indicating "remember me" session
  - Use persistent cookie (not session cookie)
  - Still require TOTP on next login (no bypass)

#### Scenario: Remember me is opt-in

- **WHEN** "Remember me" is NOT selected
- **THEN** the system MUST:
  - Use default session expiry (8 hours)
  - Use session cookie (expires on browser close)
  - Require login again after browser restart

#### Scenario: Remember me security notice

- **WHEN** operator sees "Remember me" option
- **THEN** the UI MUST display:
  - Clear explanation: "Keep me logged in for 30 days"
  - Warning: "Only use on personal devices"
  - Recommendation: "Do not use on shared or public computers"

### Requirement: Login rate limiting

Login attempts MUST be rate-limited to prevent brute force attacks.

#### Scenario: TOTP login rate limiting

- **WHEN** an email/WhatsApp number exceeds 5 failed TOTP attempts in 15 minutes
- **THEN** the system MUST:
  - Block further TOTP attempts for 30 minutes
  - Return generic error: "Too many failed attempts. Try again later."
  - Log the event for security monitoring
  - Optionally notify operator via WhatsApp (if configured)
  - NOT reveal whether email/number exists

#### Scenario: Backup code rate limiting

- **WHEN** backup code attempts exceed 3 failed attempts in 15 minutes
- **THEN** the system MUST:
  - Block further backup code attempts for 60 minutes
  - Return generic error: "Too many failed attempts. Try again later."
  - Log the event for security review
  - Notify operator via WhatsApp: "Suspicious backup code attempts detected"
  - Alert tenant admins if pattern suggests attack

#### Scenario: Rate limit by IP (optional)

- **WHEN** an IP address exceeds 20 failed login attempts in 1 hour
- **THEN** the system MAY:
  - Block all login attempts from that IP for 24 hours
  - Log for security review (potential brute force attack)
  - Alert security team
  - Add IP to watchlist

### Requirement: Security and privacy

Login MUST implement security best practices to protect operators.

#### Scenario: Constant-time comparison

- **WHEN** verifying TOTP codes or backup codes
- **THEN** the system MUST:
  - Use constant-time comparison algorithms
  - NOT short-circuit on first mismatch
  - Maintain consistent response time regardless of failure point
  - Prevent timing attacks

#### Scenario: Credential enumeration prevention

- **WHEN** a login attempt fails
- **THEN** the system MUST:
  - Return the same generic error regardless of failure reason:
    - Invalid tenant ID
    - Invalid email/WhatsApp number
    - Invalid TOTP code
    - Invalid backup code
    - Account not found
    - Account deactivated
  - Use consistent response time for all failure scenarios
  - NOT reveal whether an account exists

#### Scenario: Secure cookie flags

- **WHEN** setting session cookie
- **THEN** the system MUST:
  - Set `HttpOnly` flag (prevent JavaScript access)
  - Set `Secure` flag in production (HTTPS only)
  - Set `SameSite=Lax` (prevent CSRF)
  - Set appropriate `Path` (/dashboard)
  - Set appropriate `Expires` or `Max-Age`
  - NOT include sensitive data in cookie value

#### Scenario: Login audit logging

- **WHEN** a login attempt occurs (success or failure)
- **THEN** the system MUST log:
  - Timestamp (UTC)
  - Tenant ID (masked in logs)
  - Email/WhatsApp number (masked: `ope***@***.com` or `+1***555****`)
  - Login method: "totp", "backup_code"
  - Success/failure status
  - IP address (for security analysis)
  - User agent / device type
  - Failure reason (for monitoring, not exposed to user)
- **AND** NOT log:
  - Full email addresses
  - Full WhatsApp numbers
  - TOTP codes
  - Backup codes
  - TOTP secrets

### Requirement: Multi-device and multi-session support

Operators MAY log in from multiple devices simultaneously.

#### Scenario: Concurrent sessions allowed

- **GIVEN** an operator logged in on one device
- **WHEN** they log in from another device
- **THEN** the system MUST:
  - Allow both sessions to coexist
  - Create a new session record for the new device
  - NOT invalidate the existing session
  - Track both sessions independently
  - Allow operator to view active sessions (optional, future feature)

#### Scenario: Session revocation (future)

- **GIVEN** multiple active sessions
- **WHEN** operator revokes a session (manual action)
- **THEN** the system MUST:
  - Delete the specific session from database
  - Invalidate the session cookie
  - Log the revocation event
  - Notify operator of successful revocation
- **NOTE**: This is deferred to a future "session management" feature

### Requirement: Accessibility requirements

The login form MUST be accessible to all users.

#### Scenario: Screen reader support

- **WHEN** login page is accessed with screen reader
- **THEN** the system MUST:
  - Provide proper ARIA labels for all fields
  - Announce form errors clearly
  - Announce success/error states
  - Provide descriptive text for TOTP code field
  - Include instructions for backup code login

#### Scenario: Keyboard navigation

- **WHEN** operator uses keyboard only
- **THEN** the system MUST:
  - Support tab navigation through all fields
  - Show visible focus indicators
  - Support Enter key to submit form
  - Support Escape key to clear fields (optional)
  - Maintain logical tab order

#### Scenario: Visual accessibility

- **WHEN** login page is rendered
- **THEN** it MUST:
  - Meet WCAG AA contrast requirements (4.5:1 for text)
  - Not rely solely on color to convey information
  - Provide text alternatives for icons
  - Support browser zoom up to 200% without breaking layout
  - Use clear, readable fonts (minimum 16px for input fields)

#### Scenario: Mobile accessibility

- **WHEN** login page is accessed on mobile device
- **THEN** it MUST:
  - Display numeric keypad for TOTP code input
  - Support touch targets minimum 44x44 pixels
  - Prevent auto-zoom on input focus (iOS)
  - Support landscape and portrait orientations
  - Load quickly on mobile networks (< 3 seconds)

---

## Out of Scope (Deferred)

- Social login (Google, Microsoft, Apple)
- SSO/SAML integration (enterprise feature)
- Biometric authentication (Touch ID, Face ID)
- Hardware token support (YubiKey)
- Session management UI (view/revoke active sessions)
- Geographic login restrictions
- Device trust / "remember this device" beyond session cookies
