# Dashboard inbox access

## ADDED Requirements

### Requirement: Operators can read and write the inbox via their session

A logged-in operator MUST be able to list conversations, load timelines and
activities, and send messages using their dashboard session, without needing an
API key.

#### Scenario: Inbox loads after sign-in

- **WHEN** an authenticated operator opens the inbox
- **THEN** the service MUST accept their session and return their tenant's
  conversations
- **AND** the dashboard MUST render the conversation list without a 401

#### Scenario: Send a reply

- **WHEN** an authenticated operator sends a message from the conversation view
- **THEN** the service MUST resolve the operator's tenant from the session and
  deliver the message

### Requirement: API-key access is preserved

External clients MUST continue to authenticate with an API key.

#### Scenario: API key still works

- **WHEN** a client calls `/api/v1/*` with a valid `X-API-Key`
- **THEN** the service MUST resolve the tenant from the key and serve the request

### Requirement: Session and key access are tenant-scoped

A session or key MUST only ever resolve to its own tenant, and invalid or
expired credentials MUST be rejected.

#### Scenario: Expired session

- **WHEN** a request carries an expired or invalid session cookie and no API key
- **THEN** the service MUST return 401

#### Scenario: Cross-tenant access

- **WHEN** a session for tenant A requests tenant B's conversation id
- **THEN** the service MUST NOT return tenant B's data

### Requirement: Operator identity is recorded

Actions performed via a session MUST record the operator's identity for audit
and actor fields.

#### Scenario: Assign via session

- **WHEN** an operator assigns a conversation via the dashboard
- **THEN** the audit/actor attribution MUST reflect that operator, not `"api"`
