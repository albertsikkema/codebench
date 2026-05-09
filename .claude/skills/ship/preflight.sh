#!/usr/bin/env bash
set -euo pipefail

# Ship preflight: gather all context needed for the /ship workflow.
# Outputs structured data for Claude to consume.

# --- Default branch (via gh, with git fallback) ---
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

CURRENT_BRANCH=$(git branch --show-current)

echo "=== SHIP PREFLIGHT ==="
echo "default_branch: $DEFAULT_BRANCH"
echo "current_branch: $CURRENT_BRANCH"
echo "on_default: $([ "$CURRENT_BRANCH" = "$DEFAULT_BRANCH" ] && echo true || echo false)"

# --- Commits ahead ---
git fetch origin "$DEFAULT_BRANCH" --quiet 2>/dev/null || true
MERGE_BASE=$(git merge-base HEAD "origin/$DEFAULT_BRANCH" 2>/dev/null || echo "")
if [ -n "$MERGE_BASE" ]; then
  COMMITS_AHEAD=$(git rev-list --count "$MERGE_BASE..HEAD" 2>/dev/null || echo 0)
  echo "merge_base: $MERGE_BASE"
  echo "commits_ahead: $COMMITS_AHEAD"
fi

# --- Working tree status ---
echo ""
echo "=== GIT STATUS ==="
git status --short

# --- Existing PR for this branch ---
echo ""
echo "=== PR STATUS ==="
if [ "$CURRENT_BRANCH" != "$DEFAULT_BRANCH" ]; then
  gh pr view --json number,url,title,state 2>/dev/null || echo "no_pr: true"
else
  echo "no_pr: true (on default branch)"
fi
