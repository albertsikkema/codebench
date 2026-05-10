#!/usr/bin/env bash
set -euo pipefail

# Release preflight: gather all context needed for the /release workflow.
# Outputs structured data for Claude to consume.
#
# Usage: bash .claude/skills/release/preflight.sh [patch|minor|major]

BUMP_TYPE="${1:-patch}"

# ── gtk proxy ───────────────────────────────────────────────────
GTK=".claude/hooks/gtk/gtk-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"

# ── Default branch ──────────────────────────────────────────────
DEFAULT_BRANCH=""
if command -v gh >/dev/null 2>&1; then
  DEFAULT_BRANCH=$(gh repo view --json defaultBranchRef --jq '.defaultBranchRef.name' 2>/dev/null || true)
fi
if [ -z "$DEFAULT_BRANCH" ]; then
  DEFAULT_BRANCH=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's|refs/remotes/origin/||' || true)
fi
if [ -z "$DEFAULT_BRANCH" ]; then
  for candidate in main master; do
    if git show-ref --verify --quiet "refs/remotes/origin/$candidate" 2>/dev/null; then
      DEFAULT_BRANCH="$candidate"
      break
    fi
  done
fi
if [ -z "$DEFAULT_BRANCH" ]; then
  echo "ERROR: Could not determine default branch" >&2
  exit 1
fi

# ── Workflow detection ──────────────────────────────────────────
# Two-branch: if there's a dev/develop/development branch that isn't the default
WORKFLOW="single-branch"
DEV_BRANCH=""
for candidate in dev develop development; do
  if git show-ref --verify --quiet "refs/remotes/origin/$candidate" 2>/dev/null; then
    if [ "$candidate" != "$DEFAULT_BRANCH" ]; then
      WORKFLOW="two-branch"
      DEV_BRANCH="$candidate"
      break
    fi
  fi
done

PROD_BRANCH="$DEFAULT_BRANCH"
CURRENT_BRANCH=$(git branch --show-current)

echo "=== RELEASE PREFLIGHT ==="
echo "workflow: $WORKFLOW"
echo "prod_branch: $PROD_BRANCH"
echo "dev_branch: $DEV_BRANCH"
echo "current_branch: $CURRENT_BRANCH"
echo "bump_type: $BUMP_TYPE"

# ── Working tree status ─────────────────────────────────────────
echo ""
echo "=== WORKING TREE ==="
STATUS=$(git status --porcelain)
if [ -n "$STATUS" ]; then
  echo "dirty: true"
  echo "$STATUS"
else
  echo "dirty: false"
fi

# ── Current version ─────────────────────────────────────────────
echo ""
echo "=== VERSION ==="
CURRENT_TAG=$(git tag -l 'v*' --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1 || true)

if [ -z "$CURRENT_TAG" ]; then
  echo "current_tag: (none)"
  MAJOR=0; MINOR=0; PATCH=0
else
  echo "current_tag: $CURRENT_TAG"
  VER="${CURRENT_TAG#v}"
  MAJOR="${VER%%.*}"
  REST="${VER#*.}"
  MINOR="${REST%%.*}"
  PATCH="${REST#*.}"
fi

echo "current_version: v${MAJOR}.${MINOR}.${PATCH}"

case "$BUMP_TYPE" in
  major) NEXT="v$((MAJOR + 1)).0.0" ;;
  minor) NEXT="v${MAJOR}.$((MINOR + 1)).0" ;;
  patch) NEXT="v${MAJOR}.${MINOR}.$((PATCH + 1))" ;;
  *)     echo "ERROR: Invalid bump type: $BUMP_TYPE (use patch, minor, or major)" >&2; exit 1 ;;
esac
echo "next_version: $NEXT"

# ── Commits to release ──────────────────────────────────────────
echo ""
echo "=== COMMITS ==="

git fetch origin "$PROD_BRANCH" --quiet 2>/dev/null || true

if [ "$WORKFLOW" = "two-branch" ]; then
  git fetch origin "$DEV_BRANCH" --quiet 2>/dev/null || true
  MERGE_BASE=$(git merge-base "origin/$PROD_BRANCH" "origin/$DEV_BRANCH" 2>/dev/null || echo "")
  if [ -n "$MERGE_BASE" ]; then
    COMMIT_COUNT=$(git rev-list --count "$MERGE_BASE..origin/$DEV_BRANCH" 2>/dev/null || echo 0)
    echo "commit_count: $COMMIT_COUNT"
    echo "range: ${MERGE_BASE:0:8}..origin/$DEV_BRANCH"
    echo ""
    echo "--- log ---"
    $GTK proxy git log --oneline "$MERGE_BASE..origin/$DEV_BRANCH" 2>/dev/null || true
  else
    echo "commit_count: 0"
  fi
else
  if [ -n "$CURRENT_TAG" ]; then
    COMMIT_COUNT=$(git rev-list --count "$CURRENT_TAG..origin/$PROD_BRANCH" 2>/dev/null || echo 0)
    echo "commit_count: $COMMIT_COUNT"
    echo "range: $CURRENT_TAG..origin/$PROD_BRANCH"
    echo ""
    echo "--- log ---"
    $GTK proxy git log --oneline "$CURRENT_TAG..origin/$PROD_BRANCH" 2>/dev/null || true
  else
    COMMIT_COUNT=$(git rev-list --count "origin/$PROD_BRANCH" 2>/dev/null || echo 0)
    echo "commit_count: $COMMIT_COUNT"
    echo "range: (all commits)"
    echo ""
    echo "--- log ---"
    $GTK proxy git log --oneline "origin/$PROD_BRANCH" 2>/dev/null || true
  fi
fi

# ── Changelog files ─────────────────────────────────────────────
echo ""
echo "=== CHANGELOG FILES ==="
FOUND_CHANGELOG=false
for f in CHANGELOG.md changelog.md public/changelog.md docs/changelog.md; do
  if [ -f "$f" ]; then
    echo "found: $f"
    FOUND_CHANGELOG=true
  fi
done
if [ "$FOUND_CHANGELOG" = false ]; then
  echo "found: (none -- will create CHANGELOG.md)"
fi

# ── Version files ───────────────────────────────────────────────
echo ""
echo "=== VERSION FILES ==="
for f in package.json pyproject.toml Cargo.toml; do
  if [ -f "$f" ]; then
    case "$f" in
      package.json)
        VER_IN_FILE=$(grep -oP '"version"\s*:\s*"\K[^"]+' "$f" 2>/dev/null || true)
        [ -n "$VER_IN_FILE" ] && echo "found: $f (version: $VER_IN_FILE)"
        ;;
      pyproject.toml)
        VER_IN_FILE=$(grep -oP '^version\s*=\s*"\K[^"]+' "$f" 2>/dev/null || true)
        [ -n "$VER_IN_FILE" ] && echo "found: $f (version: $VER_IN_FILE)"
        ;;
      Cargo.toml)
        VER_IN_FILE=$(grep -oP '^version\s*=\s*"\K[^"]+' "$f" 2>/dev/null || true)
        [ -n "$VER_IN_FILE" ] && echo "found: $f (version: $VER_IN_FILE)"
        ;;
    esac
  fi
done

# ── GitHub Actions release workflow ─────────────────────────────
echo ""
echo "=== CI ==="
if [ -f ".github/workflows/release.yml" ] || [ -f ".github/workflows/release.yaml" ]; then
  echo "release_workflow: true"
else
  echo "release_workflow: false"
fi
