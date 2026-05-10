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

# Copy every file under $1 into $TARGET_DIR (preserving relative paths under SCRIPT_DIR).
# Prints a single summary line. Stores file count in $LAST_COUNT.
LAST_COUNT=0
install_tree() {
    local src_root="$1" label="$2"
    local count=0
    while IFS= read -r -d '' f; do
        local rel_path="${f#"$SCRIPT_DIR/"}"
        if [ "$DRY_RUN" != true ]; then
            mkdir -p "$(dirname "$TARGET_DIR/$rel_path")"
            cp "$f" "$TARGET_DIR/$rel_path"
        fi
        count=$((count + 1))
    done < <(find "$src_root" -type f \
        -not -path '*/.venv/*' \
        -not -path '*/index/*' \
        -not -path '*/index-cache/*' \
        -not -path '*/logs/*' \
        -not -path '*/memories/*' \
        -print0)
    LAST_COUNT=$count
    local prefix="ok"
    [ "$DRY_RUN" = true ] && prefix="[DRY RUN]"
    printf "  ${GREEN}%-10s %-26s %4d files${NC}\n" "$prefix" "$label" "$count"
}

update_gitignore() {
    local gi="$TARGET_DIR/.gitignore"
    local entries=(".claude" ".mcp.json" ".env" ".playwright/" ".playwright-mcp/")

    if [ "$DRY_RUN" = true ]; then
        print_message "$YELLOW" "  [DRY RUN] would update .gitignore"
        return 0
    fi

    [ ! -f "$gi" ] && touch "$gi"

    for entry in "${entries[@]}"; do
        if ! grep -qxF "$entry" "$gi" 2>/dev/null; then
            echo "$entry" >> "$gi"
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

    # Install .claude/ per top-level subdir, plus loose files at .claude/ root.
    # Skip runtime dirs (index/, index-cache/, logs/, memories/) — created below.
    print_header "Installing .claude/"
    local total=0
    local skip_dirs=" index index-cache logs memories "
    for sub in "$SCRIPT_DIR/.claude"/*/; do
        local name
        name=$(basename "$sub")
        case "$skip_dirs" in *" $name "*) continue ;; esac
        install_tree "$sub" ".claude/$name/"
        total=$((total + LAST_COUNT))
    done
    # Loose files in .claude/ root (e.g. settings.json)
    local root_count=0
    for f in "$SCRIPT_DIR/.claude"/*; do
        [ -f "$f" ] || continue
        local rel_path="${f#"$SCRIPT_DIR/"}"
        if [ "$DRY_RUN" != true ]; then
            mkdir -p "$(dirname "$TARGET_DIR/$rel_path")"
            cp "$f" "$TARGET_DIR/$rel_path"
        fi
        root_count=$((root_count + 1))
    done
    if [ "$root_count" -gt 0 ]; then
        local prefix="ok"
        [ "$DRY_RUN" = true ] && prefix="[DRY RUN]"
        printf "  ${GREEN}%-10s %-26s %4d files${NC}\n" "$prefix" ".claude/ (root)" "$root_count"
        total=$((total + root_count))
    fi
    printf "  ${BLUE}%-10s %-26s %4d files${NC}\n" "total" "" "$total"

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

    # .gitignore — keep .claude/, .mcp.json, .env out of commits by default
    print_header "Git configuration"
    update_gitignore

    # Detect git repo state for the post-install hint.
    local git_state="ok"
    if [ ! -d "$TARGET_DIR/.git" ]; then
        git_state="not_a_repo"
    elif ! git -C "$TARGET_DIR" remote get-url origin >/dev/null 2>&1; then
        git_state="no_origin"
    fi

    print_header "Installation complete"
    if [ "$DRY_RUN" = true ]; then
        print_message "$YELLOW" "Dry run. No changes made."
    else
        print_message "$GREEN" "ok configuration installed"

        if [ "$git_state" != "ok" ]; then
            echo ""
            print_message "$YELLOW" "Heads up: this directory is not a git repo with an 'origin' remote."
            print_message "$YELLOW" "Claude needs one to push branches and open PRs. To set it up:"
            echo ""
            if [ "$git_state" = "not_a_repo" ]; then
                echo "    git init"
                echo "    git add ."
                echo "    git commit -m 'initial commit'"
            fi
            echo "    gh repo create <name> --source=. --private --remote=origin --push"
            echo "    # or, if the remote already exists on GitHub:"
            echo "    git remote add origin git@github.com:<owner>/<name>.git"
            echo "    git push -u origin main"
        fi

        echo ""
        echo "Next steps:"
        echo "  - Set up GitHub token (so Claude can push branches and open PRs in this repo):"
        echo "      uv run .claude/helpers/setup-github-token.py"
        echo "  - Slash commands: /research, /plan, /build, /review, /pr-review, /ship"
        echo "  - Run a pipeline: ./.claude/pipelines/pipeline.py .claude/pipelines/research-plan.yaml \"your topic\""
        echo "  - Optional env: export CONTEXT7_API_KEY=... in your shell profile to enable Context7"
    fi
}

main
