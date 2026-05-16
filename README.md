# WhatsApp API Service

Modular, multi-instance WhatsApp API for team communication and automation.

## Purpose
Enables central management of multiple WhatsApp accounts. Automates interactions, tracks team conversations, and synchronizes group state via unofficial API.

## Features
- **Multi-Instance**: Manage multiple personal/business accounts from single service.
- **Persistence**: PostgreSQL integration for session management, message history, receipts.
- **Automation**: Human-like presence states (typing) and randomized inter-message delays.
- **Media & Reactions**: Send/receive images, files, and emoji reactions.
- **Group Sync**: Auto-sync metadata, participants (phone numbers), and history.
- **Reliability**: Auto-reconnect, persistent session state, multi-device support.

## Tech Stack
- **Language**: Go
- **WhatsApp API**: `whatsmeow`
- **Database**: PostgreSQL (`pgx`)
- **Migrations**: `golang-migrate`
- **Config**: `Viper`

## Quick Start
1. **Env**: Set `PG_DSN` and `PORT`.
2. **Build**: `go build -o whatsapp-api .`
3. **Run**: `./whatsapp-api`
4. **Onboard**: `POST /api/onboard` with `{"email": "..."}`. Scan terminal QR.

## Media Handling Limitations
- **Memory**: Loads file to RAM. Use streams for files >50MB.
- **Expiration**: WhatsApp `DirectPaths` temporary. Cache `MediaKey` + `FileHash` for re-download.
- **MIME**: Rely on `http.DetectContentType`. Complex files need explicit extension mapping.
