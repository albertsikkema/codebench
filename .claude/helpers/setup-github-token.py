#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = [
#     "colorama>=0.4.6",
# ]
# ///
"""
GitHub Token Setup for Claude Code

Creates and configures a fine-grained GitHub PAT scoped to a single repo
with minimal permissions, then configures gh CLI to use it.

Usage:
    uv run .claude/helpers/setup-github-token.py                # Interactive setup
    uv run .claude/helpers/setup-github-token.py <token>        # Non-interactive
"""

from __future__ import annotations

import argparse
import getpass
import json
import os
import re
import subprocess
import sys
import webbrowser
from datetime import date

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


def check_gh_cli() -> None:
    """Verify gh CLI is installed."""
    try:
        run(["gh", "--version"])
    except FileNotFoundError:
        print_err("Error: gh CLI is not installed")
        print("  Install from: https://cli.github.com/")
        sys.exit(1)


def detect_repo() -> tuple[str, str, str]:
    """Detect owner/repo from git remote. Returns (full, owner, name)."""
    result = run(["git", "remote", "get-url", "origin"])
    if result.returncode != 0:
        return "", "", ""

    url = result.stdout.strip()
    # Handle SSH (git@github.com:owner/repo.git) and HTTPS
    match = re.search(r"github\.com[:/](.+?)(?:\.git)?$", url)
    if not match:
        return "", "", ""

    full = match.group(1)
    parts = full.split("/")
    if len(parts) == 2:
        return full, parts[0], parts[1]
    return full, "", ""


def prompt_yes_no(question: str, default_yes: bool = True) -> bool:
    """Ask a yes/no question. Returns True for yes."""
    suffix = "[Y/n]" if default_yes else "[y/N]"
    try:
        answer = input(f"{question} {suffix} ").strip().lower()
    except (EOFError, KeyboardInterrupt):
        print()
        return default_yes
    if not answer:
        return default_yes
    return answer.startswith("y")


def guide_token_creation(repo: str, repo_name: str) -> str:
    """Guide user through token creation and return the token."""
    token_name = f"claude-code-{repo_name or 'your-repo'}"
    description = (
        f"Claude Code token for {repo or 'repo'}. "
        "Scoped to this repo only with Actions (RW), Contents (RW), Pull requests (RW), and Workflows (RW)."
    )

    print("A fine-grained Personal Access Token (PAT) is needed with")
    print("minimal permissions scoped to this repository only.")
    print()
    print_info("Step 1: Create a token")
    print()
    print("  Configure:")
    print(f"    Token name:          {token_name}")
    print(f"    Description:         {description}")
    print("    Expiration:          90 days (or your preference)")
    print("    Repository access:   Only select repositories")
    if repo:
        print(f"    Selected repository: {repo}")
    else:
        print("    Selected repository: <select your repo>")
    print()
    print("  Permissions (Repository permissions only):")
    print("    Actions:             Read and Write")
    print("    Contents:            Read and Write")
    print("    Pull requests:       Read and Write")
    print("    Workflows:           Read and Write")
    print("    Metadata:            Read (auto-selected)")
    print()
    print("  Leave all other permissions at 'No access'.")
    print()
    print_info("  Copyable description:")
    print(f"  {description}")
    print()
    print_info("Step 2: Copy the token and paste it below")
    print()

    # Open browser automatically
    token_url = "https://github.com/settings/personal-access-tokens/new"
    try:
        webbrowser.open(token_url)
        print_ok("  Opened GitHub in your browser.")
    except Exception:
        print(f"  Open manually: {token_url}")
    print()

    try:
        token = getpass.getpass("Paste your token: ")
    except (EOFError, KeyboardInterrupt):
        print()
        print_err("Aborted.")
        sys.exit(1)

    if not token.strip():
        print_err("Error: no token provided")
        sys.exit(1)

    return token.strip()


def validate_token(token: str) -> None:
    """Check the token is valid against GitHub API."""
    result = run(
        [
            "curl",
            "-s",
            "-o",
            "/dev/null",
            "-w",
            "%{http_code}",
            "-H",
            f"Authorization: Bearer {token}",
            "-H",
            "Accept: application/vnd.github+json",
            "https://api.github.com/user",
        ]
    )

    code = result.stdout.strip()
    if code != "200":
        print_err(f"Error: token is not valid (HTTP {code})")
        print("  Make sure you copied the full token.")
        sys.exit(1)

    print_ok("Token is valid")


def check_token_format(token: str) -> None:
    """Warn if using a classic PAT instead of fine-grained."""
    if token.startswith("ghp_"):
        print_warn(
            "Warning: this looks like a classic PAT (ghp_*), not a fine-grained token (github_pat_*)"
        )
        print("  Classic tokens cannot be scoped to a single repository.")
        print("  Consider creating a fine-grained token instead:")
        print("  https://github.com/settings/personal-access-tokens/new")
        print()
        if not prompt_yes_no("Continue anyway?", default_yes=False):
            sys.exit(1)


