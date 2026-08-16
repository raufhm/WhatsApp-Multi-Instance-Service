# Operator dashboard real-time updates

## ADDED Requirements

### Requirement: Inbox refreshes periodically

The inbox MUST poll for updates at a configurable interval (default 30 seconds).

#### Scenario: New conversation appears

- **WHEN** a new conversation is created (via inbound message)
- **THEN** the next poll MUST fetch it
- **AND** the inbox MUST update without full page reload

#### Scenario: Conversation state changes

- **WHEN** a conversation is closed or assigned by another operator
- **THEN** the next poll MUST reflect the change
- **AND** the UI MUST update the row

### Requirement: Optimistic updates provide immediate feedback

User actions that modify server state MUST update the UI optimistically before the API response.

#### Scenario: Assign conversation

- **WHEN** an operator assigns a conversation
- **THEN** the assignee field MUST update immediately
- **AND** revert if the API call fails

#### Scenario: Send reply

- **WHEN** an operator sends a reply
- **THEN** the message MUST appear in the timeline immediately
- **AND** show a "sending..." indicator until confirmed
- **AND** revert on error

### Requirement: Manual refresh is available

Users MUST be able to trigger an immediate refresh.

#### Scenario: Pull to refresh (mobile)

- **WHEN** the user pulls down on mobile
- **THEN** the inbox MUST refresh immediately

#### Scenario: Refresh button

- **WHEN** the user clicks the refresh button
- **THEN** all data MUST be refetched
