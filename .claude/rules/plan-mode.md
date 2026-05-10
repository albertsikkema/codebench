# Enhanced Plan Mode

When plan mode is entered (via EnterPlanMode), follow this enhanced research-driven methodology instead of the default. This applies to every plan mode session.

---

## In Plan Mode: Pre-Flight

1. **Clarify intent**: Use AskUserQuestion to ask targeted questions about scope, constraints, and expected behavior. Understand what the user actually wants before researching.
2. **Identify libraries/frameworks** involved — note them for Context7 lookups.

### File Path Rule

Always reference files with their **full relative path** from the project root — never bare filenames. This ensures agents can open files directly without searching.

- Good: `.claude/pipelines/scripts/build_setup.py`, `.claude/helpers/setup-github-token.py`, `.claude/pipelines/pipeline.py`
- Bad: `build_setup.py`, `setup-github-token.py`, `pipeline.py` (without path when ambiguous)

This applies everywhere in the plan: Current State Analysis, Implementation Approach, phase descriptions, code snippets, and references.

---

## In Plan Mode: Research Phase

### A. Map the Codebase

Use the `code-index` MCP tools (`search_symbols`, `find_symbol`, `get_file_outline`, `get_project_summary`) to find entry points, relevant functions, classes, and patterns. Read all relevant source files identified.

### B. Context7 Library Lookups

For EACH library/framework involved:
1. `resolve-library-id` to get the Context7 ID
2. `query-docs` with specific questions (API signatures, config, patterns, version-specific behavior, pitfalls)

Do this BEFORE spawning agents — library docs inform better agent prompts.

### C. Spawn Research Agents in Parallel

Launch these read-only agents (they only use Read/Glob/Grep — compatible with plan mode):

| Agent | Purpose | When |
|-------|---------|------|
| `codebase-analyzer` | How relevant code works (pass file:line refs from index scan) | Always |
| `quality-risk-analyzer` | Security patterns, error handling, edge cases, testability | Always |
| `web-researcher` | Current best practices, known issues, version-specific gotchas | When external libs involved |
| `codebase-pattern-finder` | Similar implementations to follow | When building something the codebase already does variants of |

**Fallback**: If agents cannot be spawned in plan mode, use the Explore agent with detailed prompts covering the same research areas.

### D. Synthesize Research

Combine all sources before writing the plan:
- **Context7** → code snippets, API usage, configuration patterns
- **Web research** → references, version-specific gotchas
- **Quality risks** → security, error handling, edge cases for Quality Considerations sections
- **Index + analyzer** → Current State Analysis with file:line references

---

## In Plan Mode: Planning Phase

### E. Write the Plan

Use the plan template from `/plan` command (`.claude/commands/plan.md`):
- Run `.claude/helpers/get_metadata.sh` for frontmatter
- Follow the full template structure: Overview, Current State, Desired End State, What We're NOT Doing, Implementation Approach, Phases with Success Criteria, Quality Considerations, References
- Include concrete code snippets and file:line references throughout
- Populate Quality Considerations from the quality-risk-analyzer findings

### F. Quality Self-Review (3 Passes)

As defined in the `/plan` command:
1. **Pass 1 — Code Snippet Review**: Check all code blocks for logic errors, data handling, resource management, code smells, language-specific issues
2. **Pass 2 — Referenced Code Audit**: Read actual code at each file:line reference, verify it matches what the plan describes
3. **Pass 3 — Design-Level Check**: Pattern consistency, security rules, helper reuse, layer violations, error handling coverage

Fix issues directly in the plan. Append a `## Quality Review` summary of what was checked and changed.

### G. Exit Plan Mode

Call ExitPlanMode for user approval. Present:
- Plan file location
- 2-3 sentence summary of the approach
- Phase list with one-line descriptions
- Any pending design decisions requiring user input

---

## After Plan Approval (normal mode)

### H. Branch Check
Verify not on main branch. Create a feature branch if needed (`feat/description` or `fix/description`).

### I. Implement
Execute each phase of the plan sequentially.

### J. Verify
Run tests and linting after implementation (use commands from Success Criteria).

### K. Commit
Use conventional commit format referencing the plan. No AI attribution — no Co-Authored-By lines, no "Generated with Claude Code" footers, no AI markers anywhere.
