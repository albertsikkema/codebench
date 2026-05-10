# gtk — Token-Optimized CLI Proxy (automatic)

gtk filters command output to reduce LLM token consumption by 60-94%.
The consolidated PreToolUse hook binary (`hooks-logic/pre-tool-use/`)
automatically rewrites your Bash commands to use gtk — you do NOT need
to prefix commands with `gtk` yourself.

## How it works

Just run commands normally:

```bash
git log -10       # Automatically rewritten to use gtk → compact output
cargo test        # Automatically rewritten → failures only
git status        # Automatically rewritten → compact status
```

The hook rewrites commands for: git, cargo, go, gh, docker, kubectl,
npm, pnpm, npx, pytest, ruff, mypy, pip, dotnet, eslint, tsc, prettier,
vitest, playwright, prisma, next, curl, ls, find, grep.

Commands with pipes (`|`) or chains (`&&`, `;`) are left unchanged.

## Token savings hint

Every filtered command prints a savings hint to stderr:
```
[gtk: 2054 → 118 tokens, 94% saved]
```

## Bypass (IMPORTANT)

If filtered output seems incomplete or you need full output for debugging,
use the gtk binary directly with `proxy`:

```bash
.claude/hooks/gtk/gtk-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') proxy git log -10
```

**When to bypass:**
- Filtered output doesn't explain a failure
- You need to see passing tests (not just failures)
- You need full diff content (not just stats)
- A warning or log line might be relevant to the issue
- Output seems truncated or missing expected information
