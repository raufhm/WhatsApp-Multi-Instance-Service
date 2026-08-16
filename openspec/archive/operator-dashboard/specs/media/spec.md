# Media handling requirements for operator dashboard

## ADDED Requirements

### Requirement: Media previews are safe

Media attachments MUST be previewed safely without automatic download.

#### Scenario: Image preview

- **WHEN** a conversation contains images
- **THEN** thumbnails MUST be shown in the timeline
- **AND** full-size view MUST open in a modal or new tab
- **AND** use `loading="lazy"` for below-the-fold images

#### Scenario: Video preview

- **WHEN** a conversation contains videos
- **THEN** a poster frame or play button MUST be shown
- **AND** playback MUST require explicit user action

#### Scenario: Document preview

- **WHEN** a conversation contains documents (PDF, DOC)
- **THEN** a file icon and name MUST be shown
- **AND** download MUST require explicit click
- **AND** warn before opening potentially unsafe file types

### Requirement: Media upload provides feedback

Uploading media from the reply composer MUST show progress.

#### Scenario: Upload progress

- **WHEN** attaching a file
- **THEN** a progress bar or spinner MUST appear
- **AND** show percentage or "uploading..." status

#### Scenario: Upload cancel

- **WHEN** an upload is in progress
- **THEN** a Cancel button MUST be available
- **AND** abort the upload and remove the attachment

### Requirement: Media size limits are enforced

Large files MUST be rejected before upload begins.

#### Scenario: File too large

- **WHEN** selecting a file exceeding the limit
- **THEN** the dashboard MUST show an error immediately
- **AND** not attempt upload

### Requirement: Multiple media attachments

The composer MUST support attaching multiple files.

#### Scenario: Multiple attachments

- **WHEN** selecting multiple files
- **THEN** all attachments MUST appear in the composer
- **AND** each MUST have individual remove and progress indicators
