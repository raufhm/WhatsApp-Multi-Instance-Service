# Security requirements for operator dashboard

## ADDED Requirements

### Requirement: Authentication uses secure cookies

Session cookies MUST be HttpOnly, Secure (in production), and SameSite=Lax.

#### Scenario: Login sets secure cookie

- **WHEN** an operator logs in
- **THEN** the server MUST set a cookie with HttpOnly, Secure, SameSite=Lax flags
- **AND** the cookie MUST have a reasonable expiry (e.g., 8 hours)

### Requirement: CSRF protection is enforced

All state-changing requests MUST include and validate a CSRF token.

#### Scenario: POST with CSRF token

- **WHEN** the dashboard makes a POST request
- **THEN** it MUST include a CSRF token (from meta tag or cookie)
- **AND** the server MUST validate it

### Requirement: Sensitive data is not logged

PII and message content MUST NOT appear in server logs.

#### Scenario: Login attempt logging

- **WHEN** logging authentication events
- **THEN** logs MUST NOT include passwords or full message bodies
- **AND** SHOULD only log email (masked), success/failure, timestamp

### Requirement: Session hijacking is mitigated

Sessions MUST be rotated on privilege changes and support revocation.

#### Scenario: Password change

- **WHEN** an operator changes their password
- **THEN** existing sessions MUST be invalidated
- **AND** the operator MUST log in again

#### Scenario: Admin revokes session

- **WHEN** an admin revokes a session
- **THEN** subsequent requests with that session cookie MUST fail

### Requirement: Rate limiting protects auth endpoints

Login and password endpoints MUST be rate-limited to prevent brute force.

#### Scenario: Repeated failed logins

- **WHEN** an IP exceeds failed login attempts
- **THEN** further attempts MUST be blocked temporarily
- **AND** the block MUST be logged for security monitoring
