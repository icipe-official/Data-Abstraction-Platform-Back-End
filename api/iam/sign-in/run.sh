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

PATH_TO_TOKEN_JSON="$SCRIPT_DIR/input_token.json"
if [ ! -f "$PATH_TO_TOKEN_JSON" ]; then
    echo "ERROR: $PATH_TO_TOKEN_JSON does not exist."
    exit 1
fi

ACCESS_TOKEN=$(cat $PATH_TO_TOKEN_JSON | jq -r '.access_token')
if [ -z "$ACCESS_TOKEN"  ]; then
    echo "ERROR: Please set ACCESS_TOKEN found in $PATH_TO_TOKEN_JSON"
    exit 1
fi
ACCESS_TOKEN_EXPIRY=$(cat $PATH_TO_TOKEN_JSON | jq -r '.expires_in')
if [ -z "$ACCESS_TOKEN_EXPIRY"  ]; then
    echo "ERROR: Please set ACCESS_TOKEN_EXPIRY found in $PATH_TO_TOKEN_JSON"
    exit 1
fi

REFRESH_TOKEN=$(cat $PATH_TO_TOKEN_JSON | jq -r '.refresh_token')
if [ -z "$REFRESH_TOKEN"  ]; then
    echo "ERROR: Please set REFRESH_TOKEN found in $PATH_TO_TOKEN_JSON"
    exit 1
fi
REFRESH_TOKEN_EXPIRY=$(cat $PATH_TO_TOKEN_JSON | jq -r '.refresh_expires_in')
if [ -z "$REFRESH_TOKEN_EXPIRY"  ]; then
    echo "ERROR: Please set REFRESH_TOKEN_EXPIRY found in $PATH_TO_TOKEN_JSON"
    exit 1
fi


if [ -z "$HTTP_PATH" ]; then
    HTTP_PATH="/iam/sign-in"
fi

CURL_COMMAND="$CURL_COMMAND -X GET --url $WEB_SERVICE_API_CORE_URL$HTTP_PATH -H 'Accept: application/json'"
CURL_COMMAND="$CURL_COMMAND -H 'OpenID-Access-Token: $ACCESS_TOKEN'"
CURL_COMMAND="$CURL_COMMAND -H 'OpenID-Access-Token-Expires-In: $ACCESS_TOKEN_EXPIRY'"
CURL_COMMAND="$CURL_COMMAND -H 'OpenID-Refresh-Token: $REFRESH_TOKEN'"
CURL_COMMAND="$CURL_COMMAND -H 'OpenID-Refresh-Token-Expires-In: $REFRESH_TOKEN_EXPIRY'"

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
