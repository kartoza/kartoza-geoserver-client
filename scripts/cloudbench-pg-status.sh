#!/usr/bin/env bash
# SPDX-FileCopyrightText: Kartoza
# SPDX-License-Identifier: MIT
# Check PostgreSQL status
# Requires: POSTGRES_BIN_DIR to be set to the PostgreSQL bin directory
set -euo pipefail

: "${POSTGRES_BIN_DIR:?POSTGRES_BIN_DIR must be set}"

PGDATA="$PWD/.pgdata"
PGPORT="${PGPORT:-5432}"
PGDATABASE="${PGDATABASE:-cloudbench}"

if [ -f "$PGDATA/postmaster.pid" ]; then
    echo "PostgreSQL is running"
    echo ""
    echo "Socket:   $PGDATA"
    echo "Port:     $PGPORT"
    echo "Database: $PGDATABASE"
    echo ""
    # Check if PostgreSQL is ready
    if "$POSTGRES_BIN_DIR/pg_isready" -h "$PGDATA" -p "$PGPORT" > /dev/null 2>&1; then
        echo "Status: READY"
        echo ""
        # Show PostGIS version
        "$POSTGRES_BIN_DIR/psql" -h "$PGDATA" -p "$PGPORT" -d "$PGDATABASE" -c "SELECT PostGIS_Version();" 2>/dev/null || true
    else
        echo "Status: STARTING (wait a moment...)"
    fi
else
    echo "PostgreSQL is not running"
    echo "Start with: cloudbench-pg-start"
fi
