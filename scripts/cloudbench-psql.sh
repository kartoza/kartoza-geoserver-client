#!/usr/bin/env bash
# SPDX-FileCopyrightText: Kartoza
# SPDX-License-Identifier: MIT
# Connect to PostgreSQL with psql
# Requires: POSTGRES_BIN_DIR to be set to the PostgreSQL bin directory
set -euo pipefail

: "${POSTGRES_BIN_DIR:?POSTGRES_BIN_DIR must be set}"

PGDATA="$PWD/.pgdata"
PGPORT="${PGPORT:-5432}"
PGDATABASE="${PGDATABASE:-cloudbench}"

exec "$POSTGRES_BIN_DIR/psql" -h "$PGDATA" -p "$PGPORT" -d "$PGDATABASE" "$@"
