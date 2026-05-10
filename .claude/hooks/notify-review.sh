#!/usr/bin/env bash
# Notification hook for dev-team plan reviews.
# Called by the lead session when an item needs human review.
#
# Usage: notify-review.sh "<message>"
#
# Configure your preferred notification channel below.
# The message includes: item ID, title, reason for review, and plan file path.
#
# Examples:
#   Telegram:  curl -s "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
#                -d chat_id="${TELEGRAM_CHAT_ID}" -d text="$1"
#   Slack:     curl -s -X POST "${SLACK_WEBHOOK_URL}" \
#                -H 'Content-type: application/json' -d "{\"text\":\"$1\"}"
#   Email:     echo "$1" | mail -s "Plan Review Needed" "$NOTIFY_EMAIL"
#   ntfy:      curl -s -d "$1" "https://ntfy.sh/${NTFY_TOPIC}"
#   Desktop:   notify-send "Plan Review" "$1"

set -euo pipefail

MESSAGE="${1:?Usage: notify-review.sh '<message>'}"

# --- Uncomment and configure ONE notification method below ---

# -- Telegram --
curl -sf "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
  -d chat_id="${TELEGRAM_CHAT_ID}" \
  -d text="${MESSAGE}" \
  -d parse_mode="Markdown" >/dev/null

# -- Slack webhook --
# curl -sf -X POST "${SLACK_WEBHOOK_URL}" \
#   -H 'Content-type: application/json' \
#   -d "{\"text\":\"${MESSAGE}\"}" >/dev/null

# -- ntfy.sh (simple push notifications) --
# curl -sf -d "${MESSAGE}" "https://ntfy.sh/${NTFY_TOPIC}" >/dev/null

# -- Email --
# echo "${MESSAGE}" | mail -s "Plan Review Needed" "${NOTIFY_EMAIL}"

# -- Desktop notification (Linux) --
# notify-send "Plan Review Needed" "${MESSAGE}"

# -- Fallback: just log it --
echo "[notify-review] ${MESSAGE}" >&2
