#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = ["pyyaml"]
# ///
"""YAML-driven pipeline runner for claude-safe / claude-server.

Each step is a fresh container (= fresh context). The runner is selected per
pipeline (`runner: safe|server` in the YAML, default `safe`):

- `safe`   → local `claude-safe`. Steps can be non-interactive (stream-json
             piped through a formatter) or interactive (TTY inherited,
             prompt passed as positional arg).
- `server` → remote `claude-server`. Always uses `-p <prompt> --follow`
             (live-tails docker logs until the container exits) and then
             auto-syncs the workspace back via `claude-server --sync`. The
             step's `interactive` flag is ignored — there is no local TTY.

Usage:
    pipeline.py <pipeline.yaml> <input...>
    pipeline.py <pipeline.yaml> --runner server <input...>
"""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
import threading
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path

import yaml

# ---------------------------------------------------------------------------
# ANSI colors
# ---------------------------------------------------------------------------
BOLD = "\033[1m"
GREEN = "\033[0;32m"
CYAN = "\033[0;36m"
YELLOW = "\033[0;33m"
DIM = "\033[2m"
RED = "\033[0;31m"
WHITE = "\033[0;37m"
MAGENTA = "\033[0;35m"
BLUE = "\033[0;34m"
RESET = "\033[0m"


# ---------------------------------------------------------------------------
# claude-safe invocation (extracted from claude-setup _shared.py)
# ---------------------------------------------------------------------------
CLAUDE_SAFE_CMD = ["claude-safe", "--no-firewall", "--"]

_output_lock = threading.Lock()


def _format_tool_call(tool_name: str, tool_input: dict | str) -> str:
    if not isinstance(tool_input, dict):
        return str(tool_input)[:100]

    if tool_name == "Bash":
        desc = tool_input.get("description", "")
        cmd = tool_input.get("command", "")
        if desc:
            first_line = cmd.split("\n")[0] if cmd else ""
            if first_line and first_line != desc:
                return f"{desc[:120]}\n      {DIM}$ {first_line[:80]}{RESET}"
            return desc[:120]
        first_line = cmd.split("\n")[0] if cmd else ""
        return first_line[:80] + ("..." if len(first_line) > 80 else "")

    if tool_name == "Read":
        return tool_input.get("file_path", "")[:120]

    if tool_name == "Write":
        fp = tool_input.get("file_path", "")[:120]
        content = tool_input.get("content", "")
        line_count = len(content.split("\n"))
        return f"{fp} ({line_count} lines)"

    if tool_name == "Edit":
        fp = tool_input.get("file_path", "")[:120]
        old = tool_input.get("old_string", "")
        new = tool_input.get("new_string", "")
        old_lines = len(old.split("\n"))
        new_lines = len(new.split("\n"))
        return f"{fp} ({old_lines} → {new_lines} lines)"

    if tool_name == "Grep":
        pattern = tool_input.get("pattern", "")
        path = tool_input.get("path", "")
        return f"/{pattern}/ in {path}" if path else f"/{pattern}/"

    if tool_name == "Glob":
        pattern = tool_input.get("pattern", "")
        path = tool_input.get("path", "")
        return f"{pattern} in {path}" if path else pattern

    parts = []
    for k, v in list(tool_input.items())[:3]:
        v_str = str(v)
        max_len = 80 if k in ("file_path", "path", "pattern") else 50
        if len(v_str) > max_len:
            v_str = v_str[:max_len] + "..."
        parts.append(f"{k}={v_str}")
    return ", ".join(parts)


def format_event(line: str) -> str | None:
    try:
        event = json.loads(line)
    except (json.JSONDecodeError, ValueError):
        stripped = line.strip()
        return stripped if stripped else None

    event_type = event.get("type")

    if event_type == "system" and event.get("subtype") == "init":
        return None

    if event_type == "assistant":
        message = event.get("message", {})
        content = message.get("content", [])
        parts = []
        for item in content:
            if item.get("type") == "text":
                text = item.get("text", "").strip()
                if text:
                    parts.append(f"\n{WHITE}{text}{RESET}\n")
            elif item.get("type") == "tool_use":
                tool_name = item.get("name", "unknown")
                tool_input = item.get("input", {})
                summary = _format_tool_call(tool_name, tool_input)
                parts.append(f"  {CYAN}→ {tool_name}{RESET}({summary})")
        return "\n".join(parts) if parts else None

    if event_type == "user":
        content = event.get("message", {}).get("content", [])
        for item in content:
            if item.get("type") == "tool_result":
                result = str(item.get("content", ""))
                lines = result.split("\n")[:2]
                lines = [re.sub(r"^\s*\d+→", "", ln) for ln in lines]
                preview = " | ".join(ln.strip() for ln in lines if ln.strip())[:300]
                if preview:
                    suffix = "..." if len(result) > 150 else ""
                    is_error = item.get("is_error", False)
                    color = RED if is_error else DIM
                    return f"  {color}← {preview}{suffix}{RESET}"
        return None

    if event_type == "result":
        return f"\n{GREEN}✓ Done{RESET}\n"

    return None


