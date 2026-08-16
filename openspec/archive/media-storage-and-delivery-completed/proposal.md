# Change: Media storage, delivery, and dashboard uploads

## Why

Media is currently only half-wired. Outgoing media is sent to WhatsApp and then
optionally archived to S3, but four gaps make images effectively invisible in
the inbox and unusable without object storage:

1. The stored archive URL is `s3://bucket/key`, which is not a browser-loadable
   URL. The inbox `<img src>` therefore never renders, even with S3 configured.
2. Incoming media is reduced to a `[Image]`/`[Video]` placeholder and never
   downloaded or archived.
3. The dashboard's "Attach media" button is a no-op; there is no HTTP upload
   endpoint, so `media_path` (a server-local path) can never be produced by a
   client.
4. The service hard-requires AWS S3: with `S3_BUCKET` unset the archive is
   silently skipped and no media is ever persisted.

## What Changes

- One explicit hosting decision driven by configuration:
  - If the operator provides `S3_OBJECT_URL`, media is uploaded to S3 and every
    image is hosted there, served directly from that URL.
  - Otherwise media is written to the machine's local storage (`MEDIA_DIR`) and
    served by the application.
- Store a stable object key and resolve it to a browser-loadable URL at read
  time; stop persisting `s3://` URLs into `media_url`.
- Serve locally-stored media over authenticated, tenant-scoped HTTP endpoints
  for both the dashboard (session cookie) and the public API (API key / signed
  URL). S3-hosted media uses the configured object URL directly.
- Download and archive incoming media messages instead of discarding them.
- Add a real dashboard media upload flow (multipart endpoint + Paperclip wiring)
  with size limits and progress feedback.
- Add a `media_objects` table mapping object keys to tenants for authorization.

## Impact

- Affected specs: `media-storage-and-delivery`
- Affected code: `internal/storage`, `internal/upload`, `whatsapp`, `handler`,
  `config`, `frontend`, and a new migration
- API compatibility: existing send/receive endpoints keep their shape; new
  endpoints are additive. `media_url` values change from `s3://` references to
  either a public S3 URL or an app-served URL, which is the intended fix.
