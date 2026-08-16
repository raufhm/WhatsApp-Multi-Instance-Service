# Archived: Media Storage, Delivery, and Dashboard Uploads

**Archived**: August 16, 2026
**Status**: ✅ FULLY IMPLEMENTED

This change has been fully implemented, built, and tested. The specification is
preserved here for historical reference.

## What Was Implemented

A pluggable media pipeline that hosts images on S3 when `S3_OBJECT_URL` is
configured, and otherwise writes them to local machine storage (`MEDIA_DIR`),
served by the application. Incoming media is now downloaded and archived instead
of discarded, and the dashboard reply composer can upload and send attachments.

### Storage abstraction
- `internal/storage/media.go` — `MediaStore` interface (`Put` / `Open` / `Delete`) + `ResolveMediaURL`
- `internal/storage/disk.go` — `DiskStore` (local filesystem, path-traversal-safe)
- `internal/storage/s3.go` — `S3Storage` implements `MediaStore` (Put/Get/Delete via `S3ClientAPI`), retains legacy `Upload` for the retryable worker
- `internal/storage/media_objects.go` — `RecordMediaObject` / `GetMediaObject` (tenant-scoped lookup)

### Configuration & wiring
- `config/config.go` — `S3_OBJECT_URL`, `MEDIA_DIR`
- `main.go` — selects `S3Storage` when `S3_BUCKET` is set, otherwise `DiskStore`; wires `MediaStore`/`S3ObjectURL` into manager, API server, and dashboard API

### Serving endpoints
- `handler/http.go` — `GET /api/v1/media/{key}` (API-key auth, tenant-scoped)
- `handler/dashboard_api.go` — `GET /dashboard/api/media/{key}` (session auth) and `POST /dashboard/api/media` (multipart upload, 50MB cap)

### Incoming media
- `whatsapp/subsystem.go` — downloads image/video/audio/document payloads via `client.Download`, stores them, and resolves a served URL

### Outgoing media
- `whatsapp/subsystem.go` — `sendMedia` reads from `MediaStore` by `media_key` (or legacy `media_path`), resolves the media URL, and attaches it to the projected message

### Frontend
- `frontend/src/lib/apiClient.ts` — `mediaApi.uploadMedia` (multipart with progress)
- `frontend/src/hooks/useInbox.ts` — `useUploadMedia`, `useSendMessage` sends `media_key`
- `frontend/src/pages/ConversationDetail.tsx` — Paperclip wired to a file input, upload progress/cancel, attachment preview, and `resolveMediaUrl` for `<img>` rendering

### Migration
- `migrations/0008_media_objects.up.sql` / `.down.sql`

## Acceptance Criteria Met

- `go build ./...` passes ✅
- `go test ./...` passes ✅
- `npm run build` (frontend `tsc && vite build`) passes ✅
- `S3_OBJECT_URL` set → `media_url` is `{S3_OBJECT_URL}/{key}` ✅
- `S3_OBJECT_URL` unset → media served from `MEDIA_DIR` via app endpoints ✅

## Related

- `../reliable-s3-media-uploads-completed/` — durable retryable S3 archive worker (not yet archived)
- `../operator-permissions-completed/` — RBAC (archived)
- `../../changes/agent-implementation-plan/` — active implementation plan