def run_claude_stream(args: list[str], prefix: str = "") -> int:
    """Run claude-safe with stream-json output. Returns exit code."""
    proc = subprocess.Popen(
        CLAUDE_SAFE_CMD
        + ["--verbose", "--output-format", "stream-json"]
        + args,
        stdout=subprocess.PIPE,
        text=True,
        start_new_session=True,
    )
    assert proc.stdout is not None
    for line in proc.stdout:
        formatted = format_event(line)
        if formatted:
            if prefix:
                formatted = f"{DIM}[{prefix}]{RESET} {formatted}"
            with _output_lock:
                print(formatted, file=sys.stderr, flush=True)
    proc.wait()
    return proc.returncode


def run_claude_interactive(args: list[str]) -> int:
    """Run claude-safe interactively. Returns exit code."""
    result = subprocess.run(CLAUDE_SAFE_CMD + args, start_new_session=True)
    return result.returncode


# ---------------------------------------------------------------------------
# claude-server invocation (remote, headless)
# ---------------------------------------------------------------------------
_SESSION_RE = re.compile(r"^==>\s*Session:\s*(\S+)")


def run_claude_server(
    prompt: str,
    model: str,
    runner_args: list[str],
    prefix: str = "",
    system_prompt: str = "",
) -> tuple[int, str | None]:
    """Run `claude-server -p <prompt> --follow`. Live-tails docker logs until
    the remote container exits. Returns (exit_code, session_id_or_None)."""
    cmd = [
        "claude-server",
        *runner_args,
        "-p",
        prompt,
        "--model",
        model,
        "--follow",
    ]
    if system_prompt:
        # claude-server passes everything after `--` straight to claude.
        cmd += ["--", "--append-system-prompt", system_prompt]
    proc = subprocess.Popen(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        start_new_session=True,
    )
    assert proc.stdout is not None
    session_id: str | None = None
    for line in proc.stdout:
        stripped = line.rstrip()
        if session_id is None:
            m = _SESSION_RE.search(stripped)
            if m:
                session_id = m.group(1)
        out = stripped
        if prefix:
            out = f"{DIM}[{prefix}]{RESET} {out}"
        with _output_lock:
            print(out, file=sys.stderr, flush=True)
    proc.wait()
    return proc.returncode, session_id


def sync_claude_server(session_id: str) -> int:
    """Sync the remote workspace back to local via `claude-server --sync`."""
    result = subprocess.run(
        ["claude-server", "--sync", session_id], start_new_session=True
    )
    return result.returncode


def run_shell(command: str) -> int:
    """Run a shell command on the host. Returns exit code."""
    result = subprocess.run(["bash", "-c", command], start_new_session=True)
    return result.returncode


# ---------------------------------------------------------------------------
# Session context (injected as --append-system-prompt before every prompt step)
# ---------------------------------------------------------------------------
GET_METADATA_SH = Path(".claude/helpers/get_metadata.sh")
NOTIFY_SH = Path(".claude/helpers/notify.sh")


def notify(message: str, title: str = "pipeline") -> None:
    """Best-effort notification via .claude/helpers/notify.sh. Silent if missing."""
    if not NOTIFY_SH.exists():
        return
    try:
        subprocess.run(
            [str(NOTIFY_SH), message, title],
            timeout=10,
            check=False,
        )
    except (subprocess.TimeoutExpired, FileNotFoundError):
        pass


def _format_duration(seconds: float) -> str:
    s = int(seconds)
    if s < 60:
        return f"{s}s"
    m, s = divmod(s, 60)
    if m < 60:
        return f"{m}m{s:02d}s"
    h, m = divmod(m, 60)
    return f"{h}h{m:02d}m"


