#!/bin/bash

# Setup Environment variables

# Global
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PATH_TO_INIT_ENV="$SCRIPT_DIR/../../init_env.sh"
source $PATH_TO_INIT_ENV

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Local
PATH_TO_ENV="$SCRIPT_DIR/env.sh"

if [ -f "$PATH_TO_ENV" ]; then
    source $PATH_TO_ENV
fi



# Build cURL command

CURL_COMMAND="curl"

OPTIONS=$(getopt -o '' -l "verbose,output-json" -n "script_name" -- "$@")

if [[ $? -ne 0 ]]; then
  echo "Error in command line arguments." >&2
  exit 1
fi

VERBOSE=false
while true; do
    case "$1" in
        --verbose)
            CURL_COMMAND="$CURL_COMMAND -v"
            VERBOSE=true
            shift
            ;;
        --output-json)
            CURL_COMMAND="$CURL_COMMAND -o $SCRIPT_DIR/output.json"
            shift
            ;;
        --)
            shift
            break
            ;;
        *)
            break
            ;;
    esac
done

if [ -z "$HTTP_PATH" ]; then
    HTTP_PATH="/iam/session-data"
fi

CURL_COMMAND="$CURL_COMMAND -X GET --url $WEB_SERVICE_API_CORE_URL$HTTP_PATH -H 'Accept: application/json'"

if [ "$VERBOSE" = true ]; then
    echo "cURL command begin..."
    echo
    echo $CURL_COMMAND
    echo
    echo "...cURL commmand end"
    echo
fi

# Execute cURL command

eval $CURL_COMMAND

echo
