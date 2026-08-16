# Testing requirements for operator dashboard

## ADDED Requirements

### Requirement: Unit tests cover components and hooks

Critical components and custom hooks MUST have unit test coverage.

#### Scenario: DataTable component

- **WHEN** the DataTable receives props
- **THEN** it MUST render correctly
- **AND** handle sorting, filtering, pagination

#### Scenario: useConversation hook

- **WHEN** the hook fetches conversation data
- **THEN** it MUST return loading, data, error states correctly

### Requirement: E2E tests cover critical user flows

Key user journeys MUST have end-to-end test coverage.

#### Scenario: Login → Inbox → Conversation → Reply

- **GIVEN** a test operator account
- **WHEN** the test logs in, navigates to inbox, opens a conversation, and sends a reply
- **THEN** the test MUST verify each step succeeds
- **AND** the reply appears in the timeline

#### Scenario: Assign conversation

- **WHEN** the test assigns a conversation to another operator
- **THEN** it MUST verify the assignee updates

### Requirement: Accessibility tests validate ARIA

Critical pages MUST pass automated accessibility checks.

#### Scenario: Axe-core scan

- **WHEN** running automated accessibility tests
- **THEN** inbox, conversation detail, and login pages MUST have no critical violations

### Requirement: Visual regression tests catch unintended changes

Key pages MUST have visual snapshots to detect regressions.

#### Scenario: Inbox visual snapshot

- **WHEN** the inbox renders with sample data
- **THEN** the snapshot MUST match the baseline
- **AND** alert on unexpected visual changes
