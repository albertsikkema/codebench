# READTHISFIRST

You just installed [codebench](https://github.com/albertsikkema/codebench) into this project. This file is your map — read it once, then delete it.

## What got installed

- `.claude/` — slash commands, skills, helpers, hooks, rules, templates, pipelines, settings, and prebuilt Go binaries (`hooks/binaries/`, `hooks/gtk/`, `mcp-index-server/`)
- `.mcp.json` — MCP server config: `code-index`, `context7`, `playwright`
- `.env` — seeded with a commented `GH_TOKEN` placeholder (gitignored). `setup-github-token.py` writes the token here.
- `.gitignore` — appended `.claude`, `.mcp.json`, `.env`, `.playwright/`, `.playwright-mcp/`, `READTHISFIRST.md` so none of this gets committed
- `.claude/index/`, `.claude/index-cache/`, `.claude/logs/`, `.claude/memories/` — runtime dirs, populated as you go

The installer overwrites files it ships and leaves any additions of yours alone. Re-run the curl one-liner any time to pull updates.

## First-run checklist

1. **Get a GitHub token** so Claude can push branches and open PRs from this repo:
   ```
   uv run .claude/helpers/setup-github-token.py
   ```
2. **Protect main** (require PRs, block direct pushes):
   ```
   uv run .claude/helpers/setup-branch-protection.py
   ```
3. **(Optional)** export `CONTEXT7_API_KEY` in your shell profile to raise Context7 rate limits. Context7 works without it; the key just lifts the throttle.

## Slash commands

| Command | What it does |
| --- | --- |
| `/research` | Parallel sub-agent investigation of a codebase question |
| `/plan` | Interactive, iterative implementation planning |
| `/build` | Execute an approved plan phase by phase |
| `/review` | Senior-engineer code review (quality, security, performance) |
| `/pr` | Generate a PR description, push, open or update the PR |
| `/pr-review` | Multi-agent PR review (4 core + up to 6 specialized) |
| `/ship` | Commit, push, open PR, merge with squash, return to default |
| `/release` | Changelog, version bump, PR, merge, tag |
| `/cleanup` | Rationalize docs and capture decisions after an implementation |
| `/vulnerability-check` | Scan dependencies (OSV, GitHub Advisory, CISA KEV, NCSC) |
| `/code-analysis` | Run code-index analysis (hotspots, coupling, unhandled errors, ...) |

Model-invoked skills (run automatically when relevant): `front-end-design`, `ui-component-creator`, `mobile-friendly-design`, `accessibility`, `visual-verify`, `api-tools`, `setup-release`.

## Pipelines

Multi-step workflows that hand artifacts between fresh Claude containers (no shared conversation history):

```
./.claude/pipelines/pipeline.py .claude/pipelines/research-plan.yaml "your topic"
./.claude/pipelines/pipeline.py .claude/pipelines/build.yaml "plans/foo.md"
```

Override the model with `--model sonnet` or `CLAUDE_MODEL=sonnet`. Force a server pipeline to run locally with `--runner safe`. Pipelines need `claude-safe` (always) and `claude-server` (only for `runner: server`).

## MCP servers

- **`code-index`** — tree-sitter AST indexer; use `find_symbol`, `find_usage`, `get_file_outline`, `get_project_summary`, `trace_data_flow` instead of grep when navigating code
- **`context7`** — live library docs; `resolve-library-id` then `query-docs` before guessing an API
- **`playwright`** — headless browser for visual verification; pairs with the `visual-verify` skill

## Where to look when stuck

- `.claude/rules/` — project-wide conventions Claude reads at session start (commits, branching, gtk, MCP usage, plan-mode methodology, engineering principles)
- `.claude/commands/<name>.md` and `.claude/skills/<name>/` — full source of every slash command and skill
- `.claude/pipelines/` — bundled YAMLs (`research-plan.yaml`, `build.yaml`) as reference for writing your own
- [github.com/albertsikkema/codebench](https://github.com/albertsikkema/codebench) — upstream README and source

Delete this file once you've read it.
