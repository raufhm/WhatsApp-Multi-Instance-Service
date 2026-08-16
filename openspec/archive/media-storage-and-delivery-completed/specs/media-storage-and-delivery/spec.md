# Media storage and delivery

## ADDED Requirements

### Requirement: Media is stored without requiring object storage

The service MUST persist media to the machine's local storage when no S3 object
URL is configured, so images and files work out of the box.

#### Scenario: No S3 object URL configured

- **WHEN** `S3_OBJECT_URL` is unset
- **THEN** outgoing and incoming media MUST be persisted to `MEDIA_DIR`
- **AND** the inbox MUST render those images via the app-served URL

### Requirement: Media is hosted on S3 when an object URL is configured

The service MUST upload media to S3 and host it at the configured object URL
when that URL is provided.

#### Scenario: S3 object URL configured

- **WHEN** `S3_OBJECT_URL` is set (and `S3_BUCKET` is configured)
- **THEN** media MUST be uploaded to the S3 bucket
- **AND** `media_url` MUST be `{S3_OBJECT_URL}/{object_key}`
- **AND** images MUST be served directly from that URL

### Requirement: Media URLs are browser-loadable

The service MUST NOT expose `s3://`-style references to clients; it MUST store a
URL that a browser can load (either the S3 object URL or an app-served URL).

#### Scenario: Image renders in inbox

- **WHEN** a conversation contains an image message with an archived object
- **THEN** the timeline MUST return a browser-loadable URL
- **AND** the dashboard MUST render the image in the timeline

### Requirement: Locally-stored media serving is tenant-scoped

When media is served by the application, objects MUST be associated with a
tenant and served only to that tenant's authorized sessions or API keys.

#### Scenario: Cross-tenant access denied

- **WHEN** a request for another tenant's object key is received
- **THEN** the service MUST return 404 (not 403) and MUST NOT stream the object

### Requirement: Incoming media is archived

Incoming image/video/audio/file messages MUST be downloaded and archived rather
than discarded.

#### Scenario: Incoming image

- **WHEN** a contact sends an image
- **THEN** the service MUST download and store the payload
- **AND** the resulting message MUST have a browser-loadable media URL

### Requirement: Dashboard media upload works

Operators MUST be able to attach and send media from the reply composer with
size enforcement and progress feedback.

#### Scenario: Attach and send an image

- **WHEN** an operator attaches a file and sends it
- **THEN** the file MUST be uploaded via multipart, persisted, and delivered to
  the contact
- **AND** the composer MUST show upload progress and a cancel control
