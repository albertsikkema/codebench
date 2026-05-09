#!/bin/bash
DIR="$(cd "$(dirname "$0")" && pwd)"
PLATFORM="$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"
BINARY="$DIR/code-index-server-$PLATFORM"

if [ ! -f "$BINARY" ]; then
    echo "Error: No binary for platform '$PLATFORM'" >&2
    echo "Download from: https://github.com/albertsikkema/mcp-index-server/releases" >&2
    exit 1
fi

exec "$BINARY"
