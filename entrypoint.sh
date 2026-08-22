#!/bin/sh
set -e

# Migrations are handled programmatically by the application when APP_ENV=dev.
# Production deployments should run migrations via the CI/CD pipeline.

echo "Starting application..."
exec "$@"
