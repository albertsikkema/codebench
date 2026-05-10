#!/bin/bash
# Auto-format files after Claude writes/edits them.
# Python (ruff), Go (gofmt), TypeScript/JavaScript (prettier).
# Each formatter is best-effort: missing tools silently skip.

FILE_PATH=$(jq -r '.tool_input.file_path // empty')

[ -z "$FILE_PATH" ] && exit 0
[ ! -f "$FILE_PATH" ] && exit 0

case "$FILE_PATH" in
  *.py)
    command -v ruff >/dev/null && ruff format --quiet "$FILE_PATH" 2>/dev/null
    ;;
  *.go)
    command -v gofmt >/dev/null && gofmt -w "$FILE_PATH" 2>/dev/null
    ;;
  *.ts|*.tsx|*.js|*.jsx|*.mjs|*.cjs)
    command -v prettier >/dev/null && prettier --write --log-level=silent "$FILE_PATH" 2>/dev/null
    ;;
esac

exit 0
