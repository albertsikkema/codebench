---
name: cleanup
description: Post-implementation cleanup — rationalize docs, capture learnings, update project state
user-invocable: true
---

# Cleanup Implementation

You are tasked with cleaning up after a completed implementation by analyzing what actually happened and creating documentation for significant decisions.

This implements the "Review & Rationalization Phase" from Parnas and Clements (1986): documentation should show the cleaned-up, rationalized version of what happened, not the messy discovery process.

**What cleanup does:**
- Captures what was tried (rejected alternatives in decisions.md)
- Captures implementation learnings into decisions.md
- Marks completed items in todo.md
- Updates CLAUDE.md with new patterns/conventions
- Updates documentation to reflect the final state of the codebase

## Initial Response

When this command is invoked:

1. **If file paths provided**: skip intro, read the files immediately, begin cleanup
2. **If no parameters provided**, auto-detect the plan:
   - Get the current branch name: `git branch --show-current`
   - Strip the prefix (e.g. `feat/auth-setup` → `auth-setup`)
   - List files in `.claude/workspace/plans/` and find one whose filename contains the slug
   - Also check `.claude/workspace/research/` and `.claude/workspace/reviews/` for matching files
   - If a match is found: show what was detected, confirm with the user, then begin
   - If no match is found: ask the user for file paths

## Process Steps

### Step 1: Gather Context

1. **Read all provided files FULLY** (no limit/offset):
   - **Plan file** (required)
   - **Research file** (if provided)
   - **Review file** (if provided)

2. **Check for uncommitted changes** — they are part of the implementation:
   ```bash
   git status --porcelain
   git diff HEAD
   git diff --cached
   ```

3. **Analyze git history since plan creation (including uncommitted)**:
   ```bash
   git log --since="[plan date]" --oneline --no-merges
   git diff [plan-commit]..HEAD --stat
   git diff [plan-commit]..HEAD
   git diff HEAD --stat
   ```

4. **Read 8-10 most important changed files FULLY** — compare current state to what the plan described.

### Step 2: Rationalize — What Actually Happened?

Compare what was planned vs what was actually built:

1. **Plan vs reality**:
   - Where did implementation deviate from the plan? Why?
   - What discoveries changed the approach mid-implementation?
   - What alternatives were tried and rejected? What failed and why?

2. **Identify what's worth preserving** (max 5 items):
   - Key technical decisions with rationale (what was chosen, what was rejected, why)
   - Patterns or conventions that emerged during implementation
   - Lessons learned about architecture, design, testing, or deployment
   - Dead ends — approaches that don't work for this codebase

### Step 3: Rewrite Documentation to Reflect Final State

Update all documentation so it reads as if the current code was always the intended outcome. No "we changed X because the plan said Y" — just describe what the code does and why.

**This applies to ALL text — both existing and newly written.** Scan every line you've added or modified for "trail of changes" language (e.g., "this replaces the old X", "previously we used Y", "unlike the former approach").

1. **Code comments**: Review comments in changed files. They should explain *purpose*, not history. Remove any that reference the implementation journey.

2. **README.md**: Update if user-facing behavior, installation steps, or API usage changed. Describe current state as intended design.

3. **CLAUDE.md**: Update architecture sections, conventions, and patterns to match the final implementation. Document as established practice, not a recent change.

4. **Other project docs**: Rewrite affected sections to describe the current state cleanly.

**The rule**: A reader encountering any documentation for the first time should see a coherent, intentional design — not a trail of changes.

### Step 4: Capture Learnings in decisions.md

Update decisions via project-server (`add_decision`) with implementation learnings:

- Key decisions made (what was chosen, what was rejected, why)
- Patterns discovered during implementation
- Dead ends and failed approaches (prevents re-exploration)

Decisions are stored in the project-server database.

### Step 5: Update CLAUDE.md

For each pattern/convention to add:

1. **Identify the right section** (Architecture, Coding standards, Testing, Common Pitfalls)
2. **Add the update** with code examples and file:line references where relevant
3. **Create "Common Pitfalls" section** if it doesn't exist and there are pitfalls to document

### Step 6: Update todo.md

Mark completed items via project-server (`complete_todo`):
- Find items that were implemented and mark them as completed
- Do NOT modify other items — only complete what was actually implemented

### Step 7: Remove build artifacts

Delete the build review file if it exists:
- `.claude/workspace/build-logs/REVIEW.md`
- `.claude/workspace/build-logs/review-diff.patch`
- `.claude/workspace/build-logs/tooling.json`

These are temporary build artifacts and should not be committed.

### Step 8: Commit and Summary

1. **Commit the cleanup changes**:
   - Stage all documentation updates
   - Commit with: `docs: cleanup after <feature-name> implementation`

2. **Present summary**:
   ```
   Cleanup Complete

   ## Learnings Captured:
   - decisions.md: [summary of entries added/updated]

   ## Documentation Updated:
   - CLAUDE.md: [N] patterns/conventions added
   - todo.md: [N] items marked as completed
   - decisions.md: [updated with implementation learnings]
   - README.md: [updated if user-facing changes]

   Recommended next step: /pr to create PR description
   ```

## Important Guidelines

### Be Investigative
- Look for evidence in code, git history, tests, and comments
- **Include uncommitted changes** — they are part of the implementation
- Check commit messages for context on changes
- Always check `git status` and `git diff HEAD` for uncommitted work

### Distinguish Facts from Inferences
- **Facts**: "The code does X at file.py:123"
- **Inferences**: "This approach was likely chosen because Y"
- Mark inferences clearly when rationale isn't documented

### Preserve Clean Narratives
- decisions.md documents reasoning and alternatives
- CLAUDE.md captures reusable patterns
- Rejected alternatives prevent re-exploration
- Project documentation stays in sync

## Success Criteria

A cleanup is complete when:
- [ ] Implementation learnings captured in decisions.md
- [ ] New patterns/conventions are in CLAUDE.md
- [ ] Completed items marked in todo.md
- [ ] README.md updated if user-facing changes
- [ ] All documentation reads as coherent, intentional design (no trail of changes)
- [ ] Cleanup changes committed
