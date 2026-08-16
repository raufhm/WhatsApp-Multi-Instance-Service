# Operator dashboard responsive design

## ADDED Requirements

### Requirement: Mobile layout prioritizes conversation detail

On small screens, the UI MUST adapt to show one primary view at a time.

#### Scenario: Mobile inbox

- **WHEN** viewed on mobile
- **THEN** the inbox MUST use a single-column list
- **AND** hide less critical columns (account, priority) or collapse them

#### Scenario: Mobile conversation

- **WHEN** viewing a conversation on mobile
- **THEN** the composer MUST be fixed at the bottom
- **AND** the timeline MUST scroll above it

### Requirement: Tablet layout uses split view

On medium screens, the UI SHOULD use a master-detail layout.

#### Scenario: Tablet inbox + detail

- **WHEN** viewed on tablet
- **THEN** the inbox and conversation detail SHOULD appear side-by-side
- **AND** selecting a conversation updates the detail pane without navigation

### Requirement: Desktop layout maximizes information density

On large screens, the UI MAY show additional metadata and controls.

#### Scenario: Desktop inbox

- **WHEN** viewed on desktop
- **THEN** all columns SHOULD be visible
- **AND** filters MAY appear in a sidebar or top bar

### Requirement: Breakpoints are consistent

Breakpoint values (mobile/tablet/desktop) MUST be consistent across all pages.

#### Scenario: Consistent responsive behavior

- **WHEN** resizing the browser
- **THEN** all pages MUST transition at the same breakpoints
- **AND** maintain visual consistency
