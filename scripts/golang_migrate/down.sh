#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source $SCRIPT_DIR/init_postgres_database_uri.sh
source $SCRIPT_DIR/init_psql_database_migrations.sh

while getopts m: flag
do
    case "${flag}" in
        m) PATH_TO_MIGRATE=${OPTARG};;
    esac
done

if [ -z "$PSQL_DATABASE_MIGRATION_SCRIPTS_DIRECTORY" ]; then
    PSQL_DATABASE_MIGRATION_SCRIPTS_DIRECTORY=$SCRIPT_DIR/../../database/psql_migrations
fi

if [ -z "$PATH_TO_MIGRATE" ]; then
    PATH_TO_MIGRATE=$SCRIPT_DIR/../../bin/migrate
fi

NO_OF_MIGRATIONS=$1
echo "Up: $NO_OF_MIGRATIONS"

$PATH_TO_MIGRATE -path $PSQL_DATABASE_MIGRATION_SCRIPTS_DIRECTORY -database $PSQL_DATABASE_URI down $NO_OF_MIGRATIONS