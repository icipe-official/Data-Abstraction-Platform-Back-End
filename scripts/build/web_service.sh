#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKING_DIR="$SCRIPT_DIR/../../"

mkdir -p $WORKING_DIR/bin

echo "Creating executable for cmd/web_service..."
go build -C $WORKING_DIR/cmd/web_service/ -o $WORKING_DIR/bin/web_service
echo "...complete"