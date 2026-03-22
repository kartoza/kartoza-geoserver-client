#!/usr/bin/env bash
# SPDX-FileCopyrightText: Kartoza
# SPDX-License-Identifier: MIT
# Run database migrations
set -euo pipefail

PGDATA="$PWD/.pgdata"
PGPORT="${PGPORT:-5432}"
PGDATABASE="${PGDATABASE:-cloudbench}"

# Build the migration command
cd "$PWD"

# Set DATABASE_URL for the Go application
export DATABASE_URL="postgres://$USER:@localhost:$PGPORT/$PGDATABASE?host=$PGDATA"

echo "🔄 Running migrations..."
go run ./cmd/migrate up
echo "✅ Migrations complete"