def check_repo_permissions(token: str, repo: str) -> None:
    """Check what permissions the token has on the repo."""
    result = run(
        [
            "curl",
            "-s",
            "-H",
            f"Authorization: Bearer {token}",
            "-H",
            "Accept: application/vnd.github+json",
            f"https://api.github.com/repos/{repo}",
        ]
    )

    if result.returncode != 0:
        return

    try:
        data = json.loads(result.stdout)
    except json.JSONDecodeError:
        return

    if "permissions" not in data:
        # 404 returns {"message": "Not Found"} — token can't see this repo
        if data.get("message") == "Not Found":
            print_warn(f"Warning: token cannot access {repo}")
            print("  Make sure the token has this repository selected.")
        return

    perms = data["permissions"]
    if perms.get("push"):
        print_ok(f"Token has write access to {repo}")
    elif perms.get("pull"):
        print_warn(f"Warning: token has read-only access to {repo}")
        print("  Claude needs 'Contents: Read and Write' to push branches.")
    else:
        print_warn(f"Warning: could not determine token permissions on {repo}")


def get_token_expiry(token: str) -> str:
    """Fetch token expiry date from GitHub API response headers."""
    result = run(
        [
            "curl",
            "-s",
            "-i",
            "-H",
            f"Authorization: Bearer {token}",
            "-H",
            "Accept: application/vnd.github+json",
            "https://api.github.com/user",
        ]
    )

    if result.returncode == 0:
        headers = (
            result.stdout.split("\r\n\r\n")[0] if "\r\n\r\n" in result.stdout else ""
        )
        for line in headers.split("\r\n"):
            if line.lower().startswith("github-authentication-token-expiration:"):
                return line.split(":", 1)[1].strip()

    return ""


def get_existing_env_token() -> dict[str, str]:
    """Read existing GH_TOKEN metadata from .env file. Returns dict with 'token', 'name', 'expires'."""
    env_path = os.path.join(os.getcwd(), ".env")
    result: dict[str, str] = {"token": "", "name": "", "expires": ""}
    if not os.path.exists(env_path):
        return result
    with open(env_path) as f:
        for line in f:
            m = re.match(r"^#\s*name:\s*([^,\n]+?)(?:,\s*expires:\s*(.+))?$", line)
            if m:
                result["name"] = m.group(1).strip()
                if m.group(2):
                    result["expires"] = m.group(2).strip()
            m = re.match(r"^GH_TOKEN=(.+)$", line)
            if m:
                result["token"] = m.group(1).strip()
    return result


def check_existing_token() -> dict[str, str] | None:
    """Check for existing token in .env. If found, offer renewal flow. Returns existing metadata or None."""
    existing = get_existing_env_token()
    if not existing["token"]:
        return None

    print_info("Existing token found in .env")
    if existing["name"]:
        print(f"  Name:    {existing['name']}")
    if existing["expires"]:
        print(f"  Expires: {existing['expires']}")
    print()

    if not prompt_yes_no("Renew this token?"):
        print()
        print_warn("Continuing with new token setup instead.")
        print()
        return None

    # Open browser to the fine-grained tokens page
    tokens_url = "https://github.com/settings/personal-access-tokens"
    print()
    if existing["name"]:
        print(f"  Find your token '{existing['name']}' and click 'Regenerate token'.")
    else:
        print("  Find your token and click 'Regenerate token'.")
    print()
    try:
        webbrowser.open(tokens_url)
        print_ok("  Opened GitHub token settings in your browser.")
    except Exception:
        print(f"  Open manually: {tokens_url}")
    print()

    try:
        token = getpass.getpass("Paste your regenerated token: ")
    except (EOFError, KeyboardInterrupt):
        print()
        print_err("Aborted.")
        sys.exit(1)

    if not token.strip():
        print_err("Error: no token provided")
        sys.exit(1)

    existing["token"] = token.strip()
    return existing


