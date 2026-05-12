# codebench

A Claude Code workbench you install into any project. One curl drops in slash commands, skills, sub-agents, safety hooks, prebuilt MCP servers, library docs, rules, helpers, and a YAML pipeline runner.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash
```

Pin to a branch / tag, install elsewhere, preview without writing:

```bash
curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash -s -- --branch v0.1.0
curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash -s -- /path/to/project
curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash -s -- --dry-run
```

The installer drops `.claude/` (commands, skills, agents, hooks with prebuilt Go binaries, helpers, rules, templates, library, settings, pipelines, code-index MCP server binary) and `.mcp.json` into the target, creates runtime dirs (`logs/`, `memories/`, `index/`, `index-cache/`), seeds a gitignored `.env`, and appends `.claude`, `.mcp.json`, `.env`, `.playwright/`, `.playwright-mcp/`, `READTHISFIRST.md` to `.gitignore`.

Existing files the installer ships are overwritten; anything you added stays. Re-run the curl line any time to update.

After install, see `READTHISFIRST.md` in the target for the first-run checklist (GitHub token, branch protection, optional `CONTEXT7_API_KEY`).

## What's in the box

### Slash commands (`.claude/commands/`)

User-invoked workflows.

| Command | Purpose |
| --- | --- |
| `/research` | Parallel sub-agent investigation of a codebase question, synthesised into one answer |
| `/plan` | Interactive, iterative implementation planning — skeptical, thorough, file:line references |
| `/build` | Execute an approved plan from `.claude/memories/` phase by phase against its success criteria |
| `/review` | Senior-engineer code review (quality, security, performance, maintainability) |
| `/pr-review` | Multi-agent PR review: 4 core agents always, up to 6 specialised agents picked from the diff |
| `/code-analysis` | Run code-index analysis (hotspots, coupling, unhandled errors, dead code, circular deps) |

### Skills (`.claude/skills/`)

Some user-invoked, some model-invoked when the work matches.

| Skill | Purpose |
| --- | --- |
| `/pr` | Generate PR description, sync branch, push, create or update via `gh` |
| `/ship` | Commit, push, open PR, comment, squash-merge, return to default branch |
| `/release` | Production release: changelog, version bump, PR, merge, tag (two-branch or single-branch) |
| `setup-release` | Scaffold `scripts/changelog-release.sh`, GitHub Actions release workflow, Makefile targets |
| `/cleanup` | Post-implementation cleanup: rationalise docs, capture decisions, update project state |
| `/vulnerability-check` | Scan deps against OSV, GitHub Advisory, CISA KEV, NCSC |
| `front-end-design` | Creative direction for distinctive frontends — avoid generic AI aesthetics |
| `ui-component-creator` | Structural patterns for React / TypeScript components |
| `mobile-friendly-design` | Responsive web patterns: phone, tablet, desktop |
| `accessibility` | WCAG 2.2 patterns: semantic HTML, ARIA, keyboard nav, contrast, motion safety |
| `visual-verify` | Render-and-screenshot verification via the Playwright MCP server |
| `api-tools` | Author / maintain Bruno API collections (Git-native Postman alternative) |

### Sub-agents (`.claude/agents/`)

Specialised agents spawned by slash commands. Each is a markdown file with a tool allowlist and a system prompt.

- **Code understanding**: `codebase-analyzer`, `codebase-pattern-finder`, `code-simplifier`
- **PR review fleet**: `pr-security`, `pr-code-quality`, `pr-best-practices`, `pr-test-coverage`, `pr-breaking-changes`, `pr-data-integrity`, `pr-error-handling`, `pr-observability`, `pr-privacy`, `pr-compliance`
- **Planning / quality**: `quality-risk-analyzer`, `plan-validator`, `system-architect`
- **Research**: `web-researcher`, `documentation-researcher`, `library-analyzer`, `project-context-analyzer`, `compliance-research`
- **Ops**: `oncall-guide`, `verify-app`

### Hooks (`.claude/hooks/`, source in `hooks-logic/`)

Two Go binaries Claude Code runs on every tool call. Built for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64` and shipped prebuilt in `.claude/hooks/binaries/`.

- **`pre-tool-use`** — denies dangerous commands (`rm -rf /`, fork bombs, `dd of=/dev/sda`, `curl | sh`, reverse shells, container escapes, credential exfil, force-push to main, `git reset --hard`, sensitive file reads, cloud metadata probes), enforces deny patterns from `settings.json`, and **rewrites common CLI commands through `gtk`** to cut output tokens 60-94%. Plain bash is rewritten transparently — no need to type `gtk` yourself.
- **`post-tool-use`** — logs WebSearch / WebFetch results to `.claude/logs/` and (optionally) forwards them to a project-server backend if configured.

