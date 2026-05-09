#!/usr/bin/env bash
# Changelog & release helper.
# Usage:
#   changelog-release.sh changelog   — regenerate CHANGELOG.md, commit via PR
#   changelog-release.sh release     — interactive version bump, changelog, PR, tag, push
set -euo pipefail

# ── Helpers ──────────────────────────────────────────────────────

require_default_branch() {
  local default current
  default=$(gh repo view --json defaultBranchRef --jq '.defaultBranchRef.name' 2>/dev/null || echo "main")
  current=$(git branch --show-current)
  if [ "$current" != "$default" ]; then
    echo "Error: must be on $default branch (currently on $current)."
    exit 1
  fi
  echo "$default"
}

commit_via_pr() {
  local title="$1" body="$2" branch="$3" default="$4"
  git push origin "$branch"
  echo "Creating PR..."
  gh pr create --title "$title" --body "$body" --base "$default" --head "$branch"
  echo "Merging PR..."
  gh pr merge "$branch" --squash --delete-branch
  git checkout "$default"
  git pull origin "$default"
}

# ── Changelog generation ─────────────────────────────────────────
# Generates CHANGELOG.md in Keep a Changelog 1.1.0 format from Conventional Commits.
# Arg: optional version string (e.g. "v1.2.0"). Without it, unreleased commits
# appear under [Unreleased]. With it, they appear under [1.2.0] - date.

