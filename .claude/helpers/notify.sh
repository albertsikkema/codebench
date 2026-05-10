#!/usr/bin/env bash
# Generic notification dispatcher.
# Picks a channel from environment variables; falls back to stderr.
#
# Usage:
#   notify.sh "<message>" [title]
#   echo "<message>" | notify.sh - [title]
#
# Channel selection (first configured wins, or set NOTIFY_CHANNEL to force):
#   telegram  TELEGRAM_BOT_TOKEN + TELEGRAM_CHAT_ID
#   slack     SLACK_WEBHOOK_URL
#   ntfy      NTFY_TOPIC (NTFY_SERVER defaults to https://ntfy.sh)
#   desktop   notify-send (Linux) or osascript (macOS)
#   stderr    fallback — always works
#
# Exit codes:
#   0  delivered (or stderr fallback used)
#   2  no message provided

set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: notify.sh '<message>' [title]   |   echo '<msg>' | notify.sh - [title]" >&2
  exit 2
fi

if [ "$1" = "-" ]; then
  MESSAGE=$(cat)
else
  MESSAGE="$1"
fi
TITLE="${2:-Notification}"

if [ -z "$MESSAGE" ]; then
  echo "notify.sh: empty message" >&2
  exit 2
fi

pick_channel() {
  if [ -n "${NOTIFY_CHANNEL:-}" ]; then
    echo "$NOTIFY_CHANNEL"
    return
  fi
  if [ -n "${TELEGRAM_BOT_TOKEN:-}" ] && [ -n "${TELEGRAM_CHAT_ID:-}" ]; then
    echo "telegram"; return
  fi
  if [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
    echo "slack"; return
  fi
  if [ -n "${NTFY_TOPIC:-}" ]; then
    echo "ntfy"; return
  fi
  if command -v notify-send >/dev/null 2>&1 || command -v osascript >/dev/null 2>&1; then
    echo "desktop"; return
  fi
  echo "stderr"
}

CHANNEL=$(pick_channel)

send_telegram() {
  curl -sf "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
    -d chat_id="${TELEGRAM_CHAT_ID}" \
    -d text="${MESSAGE}" \
    -d parse_mode="Markdown" >/dev/null
}

send_slack() {
  payload=$(printf '{"text": %s}' "$(printf '%s' "$MESSAGE" | jq -Rs .)")
  curl -sf -X POST "${SLACK_WEBHOOK_URL}" \
    -H 'Content-type: application/json' \
    -d "$payload" >/dev/null
}

send_ntfy() {
  server="${NTFY_SERVER:-https://ntfy.sh}"
  curl -sf -H "Title: ${TITLE}" -d "${MESSAGE}" "${server%/}/${NTFY_TOPIC}" >/dev/null
}

send_desktop() {
  if command -v notify-send >/dev/null 2>&1; then
    notify-send "${TITLE}" "${MESSAGE}"
  elif command -v osascript >/dev/null 2>&1; then
    osascript -e "display notification \"${MESSAGE//\"/\\\"}\" with title \"${TITLE//\"/\\\"}\""
  else
    return 1
  fi
}

case "$CHANNEL" in
  telegram) send_telegram ;;
  slack)    send_slack ;;
  ntfy)     send_ntfy ;;
  desktop)  send_desktop ;;
  stderr)   ;;  # handled by fallthrough log
  *)
    echo "notify.sh: unknown channel '$CHANNEL'" >&2
    exit 1
    ;;
esac

echo "[notify:${CHANNEL}] ${TITLE}: ${MESSAGE}" >&2
