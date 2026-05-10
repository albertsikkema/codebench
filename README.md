# pipeline

YAML-driven multi-step pipeline runner for [`claude-safe`](examples/claude-safe) and [`claude-server`](examples/project-server-main). Each step launches a fresh container — i.e. fresh Claude context — so a pipeline chains prompts that hand off artifacts on disk rather than conversation history.

Steps run either locally (`runner: safe` → `claude-safe`), remotely on a dev VM (`runner: server` → `claude-server`), or as plain shell on the host (`command:`, regardless of runner).

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