Source lives in `hooks-logic/pre-tool-use/` and `hooks-logic/post-tool-use/`. See [Developing](#developing) below for the build flow.

### MCP servers (`.mcp.json`)

| Server | Source | Purpose |
| --- | --- | --- |
| `code-index` | bundled Go binary in `.claude/mcp-index-server/` | Tree-sitter AST indexer for Python, JS/TS, Go, C/C++, Rust — `find_symbol`, `find_usage`, `get_call_graph`, `trace_data_flow`, `find_unhandled_errors`, `find_hotspots`, `analyze_coupling`, etc. Auto-indexes on startup, watches for changes, writes `.claude/index/*.md`. |
| `context7` | `npx @upstash/context7-mcp` | Up-to-date library docs. Resolve a library ID, then query docs — used before guessing an API. |
| `playwright` | `npx @playwright/mcp` (headless Chromium) | Visual verification, form filling, accessibility-tree inspection. |

The `code-index` server keeps `.claude/index/` fresh on every file change and branch switch — agents treat its markdown indexes as the canonical project map (see `.claude/rules/session-startup.md`).

### Pipelines (`.claude/pipelines/`)

YAML-driven multi-step runner for `claude-safe` and `claude-server`. Each step launches a fresh container — i.e. fresh Claude context — so a pipeline chains prompts that hand off artifacts on disk rather than conversation history.

```bash
./.claude/pipelines/pipeline.py .claude/pipelines/research-plan.yaml "rate limiting strategies"
./.claude/pipelines/pipeline.py .claude/pipelines/build.yaml "plans/foo.md"
./.claude/pipelines/pipeline.py .claude/pipelines/research-plan.yaml --model sonnet "topic"
./.claude/pipelines/pipeline.py .claude/pipelines/build.yaml --runner safe "plans/foo.md"
```

Steps run locally (`runner: safe` → `claude-safe`), remotely on a dev VM (`runner: server` → `claude-server`), or as plain shell on the host (`command:`, regardless of runner). The runner prints a colored stream of each step's tool calls to stderr, inherits the TTY for interactive `safe` steps, and auto-`rsync`s the workspace back after every server step.

Bundled pipelines:

| File | Runner | What it does |
| --- | --- | --- |
| `.claude/pipelines/research-plan.yaml` | safe | `/research <topic>` → `/plan` (interactive) |
| `.claude/pipelines/build.yaml` | server | local shell setup (branch + PR via `scripts/build_setup.py`) → `/build` → `/pr-review` → triage |

YAML schema:

```yaml
name: <pipeline name>            # required
description: <one-liner>         # optional
runner: safe|server              # optional, default `safe` — applies to prompt: steps
runner_args: ["--github", ...]   # optional, default [] — passed through to claude-server

steps:
  - id: <step id>                # required, used in {{ steps.<id>.output }}
    interactive: true|false      # required (ignored for server / command, but must be set)
    output: <path template>      # optional, exposed as {{ output }} inside this step
    # exactly one of:
    prompt: |                    # LLM step; runner-dispatched
      ...
    command: |                   # shell step on host (bash -c), regardless of runner
      ...
```

Template variables: `{{ input }}`, `{{ timestamp }}`, `{{ pipeline_dir }}`, `{{ output }}`, `{{ steps.<id>.output }}`. Unknown variables raise `ValueError`. `output` is a path template, not a promise — the prompt / command has to write there.

Before every `prompt:` step the runner shells out to `.claude/helpers/get_metadata.sh` (single source of truth for date/repo/branch/commit), appends a `GitHub owner/repo:` line derived from `origin`, and passes the result via `--append-system-prompt`. Recomputed per step, so post-`setup` steps see the new branch.

### Library (`.claude/library/`)

Reference material agents pull into context when relevant.

- `best_practices/` — accessibility, api-design, authorization, container-security, data-integrity, error-handling, layered-architecture, llm-integration-patterns, observability, performance, privacy-by-design, resilience, structured-logging, testing-strategy, zero-downtime-deployment, and more
- `compliance_rules/` — audit-trail, auth-boundaries, configuration-security, cryptography, data-lifecycle, gdpr-*, resilience, secure-coding, secure-development, session-cookie, supply-chain, plus a `standards-index.md` mapping rules to ISO 27001 / NIS2 / OWASP ASVS / GDPR
- `security_rules/core/`, `security_rules/owasp/` — OWASP Top 10 and core security rules
- `documentation/` — evidence documents the compliance agents cite

### Rules (`.claude/rules/`)

Per-session instructions auto-included via `CLAUDE.md`. The interesting ones:

- `engineering-principles.md` — KISS, YAGNI, fight bloat, fail fast, work in steps, naming matters
- `caveman.md` — token-optimised output style for commits, reviews, PRs, and chat
- `mcp-servers.md` — when to use `code-index` instead of grep, Context7 workflow, Playwright host.docker.internal note
- `gtk.md` — how the pre-tool-use hook rewrites commands through `gtk`
- `commits.md`, `branching.md`, `plan-mode.md`, `claude-settings.md`, `session-startup.md`, `scripts.md`

### Helpers (`.claude/helpers/`)

Standalone scripts callable from commands, skills, or pipelines:

- `setup-github-token.py` — interactive `gh auth` flow, writes `GH_TOKEN` into `.env`
- `setup-branch-protection.py` — protect `main` (require PRs, block direct pushes)
- `get_metadata.sh` — date / time / repo / branch / commit, single source of truth
- `notify.sh` — desktop notification used by review hooks
- `codebase-graph/` — Go program that renders an interactive HTML dependency / call graph (`go run .claude/helpers/codebase-graph/main.go`)
- `vulnerability-check/` — backing scripts for the `/vulnerability-check` skill

## Requirements

- [`uv`](https://github.com/astral-sh/uv) — `pipeline.py`, helpers, and pipeline scripts are `uv run --script` shebangs with PEP 723 inline metadata
- [`claude-safe`](examples/claude-safe) on `PATH` — needed for `runner: safe` pipelines and `--no-firewall` runs
- [`claude-server`](examples/project-server-main) on `PATH` — needed for `runner: server` pipelines
- `gh` for PR / release skills, `git` for everything

The MCP servers come prebuilt: `code-index` ships as a Go binary in `.claude/mcp-index-server/` for `linux/x86_64`, `linux/aarch64`, `darwin/arm64`; `context7` and `playwright` are `npx`-launched. No global installs required.

## Developing

This repo is the source for what gets installed. Most contributions are markdown — commands, skills, agents, rules, library docs — and need nothing but an editor.

The Go pieces need a build step:

```bash
make build-hooks   # cross-compile pre-tool-use + post-tool-use for 4 platforms into .claude/hooks/binaries/
make test-hooks    # build, then run Go tests and the pre-tool-use shell test suite
```

Hook source: `hooks-logic/pre-tool-use/main.go` (deny patterns + gtk rewrites), `hooks-logic/post-tool-use/main.go` (WebSearch/WebFetch logging + optional forwarding). Both have `main_test.go` next to them; `pre-tool-use/test.sh` runs end-to-end shell cases through the binary.

The `gtk` filter binaries live in `.claude/hooks/gtk/` and are platform-suffixed (`gtk-linux-amd64`, etc.). They are not built from this repo — they are vendored.

The code-index MCP server lives in `.claude/mcp-index-server/` as a prebuilt binary plus `start.sh` (selects the right binary) and `update.sh` (pulls the latest release).

Pipeline runner: single file at `.claude/pipelines/pipeline.py`. No venv, no `pip install`, no separate test suite — `uv` resolves `pyyaml` on first run.

### Repo layout

```
.claude/agents/            # sub-agent definitions (markdown + tool allowlists)
.claude/commands/          # slash commands invoked by users / inside pipeline steps
.claude/helpers/           # standalone scripts (token setup, branch protection, metadata, ...)
.claude/hooks/             # hook entry shims + prebuilt Go binaries
.claude/hooks/gtk/         # vendored gtk binaries (output filter)
.claude/library/           # reference docs agents consult (best practices, compliance, security)
.claude/mcp-index-server/  # prebuilt code-index MCP server binary
.claude/pipelines/         # pipeline.py runner + bundled YAMLs + scripts/ helpers
.claude/rules/             # per-session instructions auto-included via CLAUDE.md
.claude/settings.json      # permissions, hook config, enabled MCP servers
.claude/skills/            # model-invocable and user-invocable skills
.claude/templates/         # output templates referenced by commands
hooks-logic/               # Go source for pre-tool-use and post-tool-use hooks
install.sh                 # bootstrap installer (downloads tarball, runs install-helper.sh)
install-helper.sh          # actual install logic
Makefile                   # build-hooks, test-hooks
examples/                  # reference repos (claude-safe, claude-setup, sandcastle, project-server-main) -- not part of the product
```

## Why fresh contexts per pipeline step

Long-running Claude sessions accumulate context that biases later turns. Splitting a workflow across containers gives each step a clean slate while the YAML pins down what artifact each step must hand to the next. Interactive `safe` steps still get a normal TTY for parts that need a human in the loop.

## License

MIT. See [LICENSE](LICENSE). Use, fork, ship, sell — whatever you want.
