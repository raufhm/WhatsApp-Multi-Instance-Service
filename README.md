# WhatsApp Multi-Instance Service

> A self-hosted WhatsApp inbox for small teams that want to keep the number customers already know.

This project explores a simple but useful idea: what if a small business could keep its existing WhatsApp number, while letting multiple teammates help answer messages from one shared dashboard?

Instead of turning WhatsApp into a giant marketing platform, this project focuses on the day-to-day reality of customer communication:

- one familiar number
- separate logins for teammates
- one shared inbox
- notes, follow-ups, and handoffs
- enough structure to keep conversations moving

It is built for teams that need an operational workspace, not a broadcast cannon.

> **Disclaimer:** This is a personal open-source research project. Linking accounts through WhatsApp Web/companion protocols may violate WhatsApp's Terms of Service. Use it at your own risk and follow the rules that apply to your account and region.

## Why this exists

Many small businesses do not need a full enterprise messaging suite on day one. They need to answer customers faster, keep context in one place, and avoid losing important details across personal phones.

This project is for that stage.

It helps a team:

- keep using the WhatsApp number customers already recognize
- share one inbox across multiple authorized users
- reply from the same connected number
- view group names and participants more clearly
- assign conversations when ownership changes
- leave internal notes that customers should not see
- track follow-ups so tasks do not disappear
- keep basic contact and deal-stage history together

The goal is not to make WhatsApp louder. The goal is to make the work behind the messages easier.

## What it is not

This project is intentionally not a bulk marketing or broadcast tool.

It is also not Meta's official Cloud API, and it does not try to replace the BSP ecosystem. If you need approved templates, campaign tooling, or enterprise messaging infrastructure, this is probably not the right fit.

## What works today

The current build supports:

- WhatsApp account connection through QR pairing
- a dashboard for incoming and outgoing conversations
- sending messages and media
- direct-chat replies and reactions
- group chat visibility and participant phone numbers
- conversation assignment and closing
- internal notes and follow-ups
- contacts, conversations, activities, and deal stages
- simple deterministic rules for first responses and handoffs

This is intentionally practical and mostly manual in the right places. A human still makes decisions, but the project gives those decisions a place to live.

## No AI yet

There is no AI assistant in the product today, by design.

The current focus is on learning the real workflows first: what operators record, where conversations get stuck, and which parts are repetitive enough to improve later. If AI is added in the future, it should support those habits instead of replacing them.

## Tech Stack

- Go backend
- React dashboard
- PostgreSQL for persistence
- `whatsmeow` for WhatsApp connectivity
- optional S3-compatible storage for media

## Quick Start

### Prerequisites

- Go 1.25 or newer
- PostgreSQL 14 or newer
- Docker and Docker Compose, recommended for local runs
- Optional S3-compatible storage for media durability

### Run with Docker Compose

1. Copy the example environment file:

```sh
cp .env.example .env
```

2. Set `PG_DSN` in `.env` to point at your PostgreSQL database.

3. Start the app:

```sh
docker compose up --build -d
```

4. Watch logs if you want to confirm everything is healthy:

```sh
docker compose logs -f app
```

Then open:

```text
http://localhost:8080/dashboard
```

### Build locally

If you prefer running the frontend and backend separately:

```sh
cd frontend
npm install
npm run build

cd ../backend
go run .
```

The frontend builds into `backend/dist/` and is served by the Go backend.

## Configuration

Start with `.env.example`, then copy it to `.env` and fill in the values you need.

Useful settings include:

- `PG_DSN` for PostgreSQL
- `PORT` for the HTTP server
- `LOG_LEVEL` for logging verbosity
- `BOT_SESSION_TIMEOUT` for session cleanup behavior
- `S3_BUCKET` and related settings for optional media storage

If `S3_BUCKET` is left empty, media can fall back to local disk storage.

## Project Status

This is an early v1 and a personal project, not a polished enterprise platform.

Expect trade-offs around:

- WhatsApp protocol changes or rate limits
- account restrictions or linked-device behavior
- manual CRM workflows
- self-hosting and backup responsibility
- media availability after remote URLs expire

That said, the core goal is stable and clear: help a small team operate the WhatsApp account it already has.

## Contributing

Contributions, issue reports, and honest feedback are very welcome.

Good places to help:

- onboarding and setup polish
- reliability and reconnect handling
- observability and error reporting
- tests and regression coverage
- documentation and examples
- dashboard UX improvements

If you spot a rough edge, feel free to open an issue or send a PR. Even small fixes help a lot.

## A Small Note From The Project

This started as an experiment in shared WhatsApp operations and has grown into a conversation inbox, a lightweight CRM, and a pile of lessons about group metadata, media expiry, retries, and the limits of building on top of a linked app account.

It is not trying to replace every product in the WhatsApp ecosystem.

It is trying to be useful to the person who says:

> We already have the number. We already have the customers. We just need the team to work together.

Built with Go, React, PostgreSQL, and [whatsmeow](https://github.com/tulir/whatsmeow).
