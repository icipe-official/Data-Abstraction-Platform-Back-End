#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKING_DIR="$SCRIPT_DIR/../../"

if [ ! -f "$WORKING_DIR/bin/migrate" ]; then
    echo "Downloading golang migrate executable..."
    bash $WORKING_DIR/scripts/golang_migrate/download.sh
    echo "...download complete"
fi

echo "Building cmd applications..."
    bash $WORKING_DIR/scripts/build_cmd_app_create_super_user.sh
    bash $WORKING_DIR/scripts/build_cmd_app_init_database.sh
echo "...building cmd applications complete"

echo "Creating executable for cmd/web_service..."
go build -C $WORKING_DIR/cmd/web_service/ -o $WORKING_DIR/bin/web_service
echo "...complete"