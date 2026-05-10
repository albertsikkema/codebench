#!/bin/bash
# Codebench bootstrap installer.
# Downloads the repo tarball and runs install-helper.sh against the caller's cwd.
#
#   curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash -s -- --branch v0.1.0
#   curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash -s -- --dry-run /path/to/project
set -e

REPO="albertsikkema/codebench"
BRANCH="main"
ORIGINAL_DIR="$(pwd)"
INSTALL_ARGS=()

usage() {
    cat << EOF
Codebench bootstrap installer

Usage: install.sh [OPTIONS] [TARGET_DIR]

OPTIONS:
    --branch <ref>      Install from a branch or tag (default: main)
    --dry-run           Forwarded: show what would be installed
    --help, -h          Show this help

ARGUMENTS:
    TARGET_DIR          Forwarded: install into this dir (default: current dir)

EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --branch)
            [ -z "${2:-}" ] && { echo "Error: --branch needs a value" >&2; exit 1; }
            BRANCH="$2"; shift 2 ;;
        --help|-h) usage ;;
        *) INSTALL_ARGS+=("$1"); shift ;;
    esac
done

for cmd in curl tar; do
    command -v "$cmd" >/dev/null 2>&1 || { echo "Error: $cmd is required" >&2; exit 1; }
done

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading $REPO@$BRANCH..."
# codeload serves both branches and tags via refs/heads/<x> or refs/tags/<x>; try heads first, then tags.
URL_HEAD="https://codeload.github.com/${REPO}/tar.gz/refs/heads/${BRANCH}"
URL_TAG="https://codeload.github.com/${REPO}/tar.gz/refs/tags/${BRANCH}"
if ! curl -fsSL "$URL_HEAD" -o "$TMP_DIR/src.tar.gz" 2>/dev/null; then
    curl -fsSL "$URL_TAG" -o "$TMP_DIR/src.tar.gz" \
        || { echo "Error: could not download $REPO@$BRANCH" >&2; exit 1; }
fi

tar -xz -C "$TMP_DIR" -f "$TMP_DIR/src.tar.gz"
EXTRACTED="$(find "$TMP_DIR" -maxdepth 1 -mindepth 1 -type d | head -1)"
[ -d "$EXTRACTED" ] || { echo "Error: tarball extraction failed" >&2; exit 1; }
[ -f "$EXTRACTED/install-helper.sh" ] || { echo "Error: install-helper.sh missing in tarball" >&2; exit 1; }

CODEBENCH_TARGET_DIR="$ORIGINAL_DIR" \
CODEBENCH_VERSION="$BRANCH" \
    bash "$EXTRACTED/install-helper.sh" "${INSTALL_ARGS[@]}"
