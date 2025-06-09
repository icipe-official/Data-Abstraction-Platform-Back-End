#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

PSQL_DATABASE_MIGRATION_SCRIPTS_DIRECTORY=$SCRIPT_DIR/../../database/psql_migrations

echo "Path to Postgres Database Migrations: $PSQL_DATABASE_MIGRATION_SCRIPTS_DIRECTORY"