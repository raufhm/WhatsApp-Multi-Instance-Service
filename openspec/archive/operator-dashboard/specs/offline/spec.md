# Offline and resilience requirements for operator dashboard

## ADDED Requirements

### Requirement: Offline detection is visible

When the dashboard loses network connectivity, users MUST be notified.

#### Scenario: Network disconnect

- **WHEN** the browser goes offline
- **THEN** a banner or toast MUST appear indicating offline status
- **AND** disable actions that require network

#### Scenario: Network reconnect

- **WHEN** connectivity is restored
- **THEN** the offline notice MUST disappear
- **AND** pending actions MAY retry automatically

### Requirement: Form data is preserved

Unsubmitted form data MUST not be lost on navigation or refresh.

#### Scenario: Reply composer

- **WHEN** typing a reply and navigating away
- **THEN** the draft MUST be saved to sessionStorage
- **AND** restored when returning to the conversation

#### Scenario: Page refresh

- **WHEN** refreshing the page with unsent data
- **THEN** the data MUST be preserved

### Requirement: Failed actions can be retried

Actions that fail due to network errors MUST offer retry.

#### Scenario: Failed message send

- **WHEN** a reply fails to send
- **THEN** an error indicator MUST appear next to the message
- **AND** a Retry button MUST be available

#### Scenario: Bulk retry

- **WHEN** multiple actions have failed
- **THEN** a "Retry all" option MAY be available
