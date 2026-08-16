# Notifications and alerts requirements for operator dashboard

## ADDED Requirements

### Requirement: Toast notifications provide feedback

User actions and system events MUST show non-blocking toast notifications.

#### Scenario: Action success

- **WHEN** an action succeeds (e.g., assign conversation)
- **THEN** a success toast MUST appear briefly
- **AND** auto-dismiss after 3-5 seconds

#### Scenario: Action failure

- **WHEN** an action fails
- **THEN** an error toast MUST appear
- **AND** include a retry option where applicable
- **AND** persist until dismissed

### Requirement: In-app alerts highlight important events

Critical system events MUST be highlighted in the UI.

#### Scenario: New conversation

- **WHEN** a new inbound message creates a conversation
- **THEN** the inbox row MUST be highlighted
- **AND** a badge or indicator MAY appear

#### Scenario: Activity requires attention

- **WHEN** an activity is created
- **THEN** the activity queue MUST show a count badge
- **AND** highlight new items

### Requirement: Browser notifications are optional

Users MAY opt into browser push notifications for high-priority events.

#### Scenario: Enable push notifications

- **WHEN** a user opts in to push notifications
- **THEN** the browser MUST prompt for permission
- **AND** notifications MUST only trigger for configured events (e.g., new conversation when idle)

#### Scenario: Respect Do Not Disturb

- **WHEN** the user has enabled Do Not Disturb or set quiet hours
- **THEN** push notifications MUST be suppressed
- **AND** in-app alerts MAY still appear
