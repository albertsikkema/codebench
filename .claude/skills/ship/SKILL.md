---
name: ship
description: Commit, create PR, add comment, merge, checkout default branch, and pull
user-invocable: true
---

You are tasked with shipping the current branch: commit, create a PR, merge it, and return to the default branch. Execute each step sequentially — stop immediately if any step fails.

## gtk bypass

The gtk pre-tool-use hook filters git output to save tokens. For diffs you need the full content — bypass gtk using the proxy binary directly:

```bash
GTK=".claude/gtk/gtk-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"
$GTK proxy git diff
$GTK proxy git diff --cached
$GTK proxy git diff <merge_base>..HEAD
```

Use `$GTK proxy` for ALL `git diff` and `git log` commands in this workflow. Other git commands (status, push, checkout, rebase) can run normally.

## Steps

### 1. Pre-flight

Run the preflight script to gather context:

```bash
bash .claude/skills/ship/preflight.sh
```

This gives you: `default_branch`, `current_branch`, `on_default`, `commits_ahead`, git status, and existing PR info.

### 2. Ensure feature branch

If `on_default` is `true`:
- Inspect the staged/unstaged changes (`$GTK proxy git diff`, `$GTK proxy git diff --cached`) to understand what changed
- Derive a branch name from the changes using the naming convention: `feat/`, `fix/`, `refactor/`, `chore/` with short lowercase hyphenated description. Include a task ID if one is apparent.
- Create and switch to the branch: `git checkout -b <branch-name>`
- Do NOT ask the user — just create it

If already on a feature branch, continue.

### 3. Commit (if there are uncommitted changes)

If there are staged or unstaged changes:
- Run `$GTK proxy git diff` and `$GTK proxy git diff --cached` to understand what changed
- Run `$GTK proxy git log --oneline -5` to match the repo's commit style
- Stage relevant files by name (never `git add -A` or `git add .`)
- Do NOT stage files that look like secrets (.env, credentials, tokens)
- Write a Conventional Commits message: `<type>[scope]: <description>`
- Keep subject line under 50 chars, imperative mood, no period
- Use a HEREDOC for the commit message
- No AI attribution — no Co-Authored-By, no "Generated with" footers

If there are NO changes and NO new commits ahead of the default branch: **STOP**. Nothing to ship.

### 4. Push the branch

- Rebase onto target: `git rebase origin/<default-branch>`
  - If conflicts: abort rebase, try merge, if that also conflicts: **STOP** and tell the user
- Push: `git push -u origin $(git branch --show-current)`
  - If rejected (non-fast-forward after rebase): `git push --force-with-lease -u origin $(git branch --show-current)`
  - If push fails for other reasons: **STOP** and show the error

### 5. Create the PR

- If the preflight showed an existing open PR, reuse it
- If no PR exists, create one:
  - Get commit history: `$GTK proxy git log --oneline <merge_base>..HEAD`
  - Get diff stats: `$GTK proxy git diff <merge_base>..HEAD --stat`
  - Generate a concise title (under 70 chars, Conventional Commits style)
  - Generate a body with a `## Summary` section (2-4 bullet points)
  - No AI attribution in the PR title or body
  - Create: `gh pr create --title "<title>" --body "$(cat <<'EOF' ... EOF)"`
- Display the PR URL

### 6. Add a review comment

Add a short comment summarizing the changes and any notes for reviewers:

```bash
gh pr comment <number> --body "$(cat <<'EOF'
<1-3 sentence summary of changes and anything worth noting>
EOF
)"
```

### 7. Merge the PR

- Check PR status: `gh pr checks <number>` — if checks are pending or failing, warn the user but proceed if they confirm
- Merge with squash: `gh pr merge <number> --squash --delete-branch`
  - If merge fails (branch protection, checks required, etc.): **STOP** and show the error with suggestions

### 8. Return to default branch

- Checkout: `git checkout <default-branch>`
- Pull latest: `git pull origin <default-branch>`
- Confirm success: show `git log --oneline -3` so the user can see the merged commit

### 9. Update plan status (if applicable)

- Check if the branch name contains a task ID (e.g., `T-001`)
- If so, check if there's an associated plan via `list_plans(todo_item_id=<task_id>)`
- If a plan exists in `building` or `pr_ready` status, update it to `merged` via `update_plan_status(plan_id, "merged")`

## Important notes

- Never skip steps or continue past a failure
- Ask the user before proceeding only if something looks unexpected
- No AI markers anywhere — commits, PR title, PR body, comments
- If the repo has required checks or branch protection, the merge step may fail — that's expected, inform the user
