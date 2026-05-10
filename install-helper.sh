#!/bin/bash
# Codebench installer - copies .claude/ and .mcp.json into a target repo.
# Invoked by install.sh, but also runnable directly from a clone:
#
#   ./install-helper.sh                       # install into current dir
#   ./install-helper.sh --dry-run             # preview only
#   ./install-helper.sh /path/to/project      # install into a specific dir
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

DRY_RUN=false
TARGET_DIR="${CODEBENCH_TARGET_DIR:-.}"

VERSION="${CODEBENCH_VERSION:-}"
if [ -z "$VERSION" ]; then
    VERSION=$(cd "$SCRIPT_DIR" && git rev-parse --short HEAD 2>/dev/null || echo "unknown")
fi

usage() {
    cat << EOF
Codebench installer v${VERSION}

Usage: $0 [OPTIONS] [TARGET_DIR]

OPTIONS:
    --dry-run           Show what would be installed without making changes
    --help, -h          Show this help

ARGUMENTS:
    TARGET_DIR          Target directory (default: current directory)

EOF
    exit 0
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=true; shift ;;
        --help|-h) usage ;;
        -*) echo -e "${RED}Error: unknown option $1${NC}" >&2; usage ;;
        *) TARGET_DIR="$1"; shift ;;
    esac
done

print_message() { echo -e "${1}${2}${NC}"; }
print_header() { echo ""; echo -e "${BLUE}=== $1 ===${NC}"; }

install_file() {
    local src="$1" dest="$2" desc="$3"
    if [ "$DRY_RUN" = true ]; then
        print_message "$GREEN" "  [DRY RUN] would install: $desc"
        return 0
    fi
    mkdir -p "$(dirname "$dest")"
    cp "$src" "$dest"
    print_message "$GREEN" "  ok $desc"
}

update_git_exclude() {
    local exclude_path="$TARGET_DIR/.git/info/exclude"
    local entries=(".claude/index" ".claude/index-cache" ".claude/logs" ".claude/memories" ".playwright/" ".playwright-mcp/")

    if [ ! -d "$TARGET_DIR/.git" ]; then
        print_message "$YELLOW" "  not a git repo, skipping git exclude"
        return 0
    fi

    if [ "$DRY_RUN" = true ]; then
        print_message "$YELLOW" "  [DRY RUN] would update .git/info/exclude"
        return 0
    fi

    mkdir -p "$TARGET_DIR/.git/info"
    [ ! -f "$exclude_path" ] && touch "$exclude_path"

    for entry in "${entries[@]}"; do
        if ! grep -qxF "$entry" "$exclude_path" 2>/dev/null; then
            echo "$entry" >> "$exclude_path"
            print_message "$GREEN" "  ok added: $entry"
        fi
    done
}

main() {
    print_header "Codebench installer v${VERSION}"

    if [ ! -d "$SCRIPT_DIR/.claude" ]; then
        print_message "$RED" "Error: .claude/ not found in $SCRIPT_DIR"
        exit 1
    fi

    mkdir -p "$TARGET_DIR"
    TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"
    print_message "$BLUE" "Target: $TARGET_DIR"
    [ "$DRY_RUN" = true ] && print_message "$YELLOW" "DRY RUN MODE"

    # Install .claude/ recursively. Exclude runtime dirs that are gitignored anyway.
    print_header "Installing .claude/"
    while IFS= read -r -d '' f; do
        rel_path="${f#"$SCRIPT_DIR/"}"
        install_file "$f" "$TARGET_DIR/$rel_path" "$rel_path"
    done < <(find "$SCRIPT_DIR/.claude" -type f \
        -not -path '*/.venv/*' \
        -not -path '*/index/*' \
        -not -path '*/index-cache/*' \
        -not -path '*/logs/*' \
        -not -path '*/memories/*' \
        -print0)

    if [ "$DRY_RUN" != true ]; then
        find "$TARGET_DIR/.claude" -name "*.sh" -exec chmod +x {} + 2>/dev/null || true
        find "$TARGET_DIR/.claude" -name "*.py" -exec chmod +x {} + 2>/dev/null || true
        # Prebuilt Go binaries (hooks, gtk, code-index server)
        find "$TARGET_DIR/.claude/hooks/binaries" -type f -exec chmod +x {} + 2>/dev/null || true
        find "$TARGET_DIR/.claude/hooks/gtk" -type f -name "gtk-*" -exec chmod +x {} + 2>/dev/null || true
        find "$TARGET_DIR/.claude/mcp-index-server" -type f -name "code-index-server-*" -exec chmod +x {} + 2>/dev/null || true
    fi

    # Install .mcp.json
    if [ -f "$SCRIPT_DIR/.mcp.json" ]; then
        print_header "Installing .mcp.json"
        install_file "$SCRIPT_DIR/.mcp.json" "$TARGET_DIR/.mcp.json" ".mcp.json"
    fi

    # Runtime directories (gitignored, populated at runtime)
    print_header "Creating runtime directories"
    local dirs=(".claude/index" ".claude/index-cache" ".claude/logs" ".claude/memories")
    for dir in "${dirs[@]}"; do
        if [ "$DRY_RUN" != true ]; then
            mkdir -p "$TARGET_DIR/$dir"
        fi
        print_message "$GREEN" "  ok $dir/"
    done

    # Stamp version
    if [ "$DRY_RUN" != true ]; then
        echo "$VERSION" > "$TARGET_DIR/.claude/VERSION"
    fi

    # Local-only git exclude (avoids committing .claude/ runtime + .mcp.json by default)
    print_header "Git configuration"
    update_git_exclude

    print_header "Installation complete"
    if [ "$DRY_RUN" = true ]; then
        print_message "$YELLOW" "Dry run. No changes made."
    else
        print_message "$GREEN" "ok configuration installed"
        echo ""
        echo "Next steps:"
        echo "  - Slash commands: /research, /plan, /build, /review, /pr-review, /ship"
        echo "  - Run a pipeline: ./.claude/pipelines/pipeline.py .claude/pipelines/research-plan.yaml \"your topic\""
        echo "  - Optional env: export CONTEXT7_API_KEY=... in your shell profile to enable Context7"
    fi
}

main
