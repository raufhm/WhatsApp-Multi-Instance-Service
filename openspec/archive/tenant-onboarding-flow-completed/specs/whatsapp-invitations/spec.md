# WhatsApp-First Invitations and TOTP Setup

## Design Principle: WhatsApp-First, TOTP-Based Authentication

The system MUST prioritize WhatsApp as the primary communication channel for:
- Operator invitations with TOTP setup
- TOTP reset notifications
- Account recovery notifications

Authentication is **TOTP-based** (no passwords):
- Operators set up TOTP when accepting invitation
- Login with tenant ID + email/WhatsApp number + TOTP code
- Backup codes for recovery
- Admin can reset TOTP if operator loses access

Email remains available ONLY as:
- Tenant admin verification (one-time during tenant signup)
- Optional fallback for critical notifications

This eliminates vendor lock-in with email providers while leveraging the existing WhatsApp infrastructure and providing superior security with TOTP.

---

## ADDED Requirements

### Requirement: Operators are invited via WhatsApp message

Tenant admins MUST be able to invite operators by sending a WhatsApp message with a signup link.

#### Scenario: Admin creates WhatsApp invitation

- **GIVEN** a tenant admin is on the invitations page
- **WHEN** they enter an operator's WhatsApp number and select a role
- **THEN** the system MUST:
  - Validate the WhatsApp number format
  - Check if the number is already an operator in the tenant
  - Generate a unique invitation token (UUID)
  - Store the invitation with expiry (7 days)
  - Send a WhatsApp message with the invitation link
  - Display success confirmation to the admin

#### Scenario: WhatsApp invitation message content

- **WHEN** an invitation WhatsApp message is sent
- **THEN** it MUST include:
  - Greeting with invitee name (if known) or generic greeting
  - Inviter name (the admin who sent it)
  - Organization name
  - Assigned role
  - Clear call-to-action: signup link with token
  - Mention: "You'll set up two-factor authentication during signup"
  - Expiry information (7 days)
  - Brief instructions
  - Example:
    ```
    Hi! [Admin Name] from [Org Name] has invited you to join their WhatsApp Operator Dashboard as an [Role].
    
    Complete your signup here: https://your-domain.com/dashboard/invitation/abc123...
    
    During signup, you'll set up two-factor authentication (TOTP) using an authenticator app like Google Authenticator or Authy.
    
    This link expires in 7 days.
    
    — WhatsApp Operator Dashboard
    ```

#### Scenario: Operator accepts WhatsApp invitation

- **GIVEN** a valid, unexpired WhatsApp invitation
- **WHEN** the operator clicks the WhatsApp link
- **THEN** the system MUST:
  - Validate the invitation token
  - Pre-fill the signup form with:
    - WhatsApp number (from invitation)
    - Role (from invitation)
    - Tenant context (from invitation)
  - Require only: name, email (optional)
  - Skip email verification (WhatsApp proves identity)
  - **Guide operator through TOTP setup**:
    - Display QR code for authenticator app
    - Provide manual entry key option
    - Require TOTP code verification to complete setup
  - **Generate backup codes** (10 single-use codes)
  - Require operator to acknowledge saving backup codes
  - Create the operator account with verified TOTP
  - Mark the invitation as accepted
  - Log the operator in automatically

#### Scenario: WhatsApp invitation to existing number is rejected

- **GIVEN** a WhatsApp number already exists as an operator in the tenant
- **WHEN** an admin tries to invite that number
- **THEN** the system MUST:
  - Reject with "This WhatsApp number is already registered"
  - Show the existing operator's name and role
  - Offer to resend invitation or modify existing operator

#### Scenario: WhatsApp invitation delivery tracking

- **WHEN** a WhatsApp invitation is sent
- **THEN** the system MUST track:
  - Sent timestamp
  - Delivery status (sent, delivered, read - if available via WhatsApp API)
  - Click timestamp (when link is clicked)
  - Acceptance timestamp (when signup completes)
- **AND** expose this data to admins in the invitations UI

### Requirement: WhatsApp number as primary identifier

Operators MAY use their WhatsApp number as their primary identifier instead of email.

#### Scenario: Signup with WhatsApp number

