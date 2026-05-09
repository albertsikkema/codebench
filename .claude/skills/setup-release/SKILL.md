---
name: setup-release
description: Set up changelog generation and release workflow for any project. Creates a shell script for changelog/release orchestration, GitHub Actions release workflow, and Makefile targets. Supports protected main branches via PR-based flow.
disable-model-invocation: true
---

You are setting up a release and changelog mechanism for a project.

## What This Creates

1. **`scripts/changelog-release.sh`** — Single script with two subcommands:
   - `changelog` — regenerates `CHANGELOG.md` in Keep a Changelog 1.1.0 format, commits via PR
   - `release` — interactive semver bump, changelog generation, PR, merge, tag, push
2. **`.github/workflows/release.yml`** — Triggered by `v*.*.*` tags, creates GitHub Release with auto-generated notes
3. **Makefile targets** — `make changelog` and `make release` as thin wrappers

## Step 1: Understand the Project

Before creating anything, gather context:

- **Language/framework**: Read build files (go.mod, package.json, Cargo.toml, pyproject.toml, etc.)
- **Existing CI**: Check `.github/workflows/` for existing pipelines
- **Existing Makefile**: Check if a Makefile already exists
- **Current tags**: Run `git tag --list 'v*'` to see existing version tags
- **Branch protection**: Ask if main/default branch is protected (determines PR-based flow)

## Step 2: Copy the Helper Script

Copy `run.sh` from this skill directory to `scripts/changelog-release.sh` in the project:

```bash
mkdir -p scripts
cp .claude/skills/setup-release/run.sh scripts/changelog-release.sh
chmod +x scripts/changelog-release.sh
```

The script handles:
- **Default branch guard** — refuses to run if not on the default branch (detected via `gh repo view`)
- **Changelog generation** — Keep a Changelog 1.1.0 format with grouped sections (Added, Fixed, Changed, Breaking Changes) parsed from Conventional Commits
- **PR-based commits** — creates a branch, commits, pushes, creates PR via `gh`, merges via `gh pr merge --squash`, returns to default branch
- **Release tagging** — after changelog PR is merged, tags the commit and pushes the tag

### Changelog format

The generated `CHANGELOG.md` follows Keep a Changelog 1.1.0:
- `## [Unreleased]` section for uncommitted changes (changelog mode)
- `## [X.Y.Z] - YYYY-MM-DD` sections for each version
- `### Added` (feat), `### Fixed` (fix), `### Changed` (refactor/docs/perf/style/test), `### Breaking Changes`
- Empty versions are omitted
- Diff links at the bottom pointing to GitHub compare URLs
- Footer referencing Keep a Changelog and SemVer specs

### Conventional commit type mapping

| Commit type | Changelog section |
|-------------|-------------------|
| `feat:` | Added |
| `fix:` | Fixed |
| `refactor:`, `docs:`, `perf:`, `style:`, `test:` | Changed |
| `BREAKING CHANGE:` footer | Breaking Changes |
| `chore:`, `ci:`, `build:` | Skipped |

Merge commits are excluded. Commit messages are stripped of type prefixes and capitalized.

## Step 3: Create the GitHub Actions Release Workflow

Create `.github/workflows/release.yml`. The minimal workflow just creates a GitHub Release:

```yaml
name: Release

on:
  push:
    tags:
      - "v*.*.*"

permissions:
  contents: write

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
        with:
          fetch-depth: 0

      - name: Validate tag format
        run: |
          if ! echo "${GITHUB_REF_NAME}" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
            echo "::error::Invalid tag format: ${GITHUB_REF_NAME}"
            exit 1
          fi

      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true
```

For language-specific build steps (Go binaries, Docker images, npm publish, PyPI upload), see [workflows.md](workflows.md).

Pin GitHub Actions to commit SHAs in production repos for security.

## Step 4: Add Makefile Targets

Add to the Makefile (create one if it doesn't exist):

```makefile
.PHONY: changelog release

# ── Changelog & Release ──────────────────────────────────────────

changelog:
	scripts/changelog-release.sh changelog

release:
	scripts/changelog-release.sh release
```

## Step 5: Prerequisites

The script requires:
- `gh` CLI (GitHub CLI) — authenticated and with repo access
- Conventional Commits enforced (commitlint + git hooks recommended)
- Branch protection on main is supported — the script works via PRs

If the repo doesn't have `gh` set up or conventional commits, note this to the user.

## Step 6: Verify Setup

1. Run `scripts/changelog-release.sh` — should print usage
2. Run `make changelog` from the default branch — should generate CHANGELOG.md and create a PR
3. Ensure `.github/workflows/release.yml` is valid YAML
4. Check that the Makefile `.PHONY` includes `changelog` and `release`

## Release Flow Summary

### `make changelog` (update changelog only)
```
default branch check → create chore/update-changelog branch → generate CHANGELOG.md
→ commit → push → create PR → merge PR → return to default branch
```

### `make release` (full release)
```
default branch check → pick version (patch/minor/major) → confirm
→ create chore/changelog-vX.Y.Z branch → generate CHANGELOG.md with new version
→ commit → push → create PR → merge PR → return to default branch
→ tag vX.Y.Z → push tag → GitHub Actions creates GitHub Release
```

## Important Notes

- Tags are the source of truth for versions. No version files to keep in sync.
- The release target is interactive — requires user input for safety.
- Both commands refuse to run unless on the default branch.
- `softprops/action-gh-release` with `generate_release_notes: true` creates release notes on GitHub separately from the committed CHANGELOG.md.
- Do NOT install changelog generation tools (git-cliff, conventional-changelog, etc.) — the script handles everything with `git log`.
