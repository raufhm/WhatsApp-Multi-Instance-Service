# Internationalization requirements for operator dashboard

## ADDED Requirements

### Requirement: UI text is externalized

All user-facing strings MUST be extracted into translation files.

#### Scenario: English default

- **WHEN** the dashboard loads with no locale preference
- **THEN** it MUST display English text

#### Scenario: Locale detection

- **WHEN** the browser reports a preferred language
- **THEN** the dashboard SHOULD attempt to load matching translations
- **AND** fall back to English if unavailable

### Requirement: Date/time formatting is locale-aware

Timestamps MUST be formatted according to the user's locale.

#### Scenario: Last activity timestamp

- **WHEN** displaying "Last activity"
- **THEN** the timestamp MUST use locale-aware formatting (e.g., "2 hours ago" vs "2 Stunden")

#### Scenario: Date filters

- **WHEN** selecting a date range in filters
- **THEN** the date picker MUST use locale-appropriate format

### Requirement: Right-to-left (RTL) layout is supported

Languages that read right-to-left MUST have mirrored layouts.

#### Scenario: Arabic or Hebrew locale

- **WHEN** the UI is set to an RTL language
- **THEN** text alignment, padding, margins, and flex directions MUST be mirrored
- **AND** icons that imply direction (arrows) MUST flip

### Requirement: Number formatting is locale-aware

Large numbers (ticket counts, activity counts) MUST use locale grouping.

#### Scenario: Ticket number display

- **WHEN** displaying counts
- **THEN** numbers MUST use locale-appropriate grouping (e.g., 1,000 vs 1.000)
