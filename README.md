# pipeline

YAML-driven multi-step pipeline runner for [`claude-safe`](examples/claude-safe) and [`claude-server`](examples/project-server-main). Each step launches a fresh container — i.e. fresh Claude context — so a pipeline chains prompts that hand off artifacts on disk rather than conversation history.

Steps run either locally (`runner: safe` → `claude-safe`), remotely on a dev VM (`runner: server` → `claude-server`), or as plain shell on the host (`command:`, regardless of runner).

## Install

One-line install into the current directory:

```bash
curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash
```

Pin to a branch or tag, install elsewhere, or preview without writing:

```bash
curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash -s -- --branch v0.1.0
curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash -s -- /path/to/project
curl -fsSL https://raw.githubusercontent.com/albertsikkema/codebench/main/install.sh | bash -s -- --dry-run
```

The installer drops `.claude/` (skills, commands, helpers, hooks with prebuilt binaries, rules, templates, settings, pipelines) and `.mcp.json` into the target, creates the runtime dirs, and adds `.claude`, `.mcp.json`, `.env`, `.playwright/`, `.playwright-mcp/` to `.gitignore` (creating it if needed) so they don't get committed. Existing files in `.claude/` that the installer ships are overwritten; anything you added stays.

Re-run the curl one-liner any time to pull updates — it always fetches the latest version of the requested branch / tag.

## Requirements

