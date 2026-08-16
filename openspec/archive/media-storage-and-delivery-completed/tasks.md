# Tasks

## Phase 1 — Storage abstraction

- [ ] Define the `MediaStore` interface (`Put` / `Open` / `Delete`).
- [ ] Implement `DiskStore` with `MEDIA_DIR` and safe key handling.
- [ ] Evolve `S3Storage` to implement `MediaStore` (`GetObject` / `DeleteObject`).
- [ ] Add the `media_objects` migration and repository methods.
- [ ] Add `S3_OBJECT_URL` / `MEDIA_DIR` configuration and wire backend selection in `main.go` (S3 when `S3_OBJECT_URL` is set, otherwise disk).

## Phase 2 — URL resolution and serving

- [ ] Resolve `media_url` to `{S3_OBJECT_URL}/{key}` when S3 is configured.
- [ ] Implement tenant-scoped `GET /api/v1/media/{key}` (API key auth) for local disk.
- [ ] Implement `GET /dashboard/api/media/{key}` (session cookie auth) for local disk.
- [ ] Update frontend to render resolved URLs for both hosting modes.

## Phase 3 — Incoming media

- [ ] Download incoming image/video/audio/file payloads via whatsmeow.
- [ ] Persist incoming media and set `media_url` on projected messages.

## Phase 4 — Dashboard upload

- [ ] Add `POST /dashboard/api/media` multipart upload with size limit.
- [ ] Wire the Paperclip button, file input, progress, and cancel in `ConversationDetail`.
- [ ] Update `useSendMessage` / `sendMedia` to send by media key.

## Phase 5 — Tests and docs

- [ ] Unit tests for `DiskStore`, `S3Store`, serving authz, and URL resolution.
- [ ] Integration tests for upload → serve → inbox display (both disk and S3 modes).
- [ ] Update README and configuration docs (`S3_OBJECT_URL`, `MEDIA_DIR`).
