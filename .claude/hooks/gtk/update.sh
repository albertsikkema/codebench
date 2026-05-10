#!/bin/bash
# Update gtk binaries from the latest GitHub release
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
REPO="albertsikkema/gtk"

cd "$DIR"
rm -f gtk-*
gh release download --repo "$REPO" --pattern 'gtk-*' --skip-existing
# Remove checksums file if downloaded
rm -f checksums.txt
chmod +x gtk-*

echo "Updated to $(gh release view --repo "$REPO" --json tagName -q .tagName):"
ls -lh gtk-*
