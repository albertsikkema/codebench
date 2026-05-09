#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = []
# ///
"""Pre-build local setup: create feature branch, push, open PR.

Adapted from project-server-main worker (build.py:78-191), but standalone:
no project-server, no API client. Run before the server-side build/review/
triage steps so those see an existing branch + PR to work on.

Outputs the PR URL to `--out-pr-url <path>` for the next pipeline step.
"""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from pathlib import Path


def log(msg: str) -> None:
    print(f"[build_setup] {msg}", file=sys.stderr, flush=True)


def slug_from_plan_path(plan_path: Path) -> str:
    """Drop a leading `YYYY-MM-DD-` date prefix and the `.md` suffix."""
    stem = plan_path.stem
    stem = re.sub(r"^\d{4}-\d{2}-\d{2}-", "", stem)
    slug = re.sub(r"[^a-z0-9-]+", "-", stem.lower()).strip("-")
    return slug or "plan"


def existing_branches() -> set[str]:
    """Union of local and remote (origin) branch names."""
    names: set[str] = set()
    local = subprocess.run(
        ["git", "for-each-ref", "--format=%(refname:short)", "refs/heads/"],
        capture_output=True,
        text=True,
        timeout=15,
    )
    if local.returncode == 0:
        names.update(line.strip() for line in local.stdout.splitlines() if line.strip())
    remote = subprocess.run(
        ["git", "ls-remote", "--heads", "origin"],
        capture_output=True,
        text=True,
        timeout=30,
    )
    if remote.returncode == 0:
        for line in remote.stdout.splitlines():
            if "refs/heads/" in line:
                names.add(line.split("refs/heads/", 1)[1].strip())
    return names


def unique_branch(base: str) -> str:
    taken = existing_branches()
    if base not in taken:
        return base
    n = 2
    while f"{base}-{n}" in taken:
        n += 1
    return f"{base}-{n}"


def first_heading(plan_text: str) -> str | None:
    for line in plan_text.splitlines():
        if line.startswith("# ") and not line.startswith("# ---"):
            return line[2:].strip()
    return None


def run_git(*args: str, timeout: int = 60) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["git", *args], capture_output=True, text=True, timeout=timeout
    )


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--plan", required=True, help="Path to the approved plan markdown")
    ap.add_argument(
        "--out-pr-url", required=True, help="File to write the resulting PR URL into"
    )
    ap.add_argument("--base", default="main", help="Base branch (default: main)")
    args = ap.parse_args()

    plan_path = Path(args.plan).resolve()
    if not plan_path.is_file():
        log(f"FATAL: plan not found: {plan_path}")
        return 1
    plan_text = plan_path.read_text(encoding="utf-8")
    if not plan_text.strip():
        log(f"FATAL: plan is empty: {plan_path}")
        return 1

    # Refuse to start on a non-base branch — caller should be on `main` (or
    # whatever --base is) so the new feature branch forks cleanly.
    head = run_git("symbolic-ref", "--quiet", "--short", "HEAD")
    current = head.stdout.strip()
    if current != args.base:
        log(
            f"FATAL: current branch is `{current}`, expected `{args.base}`. "
            "Switch to the base branch before running the pipeline."
        )
        return 1

    slug = slug_from_plan_path(plan_path)
    base_branch = f"feat/{slug}"
    branch = unique_branch(base_branch)
    if branch != base_branch:
        log(f"`{base_branch}` already exists, using `{branch}`")

    log(f"Creating branch {branch}")
    co = run_git("checkout", "-b", branch)
    if co.returncode != 0:
        log(f"FATAL: checkout -b failed: {co.stderr.strip()}")
        return 1

    log("Empty commit to diverge from base")
    commit = run_git(
        "commit", "--allow-empty", "-m", f"chore({slug}): start build", timeout=30
    )
    if commit.returncode != 0:
        log(f"FATAL: empty commit failed: {commit.stderr.strip()}")
        return 1

    log("Pushing to origin")
    push = run_git("push", "-u", "origin", branch, timeout=120)
    if push.returncode != 0:
        log(f"FATAL: push failed: {push.stderr.strip()}")
        return 1

    heading = first_heading(plan_text)
    title = f"feat({slug}): {heading}" if heading else f"feat({slug}): build"

    log(f"Creating PR: {title}")
    pr = subprocess.run(
        [
            "gh",
            "pr",
            "create",
            "--base",
            args.base,
            "--head",
            branch,
            "--title",
            title,
            "--body-file",
            str(plan_path),
        ],
        capture_output=True,
        text=True,
        timeout=60,
    )
    if pr.returncode != 0:
        log(f"FATAL: gh pr create failed: {pr.stderr.strip()}")
        return 1
    pr_url = pr.stdout.strip().splitlines()[-1].strip()
    if not pr_url.startswith("http"):
        log(f"FATAL: unexpected gh output: {pr.stdout!r}")
        return 1

    out = Path(args.out_pr_url)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(pr_url + "\n", encoding="utf-8")
    log(f"PR ready: {pr_url}")
    log(f"Wrote PR URL to {out}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
