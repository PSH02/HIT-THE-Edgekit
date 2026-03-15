#!/usr/bin/env bash
set -euo pipefail

DATABASE_URL="${EDGEKIT_DATABASE_URL:-postgres://edgekit:edgekit@localhost:5432/edgekit?sslmode=disable}"
MIGRATIONS_DIR="internal/adapters/repository/postgres/migrations"

for f in "$MIGRATIONS_DIR"/*.sql; do
  echo "applying $(basename "$f")..."
  psql "$DATABASE_URL" -f "$f"
done

echo "migrations complete"
