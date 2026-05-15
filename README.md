# WhatsApp Multi-Instance Service

A production-ready WhatsApp service for managing multiple numbers, automating workflows, and monitoring team communications — built in Go with `whatsmeow`.

## Overview

Designed for teams that need a single, reliable gateway across multiple WhatsApp numbers. Handles session persistence, message history, group sync, and automation without the overhead of managing separate instances manually.

## Features

- **Multi-Instance Management** — Onboard and operate multiple WhatsApp numbers from a single service
- **Session Persistence** — PostgreSQL-backed session storage with full message history and receipt tracking
- **Group Sync** — Auto-sync group metadata, participant lists, and historical messages on connect
- **Media Support** — Send and receive images, documents, and emoji reactions
- **Human-Like Automation** — Randomized delays and typing states for natural interaction patterns
- **Clean REST API** — Simple HTTP endpoints for onboarding and messaging operations

## Tech Stack

| Layer      | Tool                |
|------------|---------------------|
| Language   | Go                  |
| WhatsApp   | `whatsmeow`         |
| Database   | PostgreSQL (`pgx`)  |
| Migration  | `golang-migrate`    |
| Config     | Viper               |

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 14+

### Setup

```bash
# Clone and build
git clone https://github.com/yourorg/whatsapp-service
cd whatsapp-service
go build -o whatsapp-api .

# Configure
export PG_DSN="postgres://user:pass@localhost:5432/whatsapp"
export PORT=8080

# Run
./whatsapp-api
```

### Onboard a Number

```bash
POST /api/onboard
Content-Type: application/json

{ "email": "user@example.com" }
```

Scan the QR code printed in the terminal. The session is persisted automatically — no re-scan needed on restart.

## API Reference

| Method | Endpoint              | Description                   |
|--------|-----------------------|-------------------------------|
| `POST` | `/api/onboard`        | Register a new WhatsApp number |
| `GET`  | `/api/instances`      | List active instances         |
| `POST` | `/api/send`           | Send a message                |
| `GET`  | `/api/messages/:jid`  | Fetch message history         |

## Project Structure

```
.
├── cmd/            # Entry point
├── internal/
│   ├── instance/   # WhatsApp instance lifecycle
│   ├── store/      # PostgreSQL layer
│   └── api/        # HTTP handlers
├── migrations/     # SQL migration files
└── config/         # Viper config setup
```

## Contributing

PRs welcome. Open an issue first for significant changes.

## License

MIT