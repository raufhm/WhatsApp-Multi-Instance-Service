# WhatsApp Multi-Instance Service (POC)

High-performance, multi-account WhatsApp API gateway. Designed for automation, conversation tracking, and unified messaging.

**Disclaimer:** Personal portfolio project. Currently under active development (Proof of Concept). Built for research and development purposes.

## Core Solution
Managing multiple WhatsApp business/personal accounts at scale is difficult. This service provides a unified HTTP interface to control multiple instances simultaneously, ensuring message persistence and human-simulated interaction patterns.

## Current Capabilities
- **Multi-Instance Management**: Spawn and control multiple WhatsApp accounts (Instances) concurrently.
- **Message Persistence**: Reliable storage of messages, delivery receipts, and group metadata in PostgreSQL.
- **Rich Media Support**: Send and receive Images, Videos, Audio, and Documents.
- **Media Archiving**: Automatic upload of outgoing media to AWS S3 for persistent URLs.
- **Human-Simulated Interaction**: Configurable "composing" states and randomized delays to mimic human behavior.
- **Group Intelligence**: Automatic tracking of group participants, ownership, and history synchronization.
- **Session Reliability**: Persistent device pairing using `whatsmeow` with auto-reconnect logic.

## Technical Architecture
- **Language**: Go (Golang)
- **Engine**: `whatsmeow` (Signal protocol implementation)
- **Storage**: PostgreSQL for structured data, AWS S3 for media assets.
- **Deployment**: Docker-ready with environment-based configuration (Viper).
- **Migration**: Automated schema management via `golang-migrate`.

## Project Structure
- `whatsapp/`: Core instance management and Signal protocol orchestration.
- `handler/`: HTTP API layer for onboarding and messaging.
- `internal/storage/`: Clean abstraction for Postgres and S3 persistence.
- `domain/`: Shared models and interface definitions.

## Getting Started (Dev Mode)
1. **Configure**: Set `PG_DSN`, `S3_BUCKET`, and `PORT` in environment.
2. **Build**: `go build -o whatsapp-api .`
3. **Run**: `./whatsapp-api`
4. **Onboard**: `POST /api/onboard` → Scan QR code in terminal to pair device.

## Roadmap & Known Limitations
- [ ] Implement retry queues for failed S3 uploads.
- [ ] Add support for incoming media auto-download to S3.
- [ ] Move file processing from RAM to streaming for large assets (>50MB).
- [ ] Advanced message filtering and webhook dispatchers.
