#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKING_DIR="$SCRIPT_DIR/../../"

while getopts t:c: flag
do
    case "${flag}" in
        t) CONTAINER_TAG=${OPTARG};;
        c) CONTAINER_CLI=${OPTARG};;
    esac
done
if [ -z "$CONTAINER_TAG" ]; then
    CONTAINER_TAG="latest"
fi

if [ -z "$CONTAINER_CLI" ]; then
    CONTAINER_CLI='docker'
fi

echo "Using command '$CONTAINER_CLI' to build container image..."
eval $CONTAINER_CLI build --no-cache -t data_abstraction_platform/web_service:$CONTAINER_TAG -f $WORKING_DIR/build/Dockerfile.web_service $WORKING_DIR
echo "... container complete"