def _gh_owner_repo() -> str:
    """Best-effort `owner/repo` from `origin` URL. Empty if unavailable."""
    try:
        out = subprocess.run(
            ["git", "remote", "get-url", "origin"],
            capture_output=True,
            text=True,
            timeout=5,
        )
    except (subprocess.TimeoutExpired, FileNotFoundError):
        return ""
    if out.returncode != 0:
        return ""
    url = out.stdout.strip()
    if url.endswith(".git"):
        url = url[:-4]
    m = re.search(r"[:/]([^/:]+/[^/]+)$", url)
    return m.group(1) if m else ""


def collect_session_context() -> str:
    """Produce the system-prompt context appended to every Claude session.

    Shells out to `.claude/helpers/get_metadata.sh` (single source of truth
    for repo/branch/commit/timestamp) and tacks on the GitHub `owner/repo`
    derived from `origin` so prompts can reference it without `gh repo view`.
    """
    body = ""
    if GET_METADATA_SH.exists():
        try:
            out = subprocess.run(
                ["bash", str(GET_METADATA_SH)],
                capture_output=True,
                text=True,
                timeout=10,
            )
            if out.returncode == 0:
                body = out.stdout.rstrip()
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass

    owner_repo = _gh_owner_repo()

    lines = [
        "## Session context",
        "",
        "The pipeline runner injected this context. Use it instead of",
        "rederiving (no need to run `date`, `gh repo view`, or `git`).",
        "",
    ]
    if body:
        lines.append(body)
    if owner_repo:
        lines.append(f"GitHub owner/repo: {owner_repo}")
    return "\n".join(lines)


def header(title: str) -> None:
    print(f"\n{BLUE}{'=' * 60}{RESET}", file=sys.stderr, flush=True)
    print(f"{BOLD}  {title}{RESET}", file=sys.stderr, flush=True)
    print(f"{BLUE}{'=' * 60}{RESET}\n", file=sys.stderr, flush=True)


def step_banner(pipeline_name: str, step_id: str, mode: str, target: str) -> None:
    print(
        f"\n{MAGENTA}[{pipeline_name}]{RESET} step "
        f"{BOLD}{step_id}{RESET} ({mode}) → {target}",
        file=sys.stderr,
        flush=True,
    )


# ---------------------------------------------------------------------------
# Pipeline definition + template rendering
# ---------------------------------------------------------------------------
@dataclass
class Step:
    id: str
    interactive: bool
    prompt: str | None = None
    command: str | None = None
    output: str | None = None


@dataclass
class Pipeline:
    name: str
    steps: list[Step]
    description: str | None = None
    runner: str = "safe"
    runner_args: list[str] = field(default_factory=list)


_TEMPLATE_RE = re.compile(r"\{\{\s*([\w.]+)\s*\}\}")


def render(template: str, ctx: dict, extra: dict | None = None) -> str:
    extra = extra or {}

    def resolve(match: re.Match[str]) -> str:
        path = match.group(1)
        if path in extra:
            return extra[path]
        if path == "input":
            return ctx["input"]
        if path == "timestamp":
            return ctx["timestamp"]
        if path == "pipeline_dir":
            return ctx["pipeline_dir"]
        if path.startswith("steps."):
            parts = path.split(".")
            if len(parts) != 3:
                raise ValueError(
                    f"Bad reference `{path}` — expected `steps.<id>.<field>`"
                )
            _, sid, field = parts
            result = ctx["steps"].get(sid)
            if result is None:
                raise ValueError(f"No prior step `{sid}`")
            value = result.get(field)
            if value is None:
                raise ValueError(f"Step `{sid}` has no field `{field}`")
            return value
        raise ValueError(f"Unknown variable `{path}`")

    return _TEMPLATE_RE.sub(resolve, template)