- **GIVEN** an operator accessing via WhatsApp invitation
- **WHEN** they sign up
- **THEN** the system MUST:
  - Use WhatsApp number as the primary unique identifier
  - Store email as optional secondary contact
  - Allow login with either WhatsApp number OR email (if provided)
  - Display WhatsApp number in the operator profile prominently

#### Scenario: Login with WhatsApp number

- **GIVEN** an operator with a WhatsApp-linked account
- **WHEN** they log in
- **THEN** they MAY use their WhatsApp number instead of email
- **AND** the system MUST accept:
  - WhatsApp number + tenant ID + password
  - OR email + tenant ID + password (if email was provided)

#### Scenario: WhatsApp number format validation

- **WHEN** a WhatsApp number is entered
- **THEN** the system MUST validate:
  - International format (e.g., +1234567890)
  - Country code is present
  - Number length is appropriate for the country
  - Number is not on a blocklist (if applicable)
- **AND** normalize the number to E.164 format before storage

### Requirement: WhatsApp-based password reset

Operators MUST be able to reset their passwords via WhatsApp message.

#### Scenario: Request password reset via WhatsApp

- **GIVEN** an operator on the login page
- **WHEN** they click "Forgot password?"
- **THEN** they MUST see options:
  - "Send reset link via WhatsApp" (primary, recommended)
  - "Send reset link via email" (fallback, if email is on file)
- **AND** be able to choose their preferred method

#### Scenario: WhatsApp password reset message

- **WHEN** a password reset is requested via WhatsApp
- **THEN** the system MUST:
  - Generate a unique reset token (UUID, 1-hour expiry)
  - Store the token hashed in the database
  - Send a WhatsApp message with the reset link
  - Example message:
    ```
    Password Reset Request
    
    You requested to reset your password for WhatsApp Operator Dashboard.
    
    Reset here: https://your-domain.com/dashboard/reset-password/xyz789...
    
    This link expires in 1 hour and can only be used once.
    
    If you didn't request this, you can safely ignore this message.
    ```

#### Scenario: WhatsApp reset link is clicked

- **GIVEN** a valid WhatsApp reset token
- **WHEN** the operator clicks the reset link
- **THEN** the system MUST:
  - Validate the token
  - Show the password reset form
  - On successful password change:
    - Hash and store the new password
    - Invalidate all existing sessions
    - Invalidate all reset tokens
    - Send a WhatsApp confirmation message
    - Redirect to login

#### Scenario: WhatsApp password reset confirmation

- **WHEN** a password is successfully reset
- **THEN** the system MUST send a WhatsApp confirmation:
  - Example:
    ```
    Password Changed Successfully
    
    Your password for WhatsApp Operator Dashboard has been updated.
    
    If this wasn't you, contact your organization's admin immediately.
    ```

### Requirement: Email as optional fallback

Email-based invitations and password reset MUST remain available as a fallback option.

#### Scenario: Admin chooses email invitation

- **GIVEN** an admin creating an invitation
- **WHEN** the operator's WhatsApp number is unknown or unavailable
- **THEN** the admin MAY choose to send invitation via email instead
- **AND** the email invitation flow MUST work as specified in the original email spec
- **AND** email verification is required for email-based invitations

#### Scenario: Fallback to email when WhatsApp fails

- **GIVEN** a WhatsApp invitation fails to deliver (e.g., invalid number, WhatsApp API error)
- **WHEN** the admin is notified of the failure
- **THEN** the system MUST offer to:
  - Retry WhatsApp delivery
  - Send via email instead (if email is available)
  - Generate a manual signup link that the admin can share directly

#### Scenario: Email is optional for operator signup

- **GIVEN** an operator signing up via WhatsApp invitation
- **WHEN** they fill out the signup form
- **THEN** the email field MUST be marked as optional
- **AND** the operator MUST be able to complete signup without providing email
- **AND** a warning MUST explain that password reset will only be possible via WhatsApp if no email is provided

### Requirement: WhatsApp invitation rate limiting

WhatsApp invitations MUST be rate-limited to prevent spam and abuse.

#### Scenario: Rate limit by tenant

- **WHEN** a tenant sends more than 20 WhatsApp invitations in 24 hours
- **THEN** further WhatsApp invitations MUST require admin approval (for that tenant)
- **AND** the system MUST notify the tenant's primary admin of the limit

#### Scenario: Rate limit by WhatsApp number

