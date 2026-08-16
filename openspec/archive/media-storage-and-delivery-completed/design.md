# Design: Media storage, delivery, and dashboard uploads

## Hosting decision

A single configuration value decides where media lives and how it is served:

| Configuration | Backend | `media_url` stored | Who serves bytes |
| --- | --- | --- | --- |
| `S3_OBJECT_URL` set | S3 | `{S3_OBJECT_URL}/{key}` | S3 / CDN directly |
| `S3_OBJECT_URL` unset | Local disk (`MEDIA_DIR`) | app URL e.g. `/api/v1/media/{key}` | The application |

When `S3_OBJECT_URL` is provided, all images are hosted there and the app does
not proxy them. When it is absent, images are written to the machine and the app
serves them over HTTP.

## Storage abstraction

Introduce a backend interface that decouples storage from serving:

```go
type MediaStore interface {
    Put(ctx context.Context, key, mimeType string, data []byte) error
    Open(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
}
```

Two implementations:

- `DiskStore` (`internal/storage/disk.go`) writes to `MEDIA_DIR/{key}` using a
  flat, sanitized key space. `Open` is used by the app's serving endpoint.
- `S3Store` (evolved from the current `S3Storage`) does PutObject/GetObject/
  DeleteObject. `Open` exists for completeness; the hosted case is served
  directly from S3 via `S3_OBJECT_URL`.

The object key is the stable reference (already `media/<uuid>`).

## URL resolution

`conversation_messages.media_url` stores a browser-loadable URL:

- S3 backend: `{S3_OBJECT_URL}/{key}`. Object keys are unguessable UUIDs, so a
  public-read bucket or CDN prefix is effectively capability-based access. If
  the bucket is private, presigned URLs are generated at read time (noted in
  tasks).
- Disk backend: a served relative URL. The dashboard frontend maps object keys
  to `/dashboard/api/media/{key}` (cookie-auth); API clients map them to
  `/api/v1/media/{key}`.

## Serving and authorization (local disk only)

Add a `media_objects` table:

```sql
CREATE TABLE media_objects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    object_key TEXT NOT NULL UNIQUE,
    mime_type TEXT NOT NULL DEFAULT '',
    size BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Every stored object is registered with its owning tenant. Serving handlers look
up the object by key, enforce the tenant boundary, then stream bytes with the
stored `mime_type`.

Two endpoints, one per client surface:

- `GET /api/v1/media/{key}` — public API, authenticated by `X-API-Key` header
  (matching the rest of the API). For `<img>`/download use, also accept a
  short-lived signed token query parameter, since an `<img>` tag cannot send
  headers.
- `GET /dashboard/api/media/{key}` — dashboard, authenticated by the session
  cookie. Because the cookie is scoped to `Path=/dashboard`, the browser's
  `<img src>` sends credentials automatically for this mount.

Range requests are optional for v1 (full responses are acceptable); noted in
tasks.

## Incoming media download

In `handleIncomingMessage`, when the parsed type is `IMAGE`/`VIDEO`/`AUDIO`/
`FILE`, download the payload via whatsmeow `client.Download`, persist it through
`MediaStore`, register the `media_objects` row, and set `media_url` on the
dispatched `MessageMetadata`. The download runs off the event handler goroutine
(the existing `go i.handleIncomingMessage` already isolates it); media bytes and
paths are never logged.

## Dashboard media upload

Add `POST /dashboard/api/media` (multipart) that:

- requires an authenticated session,
- enforces a configured maximum size,
- reads the file, detects MIME, stores it via `MediaStore`,
- returns `{ media_key, media_url, mime_type, size }`.

Wire the Paperclip button to a hidden file input with progress/cancel, and
change `useSendMessage`/`sendMedia` to send by media key instead of a
server-local `media_path`. `sendMedia` reads bytes from the media store rather
than `os.ReadFile(req.MediaPath)`.

## Configuration

- `S3_OBJECT_URL` (optional) — public base URL for hosted media, e.g.
  `https://my-bucket.s3.amazonaws.com` or a CDN/CloudFront prefix. When set,
  media is uploaded to the S3 bucket named by `S3_BUCKET` and hosted at this
  URL.
- `S3_BUCKET` (required only when `S3_OBJECT_URL` is set) — the bucket the S3
  store writes to.
- `MEDIA_DIR` (default `./media`) — local disk directory used when
  `S3_OBJECT_URL` is unset.

Backend selection in `main.go`: `S3_OBJECT_URL != ""` → `S3Store`, otherwise
`DiskStore`.

## Migration

One new migration `0008_media_objects` adds the `media_objects` table. No
breaking changes to existing tables.
