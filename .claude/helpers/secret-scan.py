#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.10"
# dependencies = []
# ///
"""Scan for hardcoded secrets. Uses gitleaks if available, regex fallback otherwise."""

import json, os, re, subprocess, sys
from pathlib import Path

# Regex patterns for common secret types
PATTERNS = {
    "AWS Access Key": r"AKIA[0-9A-Z]{16}",
    "GitHub Token": r"gh[pousr]_[a-zA-Z0-9]{36,}",
    "Stripe Secret Key": r"sk_(live|test)_[a-zA-Z0-9]{24,}",
    "JWT Token": r"eyJ[a-zA-Z0-9_-]{10,}\.eyJ[a-zA-Z0-9_-]{10,}",
    "Private Key": r"-----BEGIN (RSA|EC|OPENSSH|DSA|PGP) PRIVATE KEY-----",
    "Connection String": r"(postgres|mongodb|mysql|redis)://[^\s\"']+:[^\s\"']+@",
    "Generic Secret Assignment": r"""(?:password|secret|token|api_key|apikey)\s*[:=]\s*['"][^'"]{8,}['"]""",
}

# File extensions to scan
SCAN_EXTENSIONS = {
    ".py", ".js", ".ts", ".jsx", ".tsx", ".go", ".rs", ".java",
    ".rb", ".php", ".cs", ".env", ".yaml", ".yml", ".json",
    ".toml", ".cfg", ".ini", ".conf", ".sh", ".bash",
}

SKIP_DIRS = {
    ".git", "node_modules", "venv", ".venv", "__pycache__",
    "dist", "build", ".next", ".nuxt",
}


def try_gitleaks(target_dir: str) -> list[dict] | None:
    """Run gitleaks if available. Returns findings or None."""
    try:
        result = subprocess.run(
            ["gitleaks", "detect", "--no-git", "-f", "json", "--source", target_dir],
            capture_output=True, text=True, timeout=120,
        )
    except (FileNotFoundError, subprocess.TimeoutExpired):
        return None

    if result.returncode == 0:
        return []  # clean
    try:
        return json.loads(result.stdout) if result.stdout.strip() else []
    except json.JSONDecodeError:
        return None


def regex_scan(target_dir: str) -> list[dict]:
    """Fallback: scan files with regex patterns."""
    findings = []
    root = Path(target_dir)
    for path in root.rglob("*"):
        if any(skip in path.parts for skip in SKIP_DIRS):
            continue
        if path.suffix not in SCAN_EXTENSIONS or not path.is_file():
            continue
        try:
            text = path.read_text(errors="ignore")
        except OSError:
            continue
        for line_num, line in enumerate(text.splitlines(), 1):
            for secret_type, pattern in PATTERNS.items():
                if re.search(pattern, line, re.IGNORECASE):
                    findings.append({
                        "file": str(path.relative_to(root)),
                        "line": line_num,
                        "type": secret_type,
                        "match": line.strip()[:120],
                    })
    return findings


def main() -> int:
    target = sys.argv[1] if len(sys.argv) > 1 else "."
    if not os.path.isdir(target):
        print(f"Error: {target} is not a directory", file=sys.stderr)
        return 1

    # Try gitleaks first
    gitleaks_results = try_gitleaks(target)
    if gitleaks_results is not None:
        print(f"[secret-scan: gitleaks found {len(gitleaks_results)} findings]",
              file=sys.stderr)
        for finding in gitleaks_results:
            print(json.dumps({
                "file": finding.get("File", ""),
                "line": finding.get("StartLine", 0),
                "type": finding.get("RuleID", "unknown"),
                "match": finding.get("Match", "")[:120],
            }))
        return 0 if not gitleaks_results else 1

    # Fallback to regex
    print("[secret-scan: gitleaks not found, using regex fallback]",
          file=sys.stderr)
    findings = regex_scan(target)
    print(f"[secret-scan: regex found {len(findings)} findings]",
          file=sys.stderr)
    for finding in findings:
        print(json.dumps(finding))
    return 0 if not findings else 1


if __name__ == "__main__":
    sys.exit(main())
