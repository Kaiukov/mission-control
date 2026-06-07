#!/usr/bin/env bash
# pull-issues.sh — GitHub Issues → tasks.json
set -euo pipefail

MC_DIR="${MC_DIR:-/root/mission-control}"
GITHUB_REPO="${GITHUB_REPO:-Kaiukov/my-portfolio}"
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
