#!/usr/bin/env bash
# tui.sh — Mission Control terminal dashboard
set -euo pipefail

MC_DIR="${MC_DIR:-/root/mission-control}"
SESSION="mc"

# Kill existing
tmux kill-session -t "$SESSION" 2>/dev/null || true

# Create session
tmux new-session -d -s "$SESSION" -x 140 -y 40

# Rename window
tmux rename-window -t "$SESSION" "🛸 mission-control"

# Top pane (70%): task board
TASKS_CMD="cat $MC_DIR/state/tasks.json 2>/dev/null | jq -r '.tasks[] | \"\(.state[:1]) #\(.number) \(.title[:50])\"' 2>/dev/null || echo 'No tasks yet — run pull-issues.sh first'"
tmux send-keys -t "$SESSION" "watch -n 5 -t '$TASKS_CMD'" C-m
tmux send-keys -t "$SESSION" "clear" C-m

# Bottom pane (30%): terminal
tmux split-window -v -t "$SESSION" -p 30
tmux send-keys -t "$SESSION" "echo '┌─────────────────────────────────────────┐'" C-m
tmux send-keys -t "$SESSION" "echo '│  🛸 Mission Control v0.1               │'" C-m
tmux send-keys -t "$SESSION" "echo '│                                         │'" C-m
tmux send-keys -t "$SESSION" "echo '│  r — run selected task                 │'" C-m
tmux send-keys -t "$SESSION" "echo '│  s — stop worker                       │'" C-m
tmux send-keys -t "$SESSION" "echo '│  Ctrl+b ↑↓ — switch panes              │'" C-m
tmux send-keys -t "$SESSION" "echo '│  q — quit                              │'" C-m
tmux send-keys -t "$SESSION" "echo '└─────────────────────────────────────────┘'" C-m

# Select top pane for navigation
tmux select-pane -t "$SESSION" -U

# Bind keys
tmux bind-key -t "$SESSION" r "run-shell 'bash $MC_DIR/scripts/spawn-worker.sh \$(tmux display-message -p \"#S\")'"
tmux bind-key -t "$SESSION" q "kill-session"

echo "🛸 Mission Control TUI launched"
echo "   Attach: tmux attach -t $SESSION"
echo "   Tasks in top pane, terminal in bottom"