- **WHEN** the same WhatsApp number receives more than 3 invitations in 7 days
- **THEN** further invitations to that number MUST be blocked for 7 days
- **AND** the inviting admin MUST be notified: "This number has received too many invitations recently"

#### Scenario: Rate limit by sending admin

- **WHEN** an admin sends more than 50 invitations in 24 hours
- **THEN** further invitations from that admin MUST be blocked for 24 hours
- **AND** the action MUST be logged for review

### Requirement: WhatsApp message templates

WhatsApp invitation and reset messages MUST follow consistent, branded templates.

#### Scenario: Template localization

- **GIVEN** a tenant with a preferred language setting
- **WHEN** a WhatsApp message is sent
- **THEN** the message SHOULD be in the tenant's language (if i18n is implemented)
- **OR** default to English
- **AND** support common languages: English, Spanish, Portuguese, Hindi, Arabic (based on tenant needs)

#### Scenario: Template customization (enterprise)

- **GIVEN** an enterprise tenant with custom branding
- **WHEN** WhatsApp messages are sent to their operators
- **THEN** the messages MAY include:
  - Custom greeting format
  - Organization-specific footer
  - Custom support contact
- **AND** the core security elements (link, expiry) MUST remain unchanged

#### Scenario: Message length optimization

- **WHEN** WhatsApp messages are composed
- **THEN** they MUST be concise and scannable
- **AND** stay under WhatsApp's message length limits (1600 characters for most messages)
- **AND** put the most important information (link, action) early in the message

### Requirement: WhatsApp delivery infrastructure

The system MUST leverage existing WhatsApp message sending infrastructure.

#### Scenario: Integration with existing WhatsApp sender

- **GIVEN** the existing WhatsApp message sending system (whatsapp/ package)
- **WHEN** an invitation or password reset is sent
- **THEN** the system MUST:
  - Use the existing message queue and delivery mechanism
  - Apply the same retry logic and backoff strategies
  - Track delivery status in the same way as other WhatsApp messages
  - Log delivery outcomes for monitoring

#### Scenario: Invitation message prioritization

- **GIVEN** the WhatsApp message queue
- **WHEN** invitation or password reset messages are queued
- **THEN** they SHOULD receive higher priority than regular bot messages
- **AND** be delivered with minimal delay (< 30 seconds P95)

#### Scenario: Fallback when WhatsApp is unavailable

- **GIVEN** the WhatsApp service is temporarily unavailable
- **WHEN** an invitation or password reset is requested
- **THEN** the system MUST:
  - Queue the message for retry (with exponential backoff)
  - Notify the admin: "WhatsApp delivery delayed, will retry shortly"
  - Offer email as immediate fallback (if available)
  - Allow manual link generation for direct sharing

### Requirement: Invitation analytics and monitoring

The system MUST track and expose metrics for WhatsApp invitations.

#### Scenario: Invitation status tracking

- **GIVEN** an invitation is created
- **WHEN** admins view the invitations list
- **THEN** they MUST see for each invitation:
  - WhatsApp number (masked for privacy: +1***555****)
  - Role
  - Status: pending, sent, delivered, clicked, accepted, expired, revoked
  - Sent timestamp
  - Accepted timestamp (if applicable)
  - Actions available (resend, revoke)

#### Scenario: Invitation funnel metrics

- **WHEN** the system tracks invitation analytics
- **THEN** it MUST calculate and expose:
  - Total invitations sent (by time period)
  - Delivery rate: delivered / sent
  - Click-through rate: clicked / delivered
  - Acceptance rate: accepted / clicked
  - Time-to-accept: median time from sent to accepted
  - Expiration rate: expired / not accepted
- **AND** make these metrics available in the admin dashboard

#### Scenario: Failed delivery alerts

- **WHEN** WhatsApp invitation delivery fails
- **THEN** the system MUST:
  - Log the failure with reason (invalid number, WhatsApp API error, etc.)
  - Notify the inviting admin
  - Retry up to 3 times with exponential backoff
  - After final failure, mark as "undeliverable" and suggest email fallback

### Requirement: Privacy and compliance for WhatsApp invitations

WhatsApp-based invitations MUST comply with privacy regulations and WhatsApp policies.

#### Scenario: Phone number masking in UI

