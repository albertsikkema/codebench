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
    local marker="# Added by codebench install-helper.sh"

    if [ "$DRY_RUN" = true ]; then
        print_message "$YELLOW" "  [DRY RUN] would update .gitignore"
        return 0
    fi

    [ ! -f "$gi" ] && touch "$gi"

    # Collect entries that aren't already present.
    local to_add=()
    for entry in "${entries[@]}"; do
        if ! grep -qxF "$entry" "$gi" 2>/dev/null; then
            to_add+=("$entry")
        fi
    done

    if [ ${#to_add[@]} -eq 0 ]; then
        return 0
    fi

    # Prepend a marker comment block on the first run only.
    if ! grep -qxF "$marker" "$gi" 2>/dev/null; then
        # Add a leading blank line if file is non-empty and doesn't end in one.
        if [ -s "$gi" ] && [ "$(tail -c1 "$gi")" != "" ]; then
            echo "" >> "$gi"
        fi
        echo "$marker" >> "$gi"
    fi

    for entry in "${to_add[@]}"; do
        echo "$entry" >> "$gi"
        print_message "$GREEN" "  ok added to .gitignore: $entry"
    done
}

update_env() {
    local envf="$TARGET_DIR/.env"
    local marker="# Added by codebench install-helper.sh"
    # Parallel arrays: variable name + the commented placeholder line to append.
    local vars=("GH_TOKEN" "CONTEXT7_API_KEY")
    local lines=(
        "# GH_TOKEN=          # written by: uv run .claude/helpers/setup-github-token.py"
        "# CONTEXT7_API_KEY=  # optional, raises Context7 rate limits (works without)"
    )

    if [ "$DRY_RUN" = true ]; then
        print_message "$YELLOW" "  [DRY RUN] would update .env"
        return 0
    fi

    [ ! -f "$envf" ] && touch "$envf"

    local to_add_idx=()
    for i in "${!vars[@]}"; do
        local var="${vars[$i]}"
        # Match VAR= or # VAR= (any leading whitespace, optional comment marker).
        if ! grep -qE "^[[:space:]]*#?[[:space:]]*${var}=" "$envf" 2>/dev/null; then
            to_add_idx+=("$i")
        fi
    done

    if [ ${#to_add_idx[@]} -eq 0 ]; then
        return 0
    fi

    if ! grep -qxF "$marker" "$envf" 2>/dev/null; then
        if [ -s "$envf" ] && [ "$(tail -c1 "$envf")" != "" ]; then
            echo "" >> "$envf"
        fi
        echo "$marker" >> "$envf"
    fi

    for i in "${to_add_idx[@]}"; do
        echo "${lines[$i]}" >> "$envf"
        print_message "$GREEN" "  ok added to .env: ${vars[$i]}"
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

    # Upgrade hint: compare existing VERSION against the one we're about to write.
    if [ -f "$TARGET_DIR/.claude/VERSION" ]; then
        local prev_version
        prev_version=$(cat "$TARGET_DIR/.claude/VERSION" 2>/dev/null | tr -d '[:space:]')
        if [ -n "$prev_version" ] && [ "$prev_version" != "$VERSION" ]; then
            print_message "$YELLOW" "Upgrading: $prev_version -> $VERSION"
        elif [ "$prev_version" = "$VERSION" ]; then
            print_message "$BLUE" "Reinstalling same version: $VERSION"
        fi
    fi

    # Pre-existing .claude/ files that will be overwritten (excludes runtime dirs).
    if [ -d "$TARGET_DIR/.claude" ]; then
        local existing
        existing=$(find "$TARGET_DIR/.claude" -type f \
            -not -path '*/index/*' \
            -not -path '*/index-cache/*' \
            -not -path '*/logs/*' \
            -not -path '*/memories/*' \
            2>/dev/null | wc -l | tr -d ' ')
        if [ "$existing" -gt 0 ]; then
            print_message "$YELLOW" "Found $existing existing files in .claude/ - shipped files will be overwritten, your additions stay."
        fi
    fi

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

    # Install READTHISFIRST.md only on first install (no existing .claude/VERSION).
    # On re-runs, skip whether the file is present (don't clobber edits) or absent
    # (user already read it and deleted it - don't bring it back).
    if [ -f "$SCRIPT_DIR/READTHISFIRST.md" ] && [ ! -f "$TARGET_DIR/.claude/VERSION" ]; then
        print_header "Installing READTHISFIRST.md"
        install_file "$SCRIPT_DIR/READTHISFIRST.md" "$TARGET_DIR/READTHISFIRST.md" "READTHISFIRST.md"
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

    update_env

    # Detect git repo state for the post-install hint.
    local git_state="ok"
    if [ ! -d "$TARGET_DIR/.git" ]; then
        git_state="not_a_repo"
    elif ! git -C "$TARGET_DIR" remote get-url origin >/dev/null 2>&1; then
        git_state="no_origin"
    fi

    # Check runtime dependencies (don't fail; just hint).
    print_header "Dependency check"
    local missing=()
    check_dep() {
        local cmd="$1" hint="$2"
        if command -v "$cmd" >/dev/null 2>&1; then
            print_message "$GREEN" "  ok $cmd"
        else
            print_message "$YELLOW" "  missing $cmd  -  $hint"
            missing+=("$cmd")
        fi
    }
    check_dep "uv"           "install: curl -fsSL https://astral.sh/uv/install.sh | sh"
    check_dep "claude-safe"  "install: curl -fsSL https://raw.githubusercontent.com/albertsikkema/claude-safe/main/install.sh | bash"
    check_dep "docker"       "required by claude-safe - install Docker Desktop or your distro's docker package"
    check_dep "gh"           "install: see https://cli.github.com/ (needed to create PRs)"
    check_dep "claude-server" "(optional, only needed for runner: server pipelines)"

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
        local n=1
        if [ -f "$TARGET_DIR/READTHISFIRST.md" ]; then
            echo "  $n. Read READTHISFIRST.md - quick map of what got installed and what to do first"
            n=$((n + 1))
        fi
        echo "  $n. Set up GitHub token (so Claude can push branches and open PRs in this repo):"
        echo "       uv run .claude/helpers/setup-github-token.py"
        n=$((n + 1))
        echo "  $n. Protect the main branch (require PRs, block direct pushes):"
        echo "       uv run .claude/helpers/setup-branch-protection.py"
        n=$((n + 1))
        echo "  $n. Try a slash command: /research, /plan, /build, /review, /pr, /pr-review, /ship, /release"
        n=$((n + 1))
        echo "  $n. Run a pipeline:"
        echo "       ./.claude/pipelines/pipeline.py .claude/pipelines/research-plan.yaml \"your topic\""
        n=$((n + 1))
        echo "  $n. Optional: export CONTEXT7_API_KEY=... in your shell profile to raise Context7 rate limits (works without)"
    fi
}

main
