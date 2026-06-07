#!/usr/bin/env bash
# tui.sh — Mission Control terminal dashboard
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MC_DIR="${MC_DIR:-$(dirname "$SCRIPT_DIR")}"
GITHUB_REPO="${GITHUB_REPO:-Kaiukov/mission-control}"
SESSION="mc"

tmux kill-session -t "$SESSION" 2>/dev/null || true

# Write watch script (double-quoted heredoc so env vars get baked in)
WATCH_SCRIPT="/tmp/mc-watch-$$.sh"
cat > "$WATCH_SCRIPT" << WATCHEOF
#!/bin/bash
while true; do
  clear
  echo "🛸 Mission Control — \$(date +%H:%M:%S)"
  echo " Repo: $GITHUB_REPO"
  echo ""
  cat $MC_DIR/state/tasks.json 2>/dev/null | jq -r '.tasks[] | "  \\(.state[:1])  #\\(.number)  \\(.title[:55])"' 2>/dev/null || echo "  (no tasks — run pull-issues.sh)"
  echo ""
  echo "────────────────────────────────────────────"
  echo " r=run  q=quit  ↑↓=switch panes"
  sleep 5
done
WATCHEOF
chmod +x "$WATCH_SCRIPT"

# Start session
tmux new-session -d -s "$SESSION" -x 140 -y 40 "$WATCH_SCRIPT"
tmux rename-window -t "$SESSION" "mc"

# Bottom pane
tmux split-window -v -t "$SESSION" -p 30
sleep 0.3
tmux send-keys -t "$SESSION" 'echo "┌──────────────────────────────────────────┐"' Enter
tmux send-keys -t "$SESSION" 'echo "│  🛸  Mission Control v0.1               │"' Enter
tmux send-keys -t "$SESSION" "echo '│  repo: $GITHUB_REPO'" Enter
tmux send-keys -t "$SESSION" 'echo "│  r=run worker  q=quit                   │"' Enter
tmux send-keys -t "$SESSION" 'echo "└──────────────────────────────────────────┘"' Enter

# Select top pane
tmux select-pane -t "$SESSION" -U

# Bind keys
tmux bind-key q "kill-session"

echo "🛸 Mission Control TUI launched"
echo "   Attach: tmux attach -t $SESSION"
echo "   Repo: $GITHUB_REPO"
