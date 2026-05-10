#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = [
#     "colorama>=0.4.6",
# ]
# ///
"""
Branch Protection Setup for Claude Code

Configures branch protection rules to prevent direct pushes to main,
requiring all changes to go through pull requests.

Note: Branch protection requires a public repository or a GitHub Pro/Team/Enterprise
plan. Free-tier private repositories do not support branch protection rules.

Usage:
    uv run .claude/helpers/setup-branch-protection.py           # Protect main
    uv run .claude/helpers/setup-branch-protection.py develop   # Protect develop
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys

from colorama import Fore, Style, init

init()


def print_header(title: str) -> None:
    print(f"\n{Fore.BLUE}{'=' * 40}{Style.RESET_ALL}")
    print(f"{Fore.BLUE}  {title}{Style.RESET_ALL}")
    print(f"{Fore.BLUE}{'=' * 40}{Style.RESET_ALL}\n")


def print_ok(msg: str) -> None:
    print(f"{Fore.GREEN}{msg}{Style.RESET_ALL}")


def print_warn(msg: str) -> None:
    print(f"{Fore.YELLOW}{msg}{Style.RESET_ALL}")


def print_err(msg: str) -> None:
    print(f"{Fore.RED}{msg}{Style.RESET_ALL}")


def print_info(msg: str) -> None:
    print(f"{Fore.BLUE}{msg}{Style.RESET_ALL}")


def run(cmd: list[str], **kwargs) -> subprocess.CompletedProcess:
    return subprocess.run(cmd, capture_output=True, text=True, **kwargs)


def check_prerequisites() -> None:
    """Verify gh CLI is installed and authenticated."""
    try:
        run(["gh", "--version"])
    except FileNotFoundError:
        print_err("Error: gh CLI is not installed")
        print("  Install from: https://cli.github.com/")
        sys.exit(1)

    result = run(["gh", "auth", "status"])
    if result.returncode != 0:
        print_err("Error: gh CLI is not authenticated")
        print("  Run: gh auth login")
        print("  Or:  uv run .claude/helpers/setup-github-token.py")
        sys.exit(1)


def detect_repo() -> str:
    """Get owner/repo from gh CLI."""
    result = run(["gh", "repo", "view", "--json", "nameWithOwner", "-q", ".nameWithOwner"])
    if result.returncode != 0 or not result.stdout.strip():
        print_err("Error: could not detect GitHub repository")
        if result.stderr.strip():
            print(f"  gh said: {result.stderr.strip()}")
        print("  Make sure you are inside a git repo with a GitHub remote.")
        sys.exit(1)
    return result.stdout.strip()


def check_admin_access(repo: str) -> None:
    """Verify current user has admin access."""
    result = run(["gh", "api", f"repos/{repo}", "--jq", ".permissions.admin"])
    if result.stdout.strip() != "true":
        print_err("Error: you need admin access to set branch protection rules")
        print(f"  Current user does not have admin permissions on {repo}")
        sys.exit(1)


def check_branch_exists(repo: str, branch: str) -> bool:
    """Check if the branch exists on the remote."""
    result = run(["gh", "api", f"repos/{repo}/branches/{branch}"])
    return result.returncode == 0


def try_rulesets(repo: str, branch: str) -> bool:
    """Try to configure protection using the rulesets API (newer). Returns True on success."""
    ruleset = {
        "name": f"Protect {branch} branch",
        "target": "branch",
        "enforcement": "active",
        "conditions": {
            "ref_name": {
                "include": [f"refs/heads/{branch}"],
                "exclude": [],
            }
        },
        "rules": [
            {
                "type": "pull_request",
                "parameters": {
                    "required_approving_review_count": 0,
                    "dismiss_stale_reviews_on_push": False,
                    "require_code_owner_review": False,
                    "require_last_push_approval": False,
                    "required_review_thread_resolution": False,
                },
            },
            {"type": "non_fast_forward"},
        ],
        "bypass_actors": [],
    }

    ruleset_name = f"Protect {branch} branch"

    # Check for existing ruleset
    result = run(["gh", "api", f"repos/{repo}/rulesets"])
    existing_id = None
    if result.returncode == 0:
        try:
            for rs in json.loads(result.stdout):
                if rs.get("name") == ruleset_name:
                    existing_id = rs["id"]
                    break
        except (json.JSONDecodeError, KeyError):
            pass

    payload = json.dumps(ruleset)

    if existing_id:
        print_warn(f"Updating existing ruleset (ID: {existing_id})...")
        result = run([
            "gh", "api", f"repos/{repo}/rulesets/{existing_id}",
            "--method", "PUT",
            "--input", "-",
        ], input=payload)
    else:
        result = run([
            "gh", "api", f"repos/{repo}/rulesets",
            "--method", "POST",
            "--input", "-",
        ], input=payload)

    if result.returncode == 0:
        action = "updated" if existing_id else "created"
        print_ok(f"Ruleset {action} successfully")
        return True

    return False


def try_legacy_protection(repo: str, branch: str) -> bool:
    """Fall back to legacy branch protection API. Returns True on success."""
    protection = {
        "required_pull_request_reviews": {
            "required_approving_review_count": 0,
            "dismiss_stale_reviews": False,
        },
        "enforce_admins": True,
        "required_status_checks": None,
        "restrictions": None,
    }

    payload = json.dumps(protection)
    result = run([
        "gh", "api", f"repos/{repo}/branches/{branch}/protection",
        "--method", "PUT",
        "--input", "-",
    ], input=payload)

    if result.returncode == 0:
        print_ok("Branch protection configured successfully (legacy API)")
        return True

    return False


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Configure branch protection to prevent direct pushes"
    )
    parser.add_argument(
        "branch",
        nargs="?",
        default="main",
        help="Branch to protect (default: main)",
    )
    args = parser.parse_args()
    branch = args.branch

    print_header("Branch Protection Setup")

    check_prerequisites()

    repo = detect_repo()

    print_info(f"Repository: {repo}")
    print_info(f"Branch:     {branch}")
    print()

    check_admin_access(repo)

    if not check_branch_exists(repo, branch):
        print_warn(f"Warning: branch '{branch}' does not exist yet")
        print("  Protection rules will apply once the branch is created.")
        print()

    print_info("Configuring branch ruleset...")
    print()

    if not try_rulesets(repo, branch):
        print_warn("Rulesets not available, falling back to branch protection API...")
        print()
        if not try_legacy_protection(repo, branch):
            print_err("Failed to configure branch protection")
            print("  You may need a GitHub Pro/Team plan for private repo branch protection.")
            print("  For public repos, ensure you have admin access.")
            sys.exit(1)

    print_header("Protection Rules Applied")
    print_ok("  - Require pull request before merging")
    print_ok("  - No bypass actors (applies to everyone)")
    print_ok("  - Prevent force pushes")
    print()
    print_info(f"Result: All changes to '{branch}' must go through a PR.")
    print_info("Even tokens with write access cannot push directly.")
    print()
    print_warn(f"To review: https://github.com/{repo}/settings/rules")
    print()


if __name__ == "__main__":
    main()
