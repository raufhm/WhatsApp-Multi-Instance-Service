# Password Reset

## Design Principle: WhatsApp-First, Email-Fallback

Password reset for operators MUST prioritize WhatsApp as the delivery channel:
- **Primary**: WhatsApp message with reset link (for operators with WhatsApp number)
- **Fallback**: Email (only if operator has email on file AND WhatsApp delivery fails)

Tenant admin password reset MAY use email (since admin signup requires email verification).

---

## ADDED Requirements

### Requirement: Self-service password reset via WhatsApp

Operators MUST be able to reset their passwords via WhatsApp message.

#### Scenario: Request password reset via WhatsApp

- **GIVEN** an operator on the login page
- **WHEN** they click "Forgot password?"
- **THEN** they MUST see a form with options:
  - WhatsApp number + tenant ID (primary, recommended)
  - Email + tenant ID (fallback, only if email is on file)
- **AND** be able to submit the request

#### Scenario: Password reset WhatsApp message sent

- **GIVEN** a valid WhatsApp number and tenant ID
- **WHEN** the reset request is submitted
- **THEN** the system MUST:
  - Look up the operator by WhatsApp number + tenant
  - Generate a unique reset token (UUID, 1-hour expiry)
  - Store it hashed in the database
  - Send a password reset WhatsApp message
  - Return generic success message (never reveal if number exists)
  - Example message:
    ```
    Password Reset Request
    
    You requested to reset your password for WhatsApp Operator Dashboard.
    
    Reset here: https://your-domain.com/dashboard/reset-password/:token
    
    This link expires in 1 hour and can only be used once.
    
    If you didn't request this, you can safely ignore this message.
    ```

#### Scenario: Password reset message content

- **WHEN** a password reset WhatsApp message is sent
- **THEN** it MUST include:
  - Clear heading: "Password Reset Request"
  - Reset link with token (valid for 1 hour)
  - Instructions for manual copy-paste
  - Warning: "If you didn't request this, ignore this message"
  - Security notice: "This link can only be used once"
  - Support contact information

#### Scenario: Reset password with valid token

- **GIVEN** a valid, unexpired reset token
- **WHEN** the user clicks the reset link
- **THEN** they MUST see a form to enter new password
- **AND** the form MUST validate password strength
- **AND** on submit, hash and store the new password
- **AND** invalidate the reset token
- **AND** invalidate all existing sessions (force re-login)
- **AND** show success message
- **AND** redirect to login

#### Scenario: Reset link expired

- **GIVEN** a reset token older than 1 hour
- **WHEN** the user attempts to use it
- **THEN** the system MUST reject with "Reset link expired"
- **AND** offer to request a new reset link

#### Scenario: Reset token already used

- **GIVEN** a reset token that has been used
- **WHEN** the user attempts to use it again
- **THEN** the system MUST reject with "This reset link has already been used"
- **AND** suggest requesting a new reset or contacting support

#### Scenario: Invalid reset token

- **WHEN** an invalid or malformed token is provided
- **THEN** the system MUST reject with "Invalid reset link"
- **AND** NOT modify any account data

### Requirement: Password reset rate limiting

Password reset requests MUST be rate-limited to prevent abuse.

#### Scenario: Rate limit by email

- **WHEN** more than 3 reset requests are made for the same email in 1 hour
- **THEN** further requests MUST be blocked for 1 hour
- **AND** return generic success message (to avoid email enumeration)
- **AND** log the event for security monitoring

#### Scenario: Rate limit by IP

- **WHEN** an IP makes more than 10 reset requests in 1 hour
- **THEN** further requests from that IP MUST be blocked
- **AND** return generic success message
- **AND** log for security monitoring

### Requirement: Password reset token security

Reset tokens MUST be cryptographically secure and single-use.

#### Scenario: Token generation

- **WHEN** a reset token is created
- **THEN** it MUST be a cryptographically secure random UUID v4
- **AND** NOT be predictable or sequential

#### Scenario: Token storage

- **WHEN** a reset token is stored
- **THEN** it MUST be hashed (SHA-256) in the database
- **AND** include:
  - Hashed token
  - Operator ID
  - Created timestamp
  - Expiry timestamp (1 hour)
  - Used flag (default false)

#### Scenario: Token invalidation on password change

- **GIVEN** pending reset tokens for an operator
- **WHEN** the operator changes their password (via reset or settings)
- **THEN** ALL reset tokens for that operator MUST be invalidated
- **AND** ALL existing sessions MUST be revoked

#### Scenario: Token invalidation on password reset

- **WHEN** a reset token is successfully used
- **THEN** that token MUST be marked as used
- **AND** all other reset tokens for the operator MUST be invalidated
- **AND** all sessions MUST be revoked

