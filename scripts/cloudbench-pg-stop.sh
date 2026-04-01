#!/usr/bin/env bash
# SPDX-FileCopyrightText: Kartoza
# SPDX-License-Identifier: MIT
# Stop PostgreSQL development server
# Requires: POSTGRES_BIN_DIR to be set to the PostgreSQL bin directory
set -euo pipefail

: "${POSTGRES_BIN_DIR:?POSTGRES_BIN_DIR must be set}"

PGDATA="$PWD/.pgdata"

if [ ! -f "$PGDATA/postmaster.pid" ]; then
    echo "PostgreSQL is not running"
    exit 0
fi

echo "🛑 Stopping PostgreSQL..."
"$POSTGRES_BIN_DIR/pg_ctl" -D "$PGDATA" stop
echo "✅ PostgreSQL stopped"
