#!/usr/bin/env bash
# tui.sh — Mission Control terminal dashboard
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MC_DIR="${MC_DIR:-$(dirname "$SCRIPT_DIR")}"
[ -f "$MC_DIR/.env" ] && source "$MC_DIR/.env"
GITHUB_REPO="${GITHUB_REPO:-Kaiukov/mission-control}"
SESSION="mc"

tmux kill-session -t "$SESSION" 2>/dev/null || true

# ── Top pane: live task board ──
WATCH_SCRIPT="/tmp/mc-watch-$$.sh"
cat > "$WATCH_SCRIPT" << WATCHEOF
#!/bin/bash
MC_DIR="$MC_DIR"
while true; do
  clear
  NOW=\$(date +%H:%M)
  printf '%s\n' ''
  printf '%s\n' '┌──────────────────────────────────────────────────────────────┐'
  printf '%s\n' "│                    🛸 MISSION CONTROL           \$NOW        │"
  printf '%s\n' '│                    v0.1 · Repo: $GITHUB_REPO'
  printf '%s\n' '├──────────────────────────────────────────────────────────────┤'
  printf '%s\n' ''

  if [ -f "\$MC_DIR/state/tasks.json" ]; then
    COUNT=\$(jq '.tasks | length' "\$MC_DIR/state/tasks.json" 2>/dev/null || echo 0)
    if [ "\$COUNT" -eq 0 ]; then
      printf '%s\n' '  (no tasks — create an issue on GitHub)'
    else
      jq -r '.tasks[] | "  [ ]  #\(.number)  \(.title)"' "\$MC_DIR/state/tasks.json" 2>/dev/null
    fi
  fi

  printf '%s\n' ''
  printf '%s\n' '├──────────────────────────────────────────────────────────────┤'
  printf '%s\n' '│  [r] run    [q] quit    [Ctrl+b ↑↓] switch pane             │'
  printf '%s\n' '└──────────────────────────────────────────────────────────────┘'
  printf '%s\n' ''
  sleep 5
done
WATCHEOF
chmod +x "$WATCH_SCRIPT"

# ── Launch tmux ──
tmux new-session -d -s "$SESSION" -x 140 -y 40 "$WATCH_SCRIPT"
tmux rename-window -t "$SESSION" "mc"

# ── Bottom pane: help ──
tmux split-window -v -t "$SESSION" -p 25
sleep 0.3

# Write help to a temp file, then cat it (clean — no bash prompts)
HELP_FILE="/tmp/mc-help.txt"
cat > "$HELP_FILE" << HELPEOF

  ▸ HOW TO USE

  1.  Look at tasks above ↑
  2.  Press r to spawn Codex worker
  3.  Switch to worker: Ctrl+b ↑↓
  4.  Or attach directly: tmux attach -t task-1

  Repo: $GITHUB_REPO

HELPEOF

tmux send-keys -t "$SESSION" "clear && cat $HELP_FILE" Enter
sleep 0.1

# Select top pane
tmux select-pane -t "$SESSION" -U

# ── Keys ──
tmux bind-key r "run-shell 'bash $MC_DIR/scripts/spawn-worker.sh'"
tmux bind-key q "kill-session"

echo "✅ TUI ready — tmux attach -t $SESSION"
