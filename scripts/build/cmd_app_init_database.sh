#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKING_DIR="$SCRIPT_DIR/../../"

mkdir -p $WORKING_DIR/bin

echo "Creating executable for cmd/cmd_app_init_database..."
go build -C $WORKING_DIR/cmd/cmd_app_init_database/ -o $WORKING_DIR/bin/cmd_app_init_database
echo "...complete"