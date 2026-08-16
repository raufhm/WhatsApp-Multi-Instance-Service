# Operator dashboard error and accessibility states

## ADDED Requirements

### Requirement: Loading states provide feedback

Every data-dependent UI element MUST show a loading indicator while fetching.

#### Scenario: Inbox loading

- **WHEN** conversations are being fetched
- **THEN** the inbox MUST display skeleton rows or a spinner

#### Scenario: Conversation detail loading

- **WHEN** message timeline is being fetched
- **THEN** the detail page MUST show a loading indicator
- **AND** disable the reply composer until loaded

### Requirement: Empty states guide the user

When a list or section has no items, the UI MUST show an illustrative empty state with a call-to-action where appropriate.

#### Scenario: No conversations

- **WHEN** the inbox has no matching conversations
- **THEN** it MUST display "No tickets match your filters"
- **AND** a button to clear filters

#### Scenario: No activities

- **WHEN** the activity queue is empty
- **THEN** it MUST display "No pending activities"

### Requirement: Errors are recoverable

API failures MUST display human-readable messages with retry options where appropriate.

#### Scenario: Network error on reply

- **WHEN** a reply fails due to network error
- **THEN** the dashboard MUST show an error toast
- **AND** revert the optimistic update
- **AND** allow retry

### Requirement: Permission denied is handled

Requests that return 403 MUST redirect to a permission-denied page.

#### Scenario: Unauthorized access

- **WHEN** an operator accesses a conversation they don't have permission for
- **THEN** the dashboard MUST show a 403 page

### Requirement: Keyboard navigation works

All interactive elements MUST be focusable and operable via keyboard.

#### Scenario: Tab through inbox

- **WHEN** an operator tabs through the inbox
- **THEN** focus MUST move logically between filter controls, table rows, and actions

#### Scenario: Send via Enter

- **WHEN** focus is in the reply composer
- **THEN** pressing Enter MUST send the message (with Shift+Enter for newline)

### Requirement: Screen readers can navigate

All interactive elements MUST have appropriate ARIA labels and roles.

#### Scenario: Screen reader announces conversation

- **WHEN** a screen reader focuses a conversation row
- **THEN** it MUST announce ticket number, contact, status, and last activity

#### Scenario: Message bubble semantics

- **WHEN** reading a message timeline
- **THEN** inbound/outbound/internal messages MUST be distinguishable by role or label