def load_pipeline(path: Path) -> Pipeline:
    data = yaml.safe_load(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise SystemExit(f"{path}: top-level must be a mapping")
    if "name" not in data or "steps" not in data:
        raise SystemExit(f"{path}: missing required keys `name` and `steps`")

    runner = data.get("runner", "safe")
    if runner not in {"safe", "server"}:
        raise SystemExit(
            f"{path}: `runner` must be `safe` or `server` (got {runner!r})"
        )

    runner_args = data.get("runner_args", [])
    if not isinstance(runner_args, list) or not all(
        isinstance(a, str) for a in runner_args
    ):
        raise SystemExit(f"{path}: `runner_args` must be a list of strings")

    steps = []
    for raw in data["steps"]:
        sid = raw.get("id", "?")
        if "interactive" not in raw:
            raise SystemExit(
                f"{path}: step `{sid}` must set `interactive: true|false`"
            )
        has_prompt = "prompt" in raw
        has_command = "command" in raw
        if has_prompt == has_command:
            raise SystemExit(
                f"{path}: step `{sid}` must set exactly one of "
                "`prompt:` or `command:`"
            )
        steps.append(
            Step(
                id=raw["id"],
                interactive=bool(raw["interactive"]),
                prompt=raw.get("prompt"),
                command=raw.get("command"),
                output=raw.get("output"),
            )
        )
    return Pipeline(
        name=data["name"],
        description=data.get("description"),
        steps=steps,
        runner=runner,
        runner_args=runner_args,
    )


def run_pipeline(
    pipeline: Pipeline, pipeline_path: Path, user_input: str, model: str, runner: str
) -> None:
    ctx: dict = {
        "input": user_input,
        "timestamp": datetime.now().strftime("%Y-%m-%dT%H-%M-%S"),
        "pipeline_dir": str(pipeline_path.parent.resolve()),
        "steps": {},
    }

    header(pipeline.name)
    started = datetime.now()

    for step in pipeline.steps:
        output = render(step.output, ctx) if step.output else ""

        if step.command is not None:
            command = render(step.command, ctx, extra={"output": output})
            target = output or "(no output file)"
            step_banner(pipeline.name, step.id, "shell", target)
            code = run_shell(command)
        else:
            assert step.prompt is not None
            prompt = render(step.prompt, ctx, extra={"output": output})
            # Recompute per step — branch/commit change between setup and build.
            session_ctx = collect_session_context()
            if runner == "server":
                target = output or "(no output file)"
                step_banner(pipeline.name, step.id, "server", target)
                code, session_id = run_claude_server(
                    prompt,
                    model,
                    pipeline.runner_args,
                    prefix=step.id,
                    system_prompt=session_ctx,
                )
                if code == 0 and session_id:
                    sync_code = sync_claude_server(session_id)
                    if sync_code != 0:
                        raise SystemExit(
                            f"{RED}Step `{step.id}`: sync failed "
                            f"(exit {sync_code}){RESET}"
                        )
            else:
                mode = "interactive" if step.interactive else "print"
                target = output or "(no output file)"
                step_banner(pipeline.name, step.id, mode, target)
                claude_args = [
                    "--model",
                    model,
                    "--append-system-prompt",
                    session_ctx,
                ]
                if step.interactive:
                    code = run_claude_interactive(claude_args + [prompt])
                else:
                    code = run_claude_stream(claude_args + ["-p", prompt])

        if code != 0:
            duration = _format_duration((datetime.now() - started).total_seconds())
            notify(
                f"[{pipeline.name}] failed at step `{step.id}` "
                f"(exit {code}) after {duration}",
                title=f"pipeline:{pipeline.name}",
            )
            raise SystemExit(
                f"{RED}Step `{step.id}` failed (exit {code}){RESET}"
            )

        ctx["steps"][step.id] = {"output": output}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("pipeline", help="Path to pipeline YAML file")
    parser.add_argument("input", nargs="+", help="Initial user input")
    parser.add_argument(
        "--model",
        default=os.environ.get("CLAUDE_MODEL", "opus"),
        help="Claude model alias or full ID (default: opus)",
    )
    parser.add_argument(
        "--runner",
        choices=["safe", "server"],
        default=None,
        help="Override the pipeline's `runner` (safe|server). "
        "Default: use the value declared in the YAML.",
    )
    args = parser.parse_args()

    user_input = " ".join(args.input).strip()
    if not user_input:
        parser.error("input must not be empty")

    pipeline_path = Path(args.pipeline)
    pipeline = load_pipeline(pipeline_path)
    runner = args.runner or pipeline.runner
    run_pipeline(pipeline, pipeline_path, user_input, args.model, runner)


if __name__ == "__main__":
    main()
