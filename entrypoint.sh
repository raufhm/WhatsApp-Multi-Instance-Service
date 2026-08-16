#!/bin/sh
set -e

# Wait for PostgreSQL to be ready
until pg_isready -h db -p 5432 -U whatsapp; do
  echo "Waiting for PostgreSQL..."
  sleep 1
done

# Run migrations
echo "Running migrations..."
migrate -path ./migrations -database "$PG_DSN" -verbose up || true

# Start the application
echo "Starting application..."
exec "$@"