generate_changelog() {
  local new_version="${1:-}"
  local output="CHANGELOG.md"
  local repo_url
  repo_url=$(git remote get-url origin 2>/dev/null | sed -E 's/\.git$//' | sed -E 's|git@github\.com:|https://github.com/|')

  # Write grouped commits for a git range into a temp file.
  # Returns 0 if any commits were written, 1 if empty.
  _write_grouped_commits() {
    local range="$1"
    local dest="$2"
    local has_content=1

    # Breaking changes
    local breaking
    breaking=$(git --no-pager log --no-merges --pretty=format:"%s%n%b" $range 2>/dev/null \
      | { grep -E "^BREAKING CHANGE:" || true; } \
      | sed 's/^BREAKING CHANGE: //' \
      | awk '{print toupper(substr($0,1,1)) substr($0,2)}' \
      | while IFS= read -r line; do [ -n "$line" ] && echo "- $line"; done) || true

    if [ -n "$breaking" ]; then
      printf "### Breaking Changes\n\n%s\n\n" "$breaking" >> "$dest"
      has_content=0
    fi

    # Added (feat)
    local added
    added=$(git --no-pager log --no-merges --grep="^feat(" --grep="^feat:" --pretty=format:"%s" $range 2>/dev/null \
      | sed -E 's/^feat(\([^)]*\))?: //' | awk '{print toupper(substr($0,1,1)) substr($0,2)}' \
      | while IFS= read -r line; do [ -n "$line" ] && echo "- $line"; done) || true

    if [ -n "$added" ]; then
      printf "### Added\n\n%s\n\n" "$added" >> "$dest"
      has_content=0
    fi

    # Fixed (fix)
    local fixed
    fixed=$(git --no-pager log --no-merges --grep="^fix(" --grep="^fix:" --pretty=format:"%s" $range 2>/dev/null \
      | sed -E 's/^fix(\([^)]*\))?: //' | awk '{print toupper(substr($0,1,1)) substr($0,2)}' \
      | while IFS= read -r line; do [ -n "$line" ] && echo "- $line"; done) || true

    if [ -n "$fixed" ]; then
      printf "### Fixed\n\n%s\n\n" "$fixed" >> "$dest"
      has_content=0
    fi

    # Changed (refactor, docs, perf, style, test)
    local changed=""
    for type in refactor docs perf style test; do
      local commits
      commits=$(git --no-pager log --no-merges --grep="^${type}(" --grep="^${type}:" --pretty=format:"%s" $range 2>/dev/null \
        | sed -E "s/^${type}(\([^)]*\))?: //" | awk '{print toupper(substr($0,1,1)) substr($0,2)}' \
        | while IFS= read -r line; do [ -n "$line" ] && echo "- $line"; done) || true
      if [ -n "$commits" ]; then
        changed="${changed}${commits}"$'\n'
      fi
    done
    changed=$(echo "$changed" | sed '/^$/d')

    if [ -n "$changed" ]; then
      printf "### Changed\n\n%s\n\n" "$changed" >> "$dest"
      has_content=0
    fi

    return $has_content
  }

  # Collect all tags
  local all_tags=()
  for tag in $(git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$'); do
    all_tags+=("$tag")
  done

  # Start file
  printf "# Changelog\n\n" > "$output"

  # Unreleased / new version section
  local latest range tmp
  latest=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
  range="HEAD"
  [ -n "$latest" ] && range="${latest}..HEAD"
  tmp=$(mktemp)

  if _write_grouped_commits "$range" "$tmp"; then
    if [ -n "$new_version" ]; then
      local ver="${new_version#v}"
      local date
      date=$(date +%Y-%m-%d)
      printf "## [%s] - %s\n\n" "$ver" "$date" >> "$output"
    else
      printf "## [Unreleased]\n\n" >> "$output"
    fi
    cat "$tmp" >> "$output"
  fi
  rm -f "$tmp"

  # Existing tagged versions
  for tag in "${all_tags[@]}"; do
    local ver="${tag#v}"
    local prev tdate
    prev=$(git describe --tags --abbrev=0 "${tag}^" 2>/dev/null || echo "")
    tdate=$(git log -1 --format=%cs "$tag")
    tmp=$(mktemp)

    range="$tag"
    [ -n "$prev" ] && range="${prev}..${tag}"

    if _write_grouped_commits "$range" "$tmp"; then
      printf "## [%s] - %s\n\n" "$ver" "$tdate" >> "$output"
      cat "$tmp" >> "$output"
    fi
    rm -f "$tmp"
  done

  # Diff links at bottom
  if [ -n "$repo_url" ]; then
    echo "" >> "$output"

    if [ -n "$new_version" ]; then
      local ver="${new_version#v}"
      latest=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
      [ -n "$latest" ] && echo "[$ver]: $repo_url/compare/$latest...$new_version" >> "$output"
    fi

    for i in "${!all_tags[@]}"; do
      tag="${all_tags[$i]}"
      local ver="${tag#v}"
      local next_idx=$((i + 1))
      if [ "$next_idx" -lt "${#all_tags[@]}" ]; then
        local prev_tag="${all_tags[$next_idx]}"
        echo "[$ver]: $repo_url/compare/$prev_tag...$tag" >> "$output"
      else
        echo "[$ver]: $repo_url/releases/tag/$tag" >> "$output"
      fi
    done
  fi

  # Footer
  cat >> "$output" <<'FOOTER'

---

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
FOOTER
}

# ── Commands ─────────────────────────────────────────────────────

cmd_changelog() {
  local default branch
  default=$(require_default_branch)
  branch="chore/update-changelog"

  echo "Generating changelog..."
  git checkout -b "$branch"
  generate_changelog
  git add CHANGELOG.md

  if git diff --cached --quiet; then
    echo "No changelog changes."
    git checkout "$default"
    git branch -D "$branch"
  else
    git commit -m "chore: Update changelog"
    commit_via_pr "chore: Update changelog" "Automated changelog update." "$branch" "$default"
    echo "Changelog updated."
  fi
}

cmd_release() {
  local default latest major minor patch new branch
  default=$(require_default_branch)

  latest=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
  if [ -z "$latest" ]; then
    echo "No existing tags found. Starting from v0.0.0."
    major=0; minor=0; patch=0
  else
    local ver="${latest#v}"
    major="${ver%%.*}"
    local rest="${ver#*.}"
    minor="${rest%%.*}"
    patch="${rest#*.}"
  fi

  echo "Current version: v${major}.${minor}.${patch}"
  echo ""
  echo "  1) patch  → v${major}.${minor}.$((patch + 1))"
  echo "  2) minor  → v${major}.$((minor + 1)).0"
  echo "  3) major  → v$((major + 1)).0.0"
  echo ""
  printf "Select bump type [1/2/3]: "
  read -r choice
  case $choice in
    1) new="v${major}.${minor}.$((patch + 1))";;
    2) new="v${major}.$((minor + 1)).0";;
    3) new="v$((major + 1)).0.0";;
    *) echo "Invalid choice"; exit 1;;
  esac

  echo ""
  echo "Will create tag: $new"
  printf "Confirm? [y/N]: "
  read -r confirm
  case $confirm in
    [yY]) ;;
    *) echo "Aborted."; exit 1;;
  esac

  branch="chore/changelog-${new}"
  echo ""
  echo "Generating changelog..."
  git checkout -b "$branch"
  generate_changelog "$new"
  git add CHANGELOG.md
  git commit -m "chore: Update changelog for ${new}"
  commit_via_pr "chore: Update changelog for ${new}" "Automated changelog update for release ${new}." "$branch" "$default"

  echo "Tagging ${new} on ${default}..."
  git tag -a "$new" -m "Release ${new}"
  git push origin "$new"
  echo ""
  echo "Released ${new}. GitHub Actions will create the GitHub Release."
}

# ── Dispatch ─────────────────────────────────────────────────────

case "${1:-}" in
  changelog) cmd_changelog;;
  release)   cmd_release;;
  *)         echo "Usage: $0 {changelog|release}"; exit 1;;
esac
