#!/bin/bash
# Update code-index-server binaries from the latest GitHub release
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
REPO="albertsikkema/mcp-index-server"

cd "$DIR"
rm -f code-index-server-*
gh release download --repo "$REPO" --pattern 'code-index-server-*'
chmod +x code-index-server-*

echo "Updated to $(gh release view --repo "$REPO" --json tagName -q .tagName):"
ls -lh code-index-server-*
