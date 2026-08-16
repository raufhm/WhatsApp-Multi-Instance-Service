# Change: Make outgoing media uploads reliable

## Why

Outgoing media currently uploads directly to S3 and ignores the returned error. A transient S3 failure can therefore leave a sent message without its archived media while giving the caller no durable way to recover. This is the first outstanding item in the project's roadmap.

## What Changes

- Persist failed outgoing-media uploads as retryable jobs.
- Retry transient failures with bounded exponential backoff.
- Mark jobs as permanently failed after the configured retry limit and retain the error for diagnosis.
- Keep successful uploads idempotent so retries do not create duplicate media records.
- Expose operational logging and metrics for queued, retried, completed, and permanently failed uploads.

## Impact

- Affected specs: `media-upload-reliability`
- Affected code: `internal/storage`, `whatsapp`, persistence migrations, and configuration
- API compatibility: Existing message-send behavior remains unchanged for successful uploads; upload failures become observable and recoverable instead of being discarded.
