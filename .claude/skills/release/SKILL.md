---
name: release
description: Create a production release with changelog, version bump, PR, tag. Supports two-branch (dev->main) and single-branch (release branch->main) workflows.
user-invocable: true
---

You are tasked with creating a production release: changelog generation, version bump, PR, merge (with confirmation), and tag. Execute each step sequentially -- stop immediately if any step fails.

## Arguments

Optional: version bump type -- `patch` (default), `minor`, or `major`.

## gtk proxy

Use the gtk proxy for all `git diff` and `git log` commands (see gtk rule for details):

```bash
GTK=".claude/hooks/gtk/gtk-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"
```

## Steps

### 1. Preflight

Run the preflight script to gather all context:

```bash
bash .claude/skills/release/preflight.sh <bump_type>
```

This gives you: `workflow` (single-branch/two-branch), `prod_branch`, `dev_branch`, `current_version`, `next_version`, `commit_count`, commit log, changelog files found, version files found, and whether a CI release workflow exists.

### 2. Validate preflight results

Check the preflight output and **STOP** if any of these are true:
- `dirty: true` -- tell the user to commit or stash changes first
- `commit_count: 0` -- nothing to release
- Bump type is invalid

### 3. Present release summary -- GATE

Show the user:
- Current version -> next version (bump type)
- Number of commits
- Grouped summary of changes (feat, fix, refactor, etc.)
- Workflow type detected

**Ask for confirmation before proceeding.** If the user says no, stop.

### 4. Checkout the source branch

**Two-branch:**
```bash
git checkout <dev-branch>
git pull origin <dev-branch>
```

**Single-branch:**
```bash
git checkout <prod-branch>
git pull origin <prod-branch>
```

### 5. Update changelogs

Look for existing changelog files (from preflight output). Read them first to understand the format and conventions used.

**If a changelog exists:** match its existing format, style, language, and level of detail.

**If no changelog exists:** create `CHANGELOG.md` in Keep a Changelog 1.1.0 format.

#### Changelog format (Keep a Changelog 1.1.0)

Group changes by parsing Conventional Commit prefixes:

| Commit prefix | Changelog section |
|---------------|-------------------|
| `feat:` / `feat(scope):` | **Added** |
| `fix:` / `fix(scope):` | **Fixed** |
| `refactor:`, `docs:`, `perf:`, `style:`, `test:` | **Changed** |
| `BREAKING CHANGE:` in commit body | **Breaking Changes** (at the top) |
| `chore:`, `ci:`, `build:` | Skipped -- not user-facing |

Rules:
- Exclude merge commits
- Strip the type prefix from each entry (e.g., `feat(auth): Add login` becomes `Add login`)
- Capitalize the first letter of each entry
- Add compare/diff links at the bottom pointing to GitHub compare URLs
- Omit empty sections

Example:
```markdown
## [1.3.0] - 2026-04-07

### Added

- User authentication via OAuth2
- Rate limiting on API endpoints

### Fixed

- Race condition in order processing

[1.3.0]: https://github.com/owner/repo/compare/v1.2.0...v1.3.0
```

### 6. Update version files

From the preflight output, update any version files that were detected:
- `package.json` / `package-lock.json` -> `"version": "X.Y.Z"`
- `pyproject.toml` -> `version = "X.Y.Z"`
- `Cargo.toml` -> `version = "X.Y.Z"`

Only update files that already contain a version field. Do not create version files. Tags remain the source of truth.

### 7. Commit changes

**Two-branch:** commit on the dev branch:
```bash
git add <changelog-files> <version-files>
git commit -m "chore: Update changelog for vX.Y.Z"
```

**Single-branch:** create a release branch, commit there:
```bash
git checkout -b release/vX.Y.Z
git add <changelog-files> <version-files>
git commit -m "chore: Update changelog for vX.Y.Z"
```

### 8. Push and create PR

**Two-branch:**
```bash
git push origin <dev-branch>
gh pr create --base <prod-branch> --head <dev-branch> --title "Release vX.Y.Z" --body "$(cat <<'EOF'
## Release vX.Y.Z

### Changes
<grouped bullet points from the changelog -- Added, Fixed, Changed, Breaking>

### Version bump
`vPREV` -> `vX.Y.Z` (<bump-type>)
EOF
)"
```

**Single-branch:**
```bash
git push -u origin release/vX.Y.Z
gh pr create --base <prod-branch> --head release/vX.Y.Z --title "Release vX.Y.Z" --body "$(cat <<'EOF'
## Release vX.Y.Z

### Changes
<grouped bullet points from the changelog -- Added, Fixed, Changed, Breaking>

### Version bump
`vPREV` -> `vX.Y.Z` (<bump-type>)
EOF
)"
```

If push fails due to branch protection: **STOP** and inform the user.

### 9. Ask before merging -- GATE

**STOP** and show the user:
- The PR URL
- The change summary
- Ask: "Ready to merge and tag vX.Y.Z?"

Do NOT proceed until the user explicitly confirms. If the user says no, stop here -- the PR stays open for review.

### 10. Merge the PR

**Two-branch -- CRITICAL: never delete the development branch:**
```bash
gh pr merge <number> --merge
```

**Single-branch:**
```bash
gh pr merge <number> --squash --delete-branch
```

Use `--merge` for two-branch (preserves commit history), `--squash` for single-branch (clean history).

If merge fails (branch protection, CI checks, merge conflicts): **STOP** and inform the user with the specific error.

### 11. Tag the release

```bash
git fetch origin <prod-branch>
git tag vX.Y.Z origin/<prod-branch>
git push origin vX.Y.Z
```

If tagging or pushing the tag fails after merge, inform the user with recovery steps:
```
The PR was merged but tagging failed. To recover manually:
  git fetch origin <prod-branch>
  git tag vX.Y.Z origin/<prod-branch>
  git push origin vX.Y.Z
```

### 12. Sync local branches

**Two-branch:**
```bash
git checkout <dev-branch>
git pull origin <dev-branch>
```

**Single-branch:**
```bash
git checkout <prod-branch>
git pull origin <prod-branch>
```

### 13. Confirm

Show the user:
- The PR URL (merged)
- The new tag: `vX.Y.Z`
- `$GTK proxy git log --oneline -5 origin/<prod-branch>` to confirm the state
- If `release_workflow: true` from preflight, mention that GitHub Actions should trigger automatically from the tag push

## Important notes

- Never skip steps or continue past a failure
- Two confirmation gates: step 3 (before any changes) and step 9 (before merge)
- No AI markers anywhere -- commits, PR title, PR body
- If the repo has required checks or branch protection, the merge step may fail -- that's expected, inform the user