- [`uv`](https://github.com/astral-sh/uv) — `pipeline.py` is a `uv run --script` shebang with PEP 723 inline metadata; `pyyaml` resolves on first run
- [`claude-safe`](examples/claude-safe) on `PATH` for `runner: safe` pipelines
- [`claude-server`](examples/project-server-main) on `PATH` for `runner: server` pipelines

No venv, no `pip install`, no tests, no CI.

## Usage

```bash
./.claude/pipelines/pipeline.py <pipeline.yaml> <input...>
```

Examples:

```bash
# Local: research then plan, fresh contexts
./.claude/pipelines/pipeline.py .claude/pipelines/research-plan.yaml "rate limiting strategies"

# Override model
./.claude/pipelines/pipeline.py .claude/pipelines/research-plan.yaml --model sonnet "topic"
CLAUDE_MODEL=sonnet ./.claude/pipelines/pipeline.py .claude/pipelines/research-plan.yaml "topic"

# Force the server pipeline to run locally
./.claude/pipelines/pipeline.py .claude/pipelines/build.yaml --runner safe "plans/foo.md"
```

The runner prints a colored, summarized stream of each step's tool calls and assistant text to stderr. Interactive `safe` steps inherit the TTY directly. Server steps live-tail the remote container's docker logs and auto-`rsync` the workspace back when the step exits.

## Bundled pipelines

| File | Runner | What it does |
| --- | --- | --- |
| `.claude/pipelines/research-plan.yaml` | safe | `/research <topic>` → `/plan` (interactive) |
| `.claude/pipelines/build.yaml` | server | shell setup (branch + PR via [`scripts/build_setup.py`](.claude/pipelines/scripts/build_setup.py)) → `/build` → `/pr-review` → triage the review comment |

## Pipeline format

```yaml
name: <pipeline name>            # required
description: <one-liner>         # optional
runner: safe|server              # optional, default `safe` — applies to prompt: steps
runner_args: ["--github", ...]   # optional, default [] — passed through to claude-server

steps:
  - id: <step id>                # required, used in {{ steps.<id>.output }}
    interactive: true|false      # required on every step (ignored for server / command)
    output: <path template>      # optional, exposed as {{ output }} inside this step
    # exactly one of:
    prompt: |                    # LLM step; runner-dispatched
      ...
    command: |                   # shell step on host (bash -c), regardless of runner
      ...
```

Rules enforced at load time:

- `name` and `steps` are required.
- Every step must set `interactive` explicitly. The flag is ignored for `command:` steps and `runner: server` (no local TTY in either case), but it must still be present.
- A step must set exactly one of `prompt:` or `command:`.
- `runner` must be `safe` or `server`.

`output` is a path template, not a promise — the runner renders it and exposes it as `{{ output }}` inside the step body, but does not create or verify the file. The prompt or command must write there.

### Template variables

| Variable | Meaning |
| --- | --- |
| `{{ input }}` | Joined CLI positional args |
| `{{ timestamp }}` | Pipeline start time (`%Y-%m-%dT%H-%M-%S`) |
| `{{ pipeline_dir }}` | Absolute path to the directory containing the YAML — use for invoking helpers like `{{ pipeline_dir }}/scripts/foo.py` so pipelines work regardless of the caller's cwd |
| `{{ output }}` | The current step's rendered `output` (only available inside its body) |
| `{{ steps.<id>.output }}` | Earlier step's rendered `output` |

Unknown variables raise `ValueError`.

### Session context (auto-injected)

Before every `prompt:` step, the runner shells out to [`.claude/helpers/get_metadata.sh`](.claude/helpers/get_metadata.sh) — single source of truth for date/time/repo/branch/commit — appends a `GitHub owner/repo:` line derived from `origin`, and passes the result via `--append-system-prompt`. Step prompts can reference owner / repo / date directly without calling `gh repo view` or `date`. Recomputed per step, so post-`setup` steps see the new branch.

### Server steps

`runner: server` dispatches via `claude-server -p <prompt> --follow`, captures the session ID printed as `==> Session: <id>`, then automatically runs `claude-server --sync <id>` to rsync the remote workspace back. Don't add a manual sync step — it's already there.

### Mixing local shell into a server pipeline

`command:` steps always run on the host. `build.yaml` uses this to create the branch + PR locally (so the PR lives in the host's clone) before any server LLM step runs; `claude-server` then rsyncs that branch state to the dev VM.

## Bundled commands and skills

Slash commands (`.claude/commands/`):

| Command | Purpose |
| --- | --- |
| `/research` | Spawn parallel sub-agents to investigate a question across the codebase, then synthesize findings |
| `/plan` | Interactive, iterative implementation planning -- skeptical, thorough, collaborative |
| `/build` | Execute an approved plan from `.claude/memories/` phase by phase, against its success criteria |
| `/review` | Senior-engineer code review for quality, security, performance, maintainability |
| `/pr-review` | Multi-agent PR review: 4 core agents always, plus up to 6 specialized agents picked from the diff |
| `/code-analysis` | Run the code-index MCP analysis tools (hotspots, coupling, unhandled errors, dead code, ...) on a scope |

Skills (`.claude/skills/`):

| Skill | Purpose |
| --- | --- |
| `/pr` | Generate a PR description, sync the branch, push, and create or update the PR via `gh` |
| `/ship` | Commit, push, open a PR, comment, merge with squash, return to the default branch |
| `/release` | Create a production release: changelog, version bump, PR, merge, tag (two-branch or single-branch) |
| `setup-release` | Scaffold `scripts/changelog-release.sh`, the GitHub Actions release workflow, and Makefile targets in a fresh project |
| `/cleanup` | Post-implementation cleanup -- rationalize docs, capture decisions and learnings, update project state |
| `/vulnerability-check` | Scan dependencies for known vulnerabilities (OSV, GitHub Advisory, CISA KEV, NCSC) |
| `front-end-design` | Creative direction for distinctive, production-grade frontend interfaces; avoid generic AI aesthetics |
| `ui-component-creator` | Structural patterns for React/TypeScript UI components (use with the project's design tokens) |
| `mobile-friendly-design` | Responsive web patterns -- phone, tablet, desktop; mobile nav, touch targets, responsive layouts |
| `accessibility` | WCAG 2.2 patterns: semantic HTML, ARIA, keyboard nav, color contrast, motion safety, screen readers |
| `visual-verify` | Render-and-screenshot verification of web UI through the Playwright MCP server |
| `api-tools` | Author and maintain Bruno API collections (Git-native, offline-first Postman alternative) |

`/release`, `/ship`, `/pr`, `/cleanup`, `/vulnerability-check` are user-invocable slash commands. The design / UI skills are model-invoked when relevant frontend work shows up.

## Why fresh contexts per step

Long-running Claude sessions accumulate context that biases later turns. Splitting a workflow across containers gives each step a clean slate while the YAML pins down what artifact each step must hand to the next. Interactive `safe` steps still get a normal TTY for parts that need a human in the loop.

## Layout

```
.claude/pipelines/         # pipeline.py runner + bundled YAMLs
.claude/pipelines/scripts/ # helper scripts called from command: steps (PEP 723 uv scripts)
.claude/commands/          # slash commands invoked inside steps (/research, /plan, /build, /pr-review, ...)
.claude/templates/         # output templates referenced by those commands
.claude/helpers/           # get_metadata.sh, notify.sh, codebase-graph, ...
examples/                  # reference repos this builds on (claude-safe, claude-setup, sandcastle, project-server-main) — not part of the product
```
