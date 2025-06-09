#!/bin/bash

PSQL_DATABASE_URI="postgres://"

if [ -z "$PSQL_USER" ]; then
    echo "ERROR: PSQL_USER not set"
    exit 1
else
    PSQL_DATABASE_URI="$PSQL_DATABASE_URI$PSQL_USER"
fi

if [ -z "$PSQL_PASSWORD" ]; then
    echo "ERROR: PSQL_PASSWORD not set"
    exit 1
else
    PSQL_DATABASE_URI="$PSQL_DATABASE_URI:$PSQL_PASSWORD"
fi

if [ -z "$PSQL_HOST" ]; then
    echo "ERROR: PSQL_HOST not set"
    exit 1
else
    PSQL_DATABASE_URI="$PSQL_DATABASE_URI@$PSQL_HOST"
fi

if [ -z "$PSQL_PORT" ]; then
    echo "ERROR: PSQL_PORT not set"
    exit 1
else
    PSQL_DATABASE_URI="$PSQL_DATABASE_URI:$PSQL_PORT"
fi

if [ -z "$PSQL_DATABASE" ]; then
    echo "ERROR: PSQL_DATABASE not set"
    exit 1
else
    PSQL_DATABASE_URI="$PSQL_DATABASE_URI/$PSQL_DATABASE"
fi

if [ -z "$PSQL_SCHEMA" ]; then
    echo "ERROR: PSQL_SCHEMA not set"
    exit 1
else 
    PSQL_DATABASE_URI="$PSQL_DATABASE_URI?search_path=$PSQL_SCHEMA"
fi

if [ "$PSQL_SEARCH_PARAMS" ]; then
    for searchParam in $(echo $PSQL_SEARCH_PARAMS | cut -d' ' -f1-)
    do
        PSQL_DATABASE_URI="$PSQL_DATABASE_URI&$searchParam"
    done
fi

echo "Postgres Database URI: $PSQL_DATABASE_URI"