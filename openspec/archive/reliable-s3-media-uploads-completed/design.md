# Design: Reliable outgoing-media uploads

## Approach

Introduce a durable upload-job repository backed by PostgreSQL and a worker owned by the application process. The message/media flow writes the upload job with a deterministic object key, attempts the upload, and records the outcome. A periodic worker claims due jobs, performs bounded retries, and updates their status atomically.

## Job lifecycle

```text
PENDING -> PROCESSING -> COMPLETED
              |              
              +-> PENDING (retryable error)
              +-> FAILED (retry limit or permanent error)
```

Each job stores tenant and message identifiers, object key, MIME type, media payload (or a durable payload reference), attempt count, next-attempt time, last error, and timestamps. A lease or equivalent claim mechanism prevents multiple workers from processing the same job concurrently.

## Retry policy

- Retry only transient storage/network failures.
- Use configurable maximum attempts and exponential backoff with a cap.
- Add jitter to avoid synchronized retries across instances.
- Treat missing configuration, invalid object keys, and other known permanent errors as `FAILED` immediately.

## Consistency and recovery

The object key is generated once and reused for every attempt. Completion is recorded only after S3 confirms the upload. If the process exits after S3 succeeds but before the database update, the next attempt overwrites the same key and safely converges to `COMPLETED`.

## Observability

Log job ID, message ID, object key, attempt number, and outcome without logging media bytes. Counters and latency measurements should distinguish successful first attempts, retries, and terminal failures.

## Configuration

Add settings for worker enablement, poll interval, maximum attempts, initial backoff, and maximum backoff. Preserve safe defaults so existing deployments do not require configuration changes to start.
