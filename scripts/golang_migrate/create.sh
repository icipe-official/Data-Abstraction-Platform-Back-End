#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source $SCRIPT_DIR/init_psql_database_migrations.sh

while getopts m: flag
do
    case "${flag}" in
        m) PATH_TO_MIGRATE=${OPTARG};;
    esac
done

if [ -z "$PATH_TO_MIGRATE" ]; then
    PATH_TO_MIGRATE=$SCRIPT_DIR/../../bin/migrate
fi

MIGRATION_NAME=$1
if [ -z "$MIGRATION_NAME" ]; then
    echo "ERROR: Please pass the migration name as the first argument"
    exit 1
fi

echo "Migration Name: $MIGRATION_NAME"

$PATH_TO_MIGRATE create -ext sql -dir $PSQL_DATABASE_MIGRATION_SCRIPTS_DIRECTORY $MIGRATION_NAME