# Media upload reliability

## ADDED Requirements

### Requirement: Failed uploads are durable

The service MUST persist an outgoing-media upload job before or atomically with handing the upload to asynchronous processing, including the tenant, message, object key, MIME type, and retry state.

#### Scenario: S3 is temporarily unavailable

- **WHEN** an outgoing media upload receives a transient S3 or network error
- **THEN** the service MUST retain a retryable job with the error and a future next-attempt time
- **AND** the original media message MUST NOT be treated as if its archive URL were available

### Requirement: Retries are bounded and scheduled

The service MUST retry transient failures using configurable exponential backoff, MUST avoid concurrent processing of the same job, and MUST stop retrying after the configured maximum attempt count.

#### Scenario: A retry eventually succeeds

- **WHEN** a queued job is claimed after its next-attempt time and S3 accepts the upload
- **THEN** the service MUST mark the job completed and persist the resulting media URL

#### Scenario: A job exceeds the retry limit

- **WHEN** all configured attempts fail with transient errors
- **THEN** the service MUST mark the job permanently failed
- **AND** the service MUST retain the last error and attempt count for diagnosis

### Requirement: Upload retries are idempotent

The service MUST reuse the same object key for every attempt of a given upload job.

#### Scenario: The process exits after S3 success

- **WHEN** S3 has accepted an upload but the process exits before completion is persisted
- **THEN** a later retry MUST reuse the original key and converge to one completed archive object

### Requirement: Upload outcomes are observable

The service MUST emit structured logs for queueing, retrying, completion, and permanent failure, including identifiers and attempt information but never media contents.

#### Scenario: An operator investigates failed uploads

- **WHEN** an upload reaches a terminal failure
- **THEN** logs and persisted job state MUST identify the affected message, job, object key, and final error
