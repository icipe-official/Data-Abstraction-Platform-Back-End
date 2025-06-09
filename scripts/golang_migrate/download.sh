#!/bin/bash

set -e # terminate script if commands fail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/../../bin"

mkdir -p $BIN_DIR

while getopts v:o:a: flag
do
    case "${flag}" in
        v) VERSION=${OPTARG};;
        o) OS=${OPTARG};;
        a) ARCH=${OPTARG};;
    esac
done
if [ -z "$VERSION" ]; then
    VERSION="v4.18.1"
fi
if [ -z "$OS" ]; then
    OS="linux"
fi
if [ -z "$ARCH" ]; then
    ARCH="amd64"
fi

COMPRESS_TYPE="tar.gz"

# Download golang migrate and place the executable in the bin directory
curl -L https://github.com/golang-migrate/migrate/releases/download/$VERSION/migrate.$OS-$ARCH.$COMPRESS_TYPE | tar xvz -C $BIN_DIR