- **GIVEN** an invitation in the admin UI
- **WHEN** displaying the WhatsApp number
- **THEN** the system MUST mask the number for privacy:
  - Show only country code and last 4 digits
  - Example: +1***555****
  - Full number only visible to users with explicit permission

#### Scenario: Opt-out mechanism

- **GIVEN** an operator receives an unwanted WhatsApp invitation
- **WHEN** they click the link
- **THEN** they MUST have the option to:
  - Decline the invitation explicitly
  - Request no further invitations from this tenant
  - Report spam/abuse (triggers admin review)

#### Scenario: WhatsApp Business Policy compliance

- **WHEN** sending invitation and reset messages
- **THEN** the system MUST comply with WhatsApp Business Policy:
  - Only send to users who have opted in (admin vouches for invitation recipients)
  - Include clear sender identification
  - Provide value to the recipient (legitimate business purpose)
  - Honor WhatsApp's messaging limits and quality standards

### Requirement: Manual invitation link generation

Admins MUST be able to generate invitation links manually for sharing through any channel.

#### Scenario: Generate manual invitation link

- **GIVEN** an admin creating an invitation
- **WHEN** they choose "Generate link manually"
- **THEN** the system MUST:
  - Generate a unique invitation token
  - Display a copyable link: `https://domain.com/dashboard/invitation/:token`
  - Display a QR code that encodes the link (optional)
  - Allow the admin to share the link via any channel (WhatsApp, email, SMS, in-person)
  - Track the invitation the same way as WhatsApp-sent invitations

#### Scenario: Manual link security

- **GIVEN** a manually generated invitation link
- **WHEN** the link is shared publicly
- **THEN** the system MUST:
  - Still enforce the 7-day expiry
  - Allow the admin to revoke the link at any time
  - Limit each link to single use
  - Warn admins: "Share this link only with intended recipients"

---

## Database Schema Additions

### Modified Table: `invitations`

```sql
ALTER TABLE invitations ADD COLUMN channel TEXT NOT NULL DEFAULT 'whatsapp' 
  CHECK (channel IN ('whatsapp', 'email', 'manual'));

ALTER TABLE invitations ADD COLUMN whatsapp_number TEXT;
ALTER TABLE invitations ADD COLUMN email_address TEXT;

-- Make email optional, add whatsapp_number as alternative
ALTER TABLE invitations ALTER COLUMN email DROP NOT NULL;

CREATE INDEX idx_invitations_whatsapp ON invitations(whatsapp_number) WHERE channel = 'whatsapp';
```

### New Table: `whatsapp_invitation_delivery`

```sql
CREATE TABLE IF NOT EXISTS whatsapp_invitation_delivery (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invitation_id   UUID NOT NULL REFERENCES invitations(id) ON DELETE CASCADE,
    status          TEXT NOT NULL DEFAULT 'pending' 
      CHECK (status IN ('pending', 'sent', 'delivered', 'read', 'failed', 'expired')),
    sent_at         TIMESTAMPTZ,
    delivered_at    TIMESTAMPTZ,
    read_at         TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    failure_reason  TEXT,
    retry_count     INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_whatsapp_delivery_status ON whatsapp_invitation_delivery(status);
CREATE INDEX idx_whatsapp_delivery_invitation ON whatsapp_invitation_delivery(invitation_id);
```

### Modified Table: `operators`

```sql
ALTER TABLE operators ADD COLUMN whatsapp_number TEXT UNIQUE;
ALTER TABLE operators ADD COLUMN email TEXT;  -- Already exists, just making it optional in application logic

-- Operators can now be identified by either email or whatsapp_number
CREATE UNIQUE INDEX idx_operators_tenant_whatsapp ON operators(tenant_id, whatsapp_number) 
  WHERE whatsapp_number IS NOT NULL;

CREATE UNIQUE INDEX idx_operators_tenant_email ON operators(tenant_id, email) 
  WHERE email IS NOT NULL;
```

---

## Out of Scope (Deferred)

- WhatsApp number verification via OTP (relies on WhatsApp's existing verification)
- Two-factor authentication (MFA) via WhatsApp
- Group invitation flows (invite multiple operators at once)
- Custom WhatsApp message templates per tenant (enterprise feature)
- WhatsApp-based session management (login via WhatsApp click-to-chat)
