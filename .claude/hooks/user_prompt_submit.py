#!/usr/bin/env python3
"""
User prompt validation hook that blocks prompts containing sensitive data.

Set CLAUDE_HOOKS_DEBUG=1 to enable debug logging.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
from datetime import datetime

# Debug mode for troubleshooting
DEBUG = os.environ.get("CLAUDE_HOOKS_DEBUG", "").lower() in ("1", "true")

# Simple substring patterns to block (case-insensitive)
BLOCKED_SUBSTRINGS = [
    ("AKIA[0-9A-Z]{16}", "AWS Access Key detected"),
    ("ghp_[A-Za-z0-9]{36,}", "GitHub Personal Access Token detected"),
    ("gho_[A-Za-z0-9]{36,}", "GitHub OAuth Token detected"),
    ("github_pat_[A-Za-z0-9_]{20,}", "GitHub PAT detected"),
    ("xoxb-[0-9]{10,}", "Slack Bot Token detected"),
    ("xoxp-[0-9]{10,}", "Slack User Token detected"),
    # OpenAI keys: sk-proj-..., sk-<org>-... (40+ chars after sk-)
    (
        r"sk-(?:proj-|ant-api03-|[a-zA-Z0-9]{20,})[a-zA-Z0-9_-]{20,}",
        "OpenAI/Anthropic API key detected",
    ),
]

# Exact string patterns (not regex)
BLOCKED_STRINGS = [
    ("-----BEGIN RSA PRIVATE KEY-----", "RSA Private Key detected"),
    ("-----BEGIN OPENSSH PRIVATE KEY-----", "SSH Private Key detected"),
    ("-----BEGIN PGP PRIVATE KEY-----", "PGP Private Key detected"),
]


def debug_log(message: str) -> None:
    """Log debug message if debug mode is enabled."""
    if DEBUG:
        print(f"[DEBUG] user_prompt_submit: {message}", file=sys.stderr)


def validate_prompt(prompt: str) -> tuple[bool, str | None]:
    """
    Validate the user prompt for security or policy violations.
    Returns tuple (is_valid, reason).
    """
    for pattern, reason in BLOCKED_SUBSTRINGS:
        if re.search(pattern, prompt):
            debug_log(f"Blocked regex pattern found: {pattern}")
            return False, reason

    prompt_lower = prompt.lower()
    for pattern, reason in BLOCKED_STRINGS:
        if pattern.lower() in prompt_lower:
            debug_log(f"Blocked string found: {pattern}")
            return False, reason

    debug_log("Prompt validation passed")
    return True, None


def main() -> None:
    try:
        # Parse command line arguments
        parser = argparse.ArgumentParser()
        parser.add_argument(
            "--validate", action="store_true", help="Enable prompt validation"
        )
        args = parser.parse_args()

        # Read JSON input from stdin
        input_data = json.loads(sys.stdin.read())

        # Extract prompt
        prompt = input_data.get("prompt", "")
        debug_log(f"Received prompt of length: {len(prompt)}")

        # Validate prompt if requested
        if args.validate:
            is_valid, reason = validate_prompt(prompt)
            if not is_valid:
                # Exit code 2 blocks the prompt with error message
                print(f"Prompt blocked: {reason}", file=sys.stderr)
                sys.exit(2)

        # Inject current date/time as context
        now = datetime.now().astimezone()
        offset = now.strftime("%z")  # e.g. +0200
        tz_name = now.tzname() or ""
        timestamp = (
            f"{now.strftime('%Y-%m-%d %H:%M')} {tz_name} (UTC{offset[:3]}:{offset[3:]})"
        )
        result = {
            "hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": timestamp,
            }
        }
        print(json.dumps(result))
        sys.exit(0)

    except json.JSONDecodeError as e:
        debug_log(f"JSON decode error: {e}")
        sys.exit(0)
    except Exception as e:
        debug_log(f"Unexpected error: {e}")
        sys.exit(0)


if __name__ == "__main__":
    main()
