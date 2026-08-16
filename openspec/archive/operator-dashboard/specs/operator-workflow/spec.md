# Operator dashboard authentication

## ADDED Requirements

### Requirement: Operators authenticate with email and password

The dashboard MUST require authentication before accessing any tenant data or operator actions.

#### Scenario: Login succeeds with valid credentials

- **GIVEN** a registered operator
- **WHEN** they POST to `/dashboard/login` with correct email and password
- **THEN** the server MUST set an HttpOnly session cookie
- **AND** redirect to the inbox (`/dashboard/inbox`)

#### Scenario: Login fails with incorrect credentials

- **WHEN** incorrect credentials are submitted
- **THEN** the server MUST return 401
- **AND** the dashboard MUST display an error message

### Requirement: Sessions expire and can be revoked

The server MUST store session expiry and validate it on every request.

#### Scenario: Session expires

- **GIVEN** an active session
- **WHEN** the session reaches its expiry time
- **THEN** subsequent requests MUST redirect to `/login`

#### Scenario: Logout clears the session

- **WHEN** an operator logs out
- **THEN** the session MUST be deleted server-side
- **AND** the cookie MUST be cleared client-side

---

# Inbox UI

## ADDED Requirements

### Requirement: Inbox lists conversations with key metadata

The inbox MUST display ticket number, contact name/number, account, status, priority, assignee, and last activity timestamp.

#### Scenario: Inbox loads conversations

- **WHEN** the inbox page mounts
- **THEN** it MUST fetch conversations from `/api/v1/conversations`
- **AND** render them in a sortable, filterable table

#### Scenario: Inbox supports status filtering

- **WHEN** an operator selects one or more statuses
- **THEN** the table MUST only show matching conversations
- **AND** the URL MUST reflect the active filters

### Requirement: Inbox shows activity queue

The inbox MUST include a section for pending activities with acknowledge actions.

#### Scenario: Activity acknowledge succeeds

- **WHEN** an operator clicks Acknowledge on an activity
- **THEN** the dashboard MUST POST to `/api/v1/activities/:id/acknowledge`
- **AND** remove the activity from the queue

---

# Conversation detail

## ADDED Requirements

### Requirement: Conversation detail shows message timeline

The detail page MUST render the full message history in chronological order with clear visual distinction between inbound, outbound, and internal messages.

#### Scenario: Internal notes are styled differently

- **GIVEN** a conversation with internal notes
- **THEN** notes MUST be visually distinct (e.g., gray background)
- **AND** MUST NOT appear to be customer-facing messages

### Requirement: Reply composer sends messages via the API

The detail page MUST include a composer that submits to `/api/v1/accounts/:account/messages`.

#### Scenario: Text reply is sent

- **GIVEN** an open conversation
- **WHEN** the operator types and submits a reply
- **THEN** the message MUST appear in the timeline immediately (optimistic update)
- **AND** POST to the message endpoint
- **AND** handle errors by reverting the optimistic update

#### Scenario: Media upload is supported

- **WHEN** the operator attaches a file
- **THEN** the dashboard MUST upload it (either direct to S3 or via backend proxy)
- **AND** send a media message

### Requirement: Ticket controls modify conversation state

The detail page MUST provide actions to assign, handoff, close, and reopen.

#### Scenario: Assign operator

- **WHEN** an operator assigns the conversation to a teammate
- **THEN** the dashboard MUST POST to `/api/v1/operator/assign?id=:conv`
- **AND** update the conversation assignee in the UI

---

# Contact management

## ADDED Requirements

### Requirement: Contact detail shows metadata and history

The contact page MUST display name, number, tags, and linked conversations.

#### Scenario: Edit contact metadata

- **WHEN** an operator updates contact fields
- **THEN** the dashboard MUST PATCH to the contact endpoint
- **AND** reflect the changes immediately

---

# Bot rule management

## ADDED Requirements

### Requirement: Bot rules editor validates before publishing

The editor MUST validate rule configuration and allow saving drafts.

#### Scenario: Publish ruleset

- **WHEN** an operator publishes a ruleset
- **THEN** the dashboard MUST POST to `/api/v1/bot-rules` (create)
- **AND** POST to `/api/v1/bot-rules/activate?version=N` (activate)
- **AND** show success feedback

#### Scenario: Rollback to previous version

- **WHEN** an operator rolls back
- **THEN** the dashboard MUST activate the previous version
- **AND** show confirmation
