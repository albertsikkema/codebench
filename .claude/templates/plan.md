---
date: [date from get_metadata.sh]
author: [Current user from get_metadata.sh]
git_commit: [Current commit hash from get_metadata.sh]
branch: [Current branch name from get_metadata.sh]
repository: [Repository name from get_metadata.sh]
topic: "[Feature/Task Name]"
status: draft
---

# [Feature/Task Name] Implementation Plan

## Overview

[Brief description of what we're implementing and why]

## Current State Analysis

[What exists now, what's missing, key constraints discovered]

### Key Discoveries:
- [Important finding with file:line reference]
- [Pattern to follow]
- [Constraint to work within]

## Desired End State

[A specification of the desired end state after this plan is complete, and how to verify it]

## What We're NOT Doing

[Explicitly list out-of-scope items to prevent scope creep]

## Implementation Approach

[High-level strategy and reasoning]

### Design Decisions
[If multiple viable approaches exist, present them here with pros/cons for each. Flag which decisions need user input before implementation can proceed.]

| Decision | Options | Recommendation | Status |
|----------|---------|----------------|--------|
| [Decision 1] | [A vs B] | [Recommended option + rationale] | Pending |

## Phase 1: [Descriptive Name]

### Overview
[What this phase accomplishes]

### Changes Required:

#### 1. [Component/File Group]
**File**: `path/to/file.ext`
**Changes**: [Summary of changes]

```[language]
// Specific code to add/modify
```

### Success Criteria:

#### Automated Verification:
- [ ] Tests pass: `[test command]`
- [ ] Type checking passes: `[typecheck command]`
- [ ] Linting passes: `[lint command]`

#### Manual Verification:
- [ ] Feature works as expected when tested
- [ ] No regressions in related features

---

## Phase 2: [Descriptive Name]

[Similar structure with both automated and manual success criteria...]

---

## Testing Strategy

### Unit Tests:
- [What to test]
- [Key edge cases]

### Integration Tests:
- [End-to-end scenarios]

### Manual Testing Steps:
1. [Specific step to verify feature]
2. [Another verification step]

## Quality Considerations

### Security Design
[Based on Quality Context from research. Reference specific rules from .claude/library/security_rules/.]
- Authentication/authorization approach for this feature
- Input validation strategy (what inputs, what validation)
- Data sensitivity and handling (what's sensitive, how it's protected)

### Error Handling Strategy
[Based on Quality Context from research. Reference existing patterns.]
- Expected failure modes and how each is handled
- External dependency failures (APIs, databases, file system)
- User input validation and error messages

### Edge Cases
[Based on Quality Context from research. Be specific, not generic.]
- Boundary conditions (empty inputs, large inputs, concurrent access)
- State edge cases (partial failures, interrupted operations)
- Data edge cases (unicode, special characters, null/missing)

### Testability Notes
[Based on Quality Context from research. Reference existing test patterns.]
- How each component will be tested in isolation
- Dependencies that need mocking/stubbing
- Which existing test patterns to follow

## API Testing Updates

**If the project has an `api_tools/` directory AND the plan involves API endpoint changes:**

### Endpoints Modified:
- [ ] `METHOD /path` — Update `path/to/file.bru` (body/headers/URL changes)

### New Endpoints:
- [ ] `METHOD /path` — Create `collection/folder/name.bru`

### Test Coverage:
- [ ] Ensure changed/new endpoints have `tests {}` blocks

**If no API endpoints are affected, omit this section entirely.**

## Makefile Updates

**If a `Makefile` exists in the project root AND the plan adds new runnable tasks:**

1. Read the existing Makefile and follow its conventions.
2. Only add/update targets that directly correspond to new functionality in the plan.
3. Do NOT add convenience wrappers, aliases, or targets "for completeness".

**If the Makefile is unaffected, omit this section entirely.**

## Performance Considerations

[Any performance implications or optimizations needed]

## Migration Notes

[If applicable, how to handle existing data/systems]

## References

- Related research: `.claude/memories/[relevant].md`
- Similar implementation: `[file:line]`
