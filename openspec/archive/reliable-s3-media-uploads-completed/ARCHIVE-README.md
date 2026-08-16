# Archived: Reliable S3 Media Uploads

**Archived**: August 16, 2026
**Status**: ✅ FULLY IMPLEMENTED

This change has been fully implemented, built, and tested. The specification is
preserved here for historical reference.

## What Was Implemented

A durable, retryable S3 archive worker for outgoing media. Failed uploads become
recoverable jobs instead of being silently discarded, with bounded exponential
backoff, idempotent object keys, and operational logging.

### Domain & repository
- `domain/models.go` — `UploadJob`, `UploadJobStatus`, `UploadJobRepository`
- `migrations/0003_upload_jobs.up.sql` / `.down.sql` — `upload_jobs` table + due/host indexes
- `internal/storage/postgres.go` — `CreateUploadJob`, `ClaimDueUploadJobs` (atomic `FOR UPDATE SKIP LOCKED` lease), `MarkUploadCompleted`, `MarkUploadRetryable`, `MarkUploadFailed`, `ListUploadJobs`, `AttachMediaURL`, `TenantForHost`

### Worker
- `internal/upload/upload.go` — `Manager` (Enqueue/Start/Wait/poll/process/finalize) with `MediaStore` boundary and `MediaURLSetter` hook
- `internal/upload/backoff.go` — bounded exponential backoff with jitter + transient/permanent error classification
- `internal/upload/upload_test.go` — retry classification, backoff, idempotent keys, terminal failures

### Wiring & config
- `main.go` — creates and starts the worker only when `S3_BUCKET` is set, wires tenant resolution, graceful shutdown via `Wait`
- `config/config.go` — `UPLOAD_WORKER_ENABLED`, `UPLOAD_POLL_INTERVAL`, `UPLOAD_MAX_ATTEMPTS`, `UPLOAD_INITIAL_BACKOFF`, `UPLOAD_MAX_BACKOFF`, `UPLOAD_LEASE`, `UPLOAD_JITTER`
- `whatsapp/subsystem.go` — enqueues a durable archive job after a successful send

## Job lifecycle

```text
PENDING -> PROCESSING -> COMPLETED
              |
              +-> PENDING (retryable error)
              +-> FAILED (retry limit or permanent error)
```

## Acceptance Criteria Met

- Failed uploads persisted as retryable jobs with next-attempt time ✅
- Retries bounded (configurable max attempts), atomic claims prevent concurrent processing ✅
- Object key generated once and reused across attempts (idempotent) ✅
- `go build ./...` and `go test ./...` pass ✅

## Related

- `../media-storage-and-delivery-completed/` — media storage/serving/upload (archived)
- `../../changes/agent-implementation-plan/` — active implementation plan
