---
name: pr
description: Generate PR description, sync branch, push, and create/update PR
user-invocable: true
---

You are tasked with generating a comprehensive pull request description and creating or updating a PR.

## Steps

### 1. Identify the PR context

- Get current branch: `git branch --show-current`
- If on the default branch, list open PRs: `gh pr list --limit 10` and ask the user which one
- Check if a PR already exists for this branch: `gh pr view --json url,number,title,state 2>/dev/null`
- Detect the default branch: `git symbolic-ref refs/remotes/origin/HEAD | sed 's|refs/remotes/origin/||'` (fallback: check for `main` then `master`)
- Get the base branch: `gh pr view --json baseRefName 2>/dev/null` or use the detected default branch

### 2. Gather change information

- Get commit history since branching: `git log --oneline $(git merge-base HEAD <default-branch>)..HEAD`
- Get full diff: `git diff $(git merge-base HEAD <default-branch>)..HEAD`
- Get diff stats: `git diff $(git merge-base HEAD <default-branch>)..HEAD --stat`
- Read any related plan files in `.claude/workspace/plans/` if commits reference them
- Read any related research files in `.claude/workspace/research/` for context

### 3. Analyze the changes thoroughly

Think deeply about the code changes:
- Read through the entire diff carefully
- For context, read any files that are referenced but not fully shown in the diff
- Understand the purpose and impact of each change
- Identify user-facing changes vs internal implementation details
- Look for breaking changes or migration requirements

### 4. Run verification — GATE

- Detect test and lint commands (check Makefile, pyproject.toml, package.json)
- Run the linter
- Run the test suite

**If either fails: STOP IMMEDIATELY.** Do not continue to step 5. Display a clear error:

```
ERROR: Code is not PR-ready.

[Linting/Tests] failed — fix these issues before creating a PR.
Run `/build` or fix manually, then retry `/pr`.
```

Do NOT generate a PR description, push, or create a PR when verification fails.
Do NOT attempt to fix the code yourself — your role is to detect and signal problems, not to fix them.

### 5. Generate the PR description

Use this format:

```markdown
## Summary
<!-- 1-3 bullet points describing what this PR does and why -->

## Changes
<!-- Organized list of what changed, grouped by component/area -->

## Test results
<!-- Output of test/lint runs, or note if manual testing needed -->

## Breaking changes
<!-- List any breaking changes, or "None" -->

## Notes for reviewers
<!-- Anything reviewers should pay special attention to -->
```

### 6. Present to user

- Display the complete PR description
- If a PR already exists: ask whether to update it or cancel
- If no PR exists: ask whether to create it or cancel

### 7. Sync with target branch and push

**Sync to be merge-ready:**
- Fetch latest from remote: `git fetch origin`
- Determine the target branch: use the base branch identified in step 1
- Rebase onto the target: `git rebase origin/<base-branch>`
- If rebase conflicts occur:
  - Abort the rebase: `git rebase --abort`
  - Try merge instead: `git merge origin/<base-branch>`
  - If merge also conflicts, inform the user and show which files conflict — they need to resolve manually before proceeding
  - Stop here if unresolved conflicts remain

**Push the branch:**
- Push with upstream tracking: `git push -u origin $(git branch --show-current)`
- If the rebase changed history, force-push may be needed: `git push --force-with-lease -u origin $(git branch --show-current)`
- If the push fails (permission denied, hook blocked, container restrictions):
  - Inform the user: "Push failed — you may need to push from outside the container:"
  - Show the exact commands: `git push -u origin <branch-name>` (or with `--force-with-lease` if rebased)
  - Continue to step 8 anyway — the PR can be created once the push succeeds

### 8. Create or update the PR (upon user confirmation)

**Creating a new PR:**
```bash
gh pr create --title "<title>" --body "$(cat <<'EOF'
<generated description>
EOF
)"
```

**Updating an existing PR:**
- Write description to a temp file
- `gh pr edit {number} --body-file <temp_file>`
- Clean up temp file

### 9. Final output

- Show the PR URL
- If any verification steps failed or need manual testing, remind the user
- If the push failed earlier, remind the user to push before the PR is visible

## Important notes

- Be thorough but concise — descriptions should be scannable
- Focus on the "why" as much as the "what"
- Include breaking changes prominently
- If the PR touches multiple components, organize the description accordingly
- Keep the title under 70 characters
- Use Conventional Commits style for the title if the repo follows that convention
