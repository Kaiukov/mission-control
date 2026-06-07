#!/usr/bin/env bash
# pull-issues.sh — GitHub Issues → tasks.json
set -euo pipefail

# Auto-detect MC_DIR from script location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MC_DIR="${MC_DIR:-$(dirname "$SCRIPT_DIR")}"

# Source .env if exists
[ -f "$MC_DIR/.env" ] && source "$MC_DIR/.env"

GITHUB_REPO="${GITHUB_REPO:-Kaiukov/mission-control}"
OUTFILE="$MC_DIR/state/tasks.json"
NOW=$(date -Iseconds)

mkdir -p "$(dirname "$OUTFILE")"

echo '{"updated":"'"$NOW"'","tasks":[' > "$OUTFILE"

FIRST=true
while IFS= read -r issue; do
  if [ "$FIRST" = true ]; then
    FIRST=false
  else
    echo "," >> "$OUTFILE"
  fi
  echo "$issue" >> "$OUTFILE"
done < <(gh issue list \
  --repo "$GITHUB_REPO" \
  --state open \
  --limit 50 \
  --json number,title,labels,state,createdAt,url \
  --jq '.[] | {number,title,labels:[.labels[].name],state,url,created:.createdAt}' 2>/dev/null)

echo ']}' >> "$OUTFILE"

COUNT=$(jq '.tasks | length' "$OUTFILE" 2>/dev/null || echo 0)
echo "✓ Pulled $COUNT issues from $GITHUB_REPO → $OUTFILE"
