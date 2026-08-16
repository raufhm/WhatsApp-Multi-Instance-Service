# Build stage: compile Go backend and build frontend assets
FROM node:22-slim AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install --no-audit --no-fund
COPY frontend/ .
RUN npm run build

FROM golang:1.25-alpine AS go-builder
WORKDIR /app/backend
RUN apk add --no-cache git
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend-builder /app/backend/dist ./dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/whatsapp-service .

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl postgresql-client

# Install golang-migrate for database migrations
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar -C /usr/local/bin -xvz migrate || true

WORKDIR /app
COPY --from=go-builder /app/whatsapp-service .
COPY --from=go-builder /app/backend/migrations ./migrations
COPY --from=frontend-builder /app/backend/dist ./dist
COPY entrypoint.sh ./
RUN chmod +x ./entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./whatsapp-service"]
