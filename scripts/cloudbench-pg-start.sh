#!/usr/bin/env bash
# SPDX-FileCopyrightText: Kartoza
# SPDX-License-Identifier: MIT
# Start PostgreSQL with PostGIS for development
# Requires: POSTGRES_BIN_DIR to be set to the PostgreSQL bin directory
set -euo pipefail

: "${POSTGRES_BIN_DIR:?POSTGRES_BIN_DIR must be set}"

PGDATA="$PWD/.pgdata"
PGPORT="${PGPORT:-5432}"
PGDATABASE="${PGDATABASE:-cloudbench}"

if [ ! -d "$PGDATA" ]; then
    echo "🗄️  Initializing PostgreSQL..."
    "$POSTGRES_BIN_DIR/initdb" -D "$PGDATA" --auth=trust --no-locale --encoding=UTF8
    cat >> "$PGDATA/postgresql.conf" << EOF
listen_addresses = 'localhost'
port = $PGPORT
unix_socket_directories = '$PGDATA'
EOF
fi

if [ -f "$PGDATA/postmaster.pid" ]; then
    echo "✅ PostgreSQL already running"
    exit 0
fi

echo "🚀 Starting PostgreSQL..."
"$POSTGRES_BIN_DIR/pg_ctl" -D "$PGDATA" -l "$PGDATA/postgresql.log" -o "-k $PGDATA" start

for _ in {1..30}; do
    "$POSTGRES_BIN_DIR/pg_isready" -h "$PGDATA" -p "$PGPORT" > /dev/null 2>&1 && break
    sleep 0.5
done

if ! "$POSTGRES_BIN_DIR/psql" -h "$PGDATA" -p "$PGPORT" -lqt | cut -d \| -f 1 | grep -qw "$PGDATABASE"; then
    echo "📦 Creating database..."
    "$POSTGRES_BIN_DIR/createdb" -h "$PGDATA" -p "$PGPORT" "$PGDATABASE"
    "$POSTGRES_BIN_DIR/psql" -h "$PGDATA" -p "$PGPORT" -d "$PGDATABASE" -c "CREATE EXTENSION IF NOT EXISTS postgis;"
    "$POSTGRES_BIN_DIR/psql" -h "$PGDATA" -p "$PGPORT" -d "$PGDATABASE" -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
fi
echo "✅ PostgreSQL running (socket: $PGDATA)"
