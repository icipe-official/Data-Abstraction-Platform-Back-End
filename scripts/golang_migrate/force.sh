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

if [ -z "$PATH_TO_MIGRATE" ]; then
    PATH_TO_MIGRATE=$SCRIPT_DIR/../../bin/migrate
fi

MIGRATION_VERSION=$1
if [ -z "$MIGRATION_VERSION" ]; then
    echo "ERROR: Please set a version using the flag -v"
    exit 1
fi

echo "Force Version: $MIGRATION_VERSION"

$PATH_TO_MIGRATE -path $PSQL_DATABASE_MIGRATION_SCRIPTS_DIRECTORY -database $PSQL_DATABASE_URI force $MIGRATION_VERSION