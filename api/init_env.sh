#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

PATH_TO_ENV="$SCRIPT_DIR/env.sh"

if [ ! -f "$PATH_TO_ENV" ]; then
    echo "ERROR: $PATH_TO_ENV does not exist."
    exit 1
fi

source $PATH_TO_ENV