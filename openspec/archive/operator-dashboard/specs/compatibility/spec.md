# Browser compatibility requirements for operator dashboard

## ADDED Requirements

### Requirement: Modern browsers are supported

The dashboard MUST work on current versions of major browsers.

#### Scenario: Chrome/Edge

- **WHEN** using Chrome or Edge (current - 2 versions)
- **THEN** all features MUST work correctly

#### Scenario: Firefox

- **WHEN** using Firefox (current - 2 versions)
- **THEN** all features MUST work correctly

#### Scenario: Safari

- **WHEN** using Safari (current - 2 versions, including iOS Safari)
- **THEN** all features MUST work correctly

### Requirement: Legacy browser detection

Users on unsupported browsers MUST see a compatibility notice.

#### Scenario: Old browser

- **WHEN** a user visits with an unsupported browser
- **THEN** a notice MUST appear recommending an upgrade
- **AND** core functionality MAY be degraded but not broken

### Requirement: Graceful degradation

Features that require modern APIs MUST degrade gracefully.

#### Scenario: No Service Worker

- **WHEN** running on a browser without Service Worker support
- **THEN** the app MUST still function (without offline caching)
- **AND** not throw errors

#### Scenario: No ResizeObserver

- **WHEN** running on a browser without ResizeObserver
- **THEN** layout MUST still work (with less dynamic responsiveness)
- **AND** not throw errors