### Requirement: Password reset UI/UX

The password reset flow MUST provide clear, accessible feedback.

#### Scenario: Forgot password page

- **WHEN** a user visits /forgot-password
- **THEN** they MUST see:
  - Clear instructions
  - Email input field
  - Tenant ID input field
  - Submit button
  - "Back to login" link

#### Scenario: Reset request submitted

- **WHEN** the reset form is submitted
- **THEN** show a loading state
- **AND** on completion, show generic success:
  - "If an account exists with this email, you will receive a password reset link"
  - "Check your spam folder if you don't see it"
  - NOT reveal whether the email exists in the system

#### Scenario: Reset password page

- **WHEN** a user visits a valid reset link
- **THEN** they MUST see:
  - "Reset your password" heading
  - New password field with strength indicator
  - Confirm password field
  - Submit button
  - Password requirements list

#### Scenario: Password validation on reset

- **WHEN** the new password doesn't meet requirements
- **THEN** the form MUST show inline errors
- **AND** prevent submission
- **AND** highlight specific requirements not met

#### Scenario: Successful password reset

- **WHEN** password reset succeeds
- **THEN** show:
  - Success icon/animation
  - "Password reset successfully!" message
  - "Your other sessions have been logged out" notice
  - "Continue to login" button
  - Auto-redirect to login after 3 seconds

#### Scenario: Failed password reset

- **WHEN** password reset fails
- **THEN** show:
  - Error icon
  - Clear error message
  - "Request new reset link" option
  - "Contact support" link if problem persists

### Requirement: Email as optional fallback

Email-based password reset MUST remain available as a fallback option for operators without WhatsApp or when WhatsApp delivery fails.

#### Scenario: Request password reset via email (fallback)

- **GIVEN** an operator on the forgot password page
- **WHEN** they choose "Send via email" instead of WhatsApp
- **THEN** the system MUST:
  - Require email + tenant ID
  - Check if the email exists on file for the tenant
  - If email exists: send reset email, return generic success
  - If email doesn't exist: return generic success (no enumeration)
  - If email is required but not on file: show error "No email on file. Use WhatsApp number instead."

#### Scenario: Fallback to email when WhatsApp fails

- **GIVEN** a WhatsApp password reset fails to deliver
- **WHEN** the delivery fails after 3 retry attempts
- **THEN** the system MUST:
  - Check if the operator has an email on file
  - If yes: send password reset email as fallback
  - Notify the operator via WhatsApp: "WhatsApp delivery failed. Check your email for reset link."
  - If no email: show error on next login attempt: "Unable to send reset message. Contact your admin."

#### Scenario: Password reset notification via WhatsApp

- **WHEN** a password is successfully reset
- **THEN** the system MUST send a WhatsApp confirmation:
  - Example:
    ```
    Password Changed Successfully
    
    Your password for WhatsApp Operator Dashboard has been updated.
    
    If this wasn't you, contact your organization's admin immediately.
    ```
- **AND** optionally send email notification if operator has email on file

#### Scenario: Admin initiates operator password reset via WhatsApp

- **GIVEN** a tenant admin
- **WHEN** they view an operator's profile
- **THEN** they MUST see a "Reset password" action
- **AND** clicking it triggers a password reset WhatsApp message to the operator
- **AND** the operator's existing sessions are NOT immediately revoked
- **AND** an audit log entry is created
- **AND** the admin MAY choose to send via email instead (if operator has email on file)

#### Scenario: Admin sets temporary password

- **GIVEN** a tenant admin
- **WHEN** they reset an operator's password
- **THEN** they MAY optionally set a temporary password
- **AND** the operator MUST change it on next login
- **AND** the temporary password MUST meet minimum requirements
- **AND** the operator sees a "Change required" prompt on login

#### Scenario: Admin password reset audit

- **WHEN** an admin resets an operator's password
- **THEN** an audit log entry MUST include:
  - Admin ID
  - Operator ID and email
  - Timestamp
  - Whether temporary password was set
  - IP address of admin

### Requirement: Password reset for unverified accounts

Special handling for password reset requests from unverified accounts.

#### Scenario: Reset request from unverified account

- **GIVEN** an operator who hasn't verified their email
- **WHEN** they request a password reset
- **THEN** the system MUST:
  - Send the reset email (proves email access)
  - OR redirect to complete email verification first
  - Policy choice: allow reset OR require verification first

#### Scenario: Reset completes verification

- **GIVEN** an unverified operator
- **WHEN** they successfully reset password via email
- **THEN** their email MAY be marked as verified
- **AND** they can log in immediately
