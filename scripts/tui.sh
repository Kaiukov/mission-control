#!/usr/bin/env bash
# tui.sh — Mission Control terminal dashboard
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MC_DIR="${MC_DIR:-$(dirname "$SCRIPT_DIR")}"
[ -f "$MC_DIR/.env" ] && source "$MC_DIR/.env"
GITHUB_REPO="${GITHUB_REPO:-Kaiukov/mission-control}"
MODEL="${MODEL:-opencode-go/deepseek-v4-flash}"
SESSION="mc"

tmux kill-session -t "$SESSION" 2>/dev/null || true

# ── Top pane: live task board ──
WATCH_SCRIPT="/tmp/mc-watch-$$.sh"
cat > "$WATCH_SCRIPT" << WATCHEOF
#!/bin/bash
MC_DIR="$MC_DIR"
while true; do
  clear

  # Check running worker
  RUNNING=""
  RUNNING_NUM=""
  if [ -f "\$MC_DIR/state/running.json" ]; then
    RUNNING_NUM=\$(jq -r '.number // ""' "\$MC_DIR/state/running.json" 2>/dev/null)
    RUNNING=" · ⏳ #\${RUNNING_NUM} running"
  fi

  NOW=\$(date +%H:%M)
  printf '%s\n' ''
  printf '%s\n' '┌──────────────────────────────────────────────────────────────┐'
  printf '%s\n' "│                    🛸 MISSION CONTROL           \$NOW        │"
  printf '%s\n' "│                    v0.2 · Repo: $GITHUB_REPO\$RUNNING"
  printf '%s\n' '├──────────────────────────────────────────────────────────────┤'
  printf '%s\n' ''

  if [ -f "\$MC_DIR/state/tasks.json" ]; then
    jq -r --arg rn "\$RUNNING_NUM" '.tasks[] |
      if (.number | tostring) == \$rn then "  [⏳]  #\(.number)  \(.title)"
      else "  [ ]  #\(.number)  \(.title)"
      end' "\$MC_DIR/state/tasks.json" 2>/dev/null
  else
    printf '%s\n' '  (no issues — pull first)'
  fi

  printf '%s\n' ''
  printf '%s\n' '├──────────────────────────────────────────────────────────────┤'
  printf '%s\n' '│  1-9 = run issue    q = quit    ↑↓ = switch pane           │'
  printf '%s\n' '└──────────────────────────────────────────────────────────────┘'
  sleep 5
done
WATCHEOF
chmod +x "$WATCH_SCRIPT"

# ── Launch tmux ──
tmux new-session -d -s "$SESSION" -x 140 -y 40 "$WATCH_SCRIPT"
tmux rename-window -t "$SESSION" "mc"

# ── Green status bar ──
tmux set -t "$SESSION" status-style "fg=black,bg=green"
tmux set -t "$SESSION" status-left " 🛸 MC "
tmux set -t "$SESSION" status-right "#(cat $MC_DIR/state/running.json 2>/dev/null | jq -r 'if .number then \"⏳ worker #\\(.number)\" else \"idle\" end') | %H:%M "

# ── Bottom pane: help + worker output ──
tmux split-window -v -t "$SESSION" -p 25
sleep 0.3

HELP_FILE="/tmp/mc-help.txt"
cat > "$HELP_FILE" << HELPEOF

  ▸ HOW TO USE

  Press 1-9 to run that issue (Codex $MODEL)
  Press r to pick by number
  Worker output appears here ↓
  Press q to quit

  Repo: $GITHUB_REPO

HELPEOF

tmux send-keys -t "$SESSION" "clear && cat $HELP_FILE" Enter
sleep 0.1

# ── Bottom pane ACTIVE by default ──
# (no select-pane -U — stay in bottom)

# ── Keys ──
tmux bind-key -n q "kill-session"
tmux bind-key -n r "command-prompt -p 'issue number:' 'run-shell \"bash $MC_DIR/scripts/spawn-worker.sh %% $MODEL\"'"

# Numeric keys 1-9
for i in $(seq 1 9); do
  tmux bind-key -n "$i" "run-shell 'bash $MC_DIR/scripts/spawn-worker.sh $i $MODEL'"
done

echo "✅ TUI ready — tmux attach -t $SESSION"
echo "   Press 1-9 to run worker, q to quit"
echo "   Model: $MODEL"
