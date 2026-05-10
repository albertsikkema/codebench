# Commit Rules

Follow the Conventional Commits specification for all commits.

## Format

```
<type>[optional scope]: <description>

[optional body]
```

## Types

| Type | Purpose |
|------|---------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code restructuring without behavior change |
| `docs` | Documentation changes |
| `style` | Formatting changes (whitespace, semicolons) |
| `test` | Adding or updating tests |
| `chore` | Maintenance (dependencies, config) |
| `perf` | Performance improvements |
| `build` | Build system or dependency changes |
| `ci` | CI/CD configuration changes |

## Subject Line Rules

- 50 characters max
- **Lowercase** the first letter after the prefix (`feat: add feature` not `feat: Add feature`)
- Imperative mood ("add feature" not "added feature")
- No period at the end
- Wrap body at 72 characters

## Process

- Stage specific files by name (never `git add -A` or `git add .`)
- Prefer atomic commits — one logical change per commit
- When a session touches multiple unrelated areas, split into separate commits
- Use the body to explain what and why, not how
- Reference the task ID (T-NNN) in the commit message when one exists

## Examples

Good:
```
feat(auth): add password reset functionality

Users can now reset their password via email link.
Addresses user feedback about account recovery.
```

```
fix: resolve race condition in order processing

The previous implementation could process the same
order twice under high load conditions.
```

```
refactor(api): extract validation into middleware
```

Bad:
- "Tweaked a few things" -- vague, no type
- "fixed bug" -- no type, no useful description
- "feat: Added the new feature that allows users to do the thing" -- too long, past tense, sentence-case
