# Performance requirements for operator dashboard

## ADDED Requirements

### Requirement: Initial load is fast

The dashboard MUST load and become interactive within acceptable time bounds.

#### Scenario: Cold load on 3G

- **WHEN** loading the dashboard on a slow connection
- **THEN** Time to Interactive (TTI) MUST be under 5 seconds
- **AND** the critical path (auth + inbox) MUST load first

#### Scenario: Bundle size

- **WHEN** building the frontend
- **THEN** the total JavaScript bundle MUST be under 500KB gzipped
- **AND** use code splitting to load routes on demand

### Requirement: Inbox renders efficiently

The inbox table MUST handle hundreds of conversations without jank.

#### Scenario: Large inbox

- **WHEN** the inbox has 500+ conversations
- **THEN** scrolling MUST remain smooth (60 FPS)
- **AND** filtering MUST respond within 200ms

#### Scenario: Virtualized list

- **WHEN** the conversation list is very long
- **THEN** the table SHOULD use virtualization to render only visible rows

### Requirement: Message timeline scrolls smoothly

The conversation detail MUST handle long message histories.

#### Scenario: Long conversation

- **WHEN** a conversation has 1000+ messages
- **THEN** scrolling the timeline MUST remain smooth
- **AND** media previews MUST lazy-load

### Requirement: Network requests are optimized

API calls MUST use appropriate caching and batching.

#### Scenario: React Query caching

- **WHEN** navigating between conversations
- **THEN** previously loaded data MUST be served from cache
- **AND** background revalidation MAY occur

#### Scenario: Debounced search

- **WHEN** typing in a search/filter input
- **THEN** API requests MUST be debounced (e.g., 300ms)
- **AND** cancel in-flight requests when input changes
