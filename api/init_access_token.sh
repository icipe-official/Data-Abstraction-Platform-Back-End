#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

PATH_TO_TOKEN_JSON="$SCRIPT_DIR/iam/login/output.json"

if [ ! -f "$PATH_TO_TOKEN_JSON" ]; then
    echo "ERROR: $PATH_TO_TOKEN_JSON does not exist."
    exit 1
fi

ACCESS_TOKEN=$(cat $PATH_TO_TOKEN_JSON | jq -r '.token.access_token')

if [ -z "$ACCESS_TOKEN"  ]; then
    echo "ERROR: Please set ACCESS_TOKEN found in $PATH_TO_TOKEN_JSON"
    exit 1
fi
