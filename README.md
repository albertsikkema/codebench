# pipeline

YAML-driven multi-step pipeline runner for [`claude-safe`](examples/claude-safe). Each step is a fresh `claude-safe` container — i.e. fresh Claude context — so a pipeline chains prompts that hand off artifacts on disk rather than conversation history.

## Requirements

- [`claude-safe`](examples/claude-safe) installed and on `PATH`
- [`uv`](https://github.com/astral-sh/uv) (the script resolves `pyyaml` itself via PEP 723 inline metadata)

## Usage

```bash
./pipeline.py <pipeline.yaml> <input...>
```

Examples:

```bash
./pipeline.py pipelines/research-plan.yaml "rate limiting strategies for our API"
./pipeline.py pipelines/research-plan.yaml --model sonnet "topic"
CLAUDE_MODEL=sonnet ./pipeline.py pipelines/research-plan.yaml "topic"
```

The runner prints a colored, summarized stream of each step's tool calls and assistant text to stderr. Interactive steps inherit the TTY directly.

## Pipeline format

```yaml
name: research-and-plan
description: Run /research then /plan with fresh contexts.

steps:
  - id: research
    interactive: false
    output: .claude/memories/{{ timestamp }}-research.md
    prompt: |
      /research {{ input }}

      Save the research document to `{{ output }}` instead of the default path.

  - id: plan
    interactive: true
    output: .claude/memories/{{ timestamp }}-plan.md
    prompt: |
      /plan based on the research at `{{ steps.research.output }}`.

      Save the plan to `{{ output }}` instead of the default path.
```

Required keys: `name`, `steps`. Each step requires `id`, `interactive` (boolean), and `prompt`. `output` is optional but is the conventional way to pass artifacts to later steps.

### Template variables

| Variable | Meaning |
| --- | --- |
| `{{ input }}` | Joined CLI positional args |
| `{{ timestamp }}` | Pipeline start time (`%Y-%m-%dT%H-%M-%S`) |
| `{{ output }}` | Rendered `output` of the current step (only inside its `prompt`) |
| `{{ steps.<id>.output }}` | Rendered `output` of an earlier step |

`pipeline.py` does not create or verify the file at `output` — it's a path your prompt is expected to write to.

## Why fresh contexts per step

Long-running Claude sessions accumulate context that biases later turns. Splitting a workflow across containers gives each step a clean slate while the YAML pins down what artifact each step must hand to the next. Interactive steps still get a normal TTY for the parts that need a human in the loop.

## Layout

```
pipeline.py              # the runner
pipelines/               # YAML pipeline definitions
.claude/commands/        # slash commands invoked inside steps (e.g. /research, /plan)
.claude/templates/       # output templates referenced by those commands
examples/                # reference repos this builds on (claude-safe, claude-setup, sandcastle)
```