def save_token_to_env(token: str, token_name: str, token_expires: str) -> None:
    """Save GH_TOKEN to .env file in the repo root.

    Three cases:
      1. real GH_TOKEN= already present -> replace it (and our comment lines above).
      2. only a commented placeholder (# GH_TOKEN=...) is present -> replace that line.
      3. nothing about GH_TOKEN -> append, or create the file.
    """
    env_path = os.path.join(os.getcwd(), ".env")
    today = date.today().isoformat()
    comment_why = "# GH_TOKEN - used by Claude Code for container-based development (repo scope only)"
    comment_added = f"# added by setup-github-token.py on {today}"
    meta_parts = []
    if token_name:
        meta_parts.append(f"name: {token_name}")
    if token_expires:
        meta_parts.append(f"expires: {token_expires}")
    comment_meta = f"# {', '.join(meta_parts)}" if meta_parts else ""
    entry = f"GH_TOKEN={token}"

    lines = [comment_why, comment_added]
    if comment_meta:
        lines.append(comment_meta)
    lines.append(entry)
    block = "\n".join(lines)

    if os.path.exists(env_path):
        with open(env_path, "r") as f:
            content = f.read()

        if re.search(r"^GH_TOKEN=", content, re.MULTILINE):
            content = re.sub(
                r"(^# GH_TOKEN[^\n]*\n)?(^# added by setup-github-token\.py[^\n]*\n)?(^# name:[^\n]*\n)?^GH_TOKEN=[^\n]*",
                block,
                content,
                count=1,
                flags=re.MULTILINE,
            )
            print_ok("Updated GH_TOKEN in .env")
        elif re.search(r"^[ \t]*#[ \t]*GH_TOKEN=", content, re.MULTILINE):
            content = re.sub(
                r"^[ \t]*#[ \t]*GH_TOKEN=[^\n]*",
                block,
                content,
                count=1,
                flags=re.MULTILINE,
            )
            print_ok("Replaced GH_TOKEN placeholder in .env")
        else:
            if not content.endswith("\n"):
                content += "\n"
            content += f"\n{block}\n"
            print_ok("Added GH_TOKEN to .env")

        with open(env_path, "w") as f:
            f.write(content)
    else:
        with open(env_path, "w") as f:
            f.write(f"{block}\n")
        print_ok("Created .env with GH_TOKEN")


def configure_gh(token: str) -> None:
    """Configure gh CLI with the token.

    In containers: runs 'gh auth login' since there's no existing session.
    On local machines: skips 'gh auth login' to avoid replacing the user's
    personal GitHub auth. The token is available via GH_TOKEN in .env.
    """
    in_container = (
        os.path.exists("/.dockerenv") or os.environ.get("CLAUDE_CONTAINER_MODE") == "1"
    )

    if not in_container:
        print_ok(
            "Skipping gh auth login (local machine — preserving your existing gh session)"
        )
        print("  Claude Code will use GH_TOKEN from .env instead.")
        return

    cmd = ["gh", "auth", "login", "--with-token", "--insecure-storage"]
    print_warn("Container detected: using file-based token storage")

    result = subprocess.run(cmd, input=token, capture_output=True, text=True)
    if result.returncode != 0:
        print_err("Error: gh auth configuration failed")
        if result.stderr:
            print(f"  {result.stderr.strip()}")
        sys.exit(1)

    # Verify
    verify = run(["gh", "auth", "status"])
    if verify.returncode != 0:
        print_err("Error: gh auth verification failed")
        sys.exit(1)

    print_ok("gh CLI configured successfully")

    # Set up git credential helper
    run(["gh", "auth", "setup-git"])


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Set up a scoped GitHub token for Claude Code"
    )
    parser.add_argument(
        "token",
        nargs="?",
        default=None,
        help="GitHub token (if not provided, interactive setup guides you)",
    )
    args = parser.parse_args()

    print_header("GitHub Token Setup for Claude Code")

    check_gh_cli()

    repo, owner, repo_name = detect_repo()

    if not repo:
        print_warn("Warning: could not detect repository from git remote")
        print("  Run this script from inside your project directory.")
        print()
    else:
        print_info(f"Detected repository: {repo}")
        print()

    # Check for existing token (renewal flow)
    renewal = None
    if not args.token:
        renewal = check_existing_token()

    if renewal:
        # Renewal flow: token and name come from existing .env
        token = renewal["token"]
        token_name = renewal["name"]
    else:
        # New token flow
        token = args.token
        if not token:
            token = guide_token_creation(repo, repo_name)
            token_name = f"claude-code-{repo_name or 'your-repo'}"
        else:
            # Argument mode: check .env for existing name, ask only if missing
            existing = get_existing_env_token()
            token_name = existing["name"]
            if not token_name:
                try:
                    token_name = input("Token name (as entered on GitHub): ").strip()
                except (EOFError, KeyboardInterrupt):
                    print()
                    token_name = ""

    print()
    print_info("Validating token...")

    check_token_format(token)
    validate_token(token)

    if repo:
        check_repo_permissions(token, repo)

    # Fetch token expiry from API
    print()
    print_info("Fetching token metadata...")
    token_expires = get_token_expiry(token)

    print()
    print_info("Configuring gh CLI...")

    configure_gh(token)

    # Save token to .env file
    save_token_to_env(token, token_name, token_expires)

    print_header("Setup Complete")
    print_ok("Token saved to .env (GH_TOKEN).")
    print()
    print("  Claude can now:")
    print_ok("    - Read repository contents")
    print_ok("    - Create and push branches")
    print_ok("    - Create and update pull requests")
    print_ok("    - Manage GitHub Actions workflows")
    print_ok("    - Push changes to .github/workflows/")
    print()
    if repo:
        print("  Next step: protect the main branch from direct pushes:")
        print()
        print(f"    uv run {os.path.dirname(__file__)}/setup-branch-protection.py")
        print()
    print()


if __name__ == "__main__":
    main()
