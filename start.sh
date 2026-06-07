#!/usr/bin/env bash
# start.sh — launch Mission Control (Go or Bash backend)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export MC_DIR="${MC_DIR:-$SCRIPT_DIR}"
cd "$MC_DIR"
[ -f "$MC_DIR/.env" ] && source "$MC_DIR/.env"

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║          🛸 MISSION CONTROL              ║"
echo "║          v0.2 — Go + Bash                ║"
echo "╚══════════════════════════════════════════╝"
echo ""
echo "→ Repo: ${GITHUB_REPO:-Kaiukov/mission-control}"
echo ""

# Pull fresh issues
echo "→ Pulling issues..."
if [ -x "$MC_DIR/mc" ]; then
  ./mc pull
else
  bash "$MC_DIR/scripts/pull-issues.sh"
fi
echo ""

# Launch TUI
echo "→ Launching..."
if [ -x "$MC_DIR/mc" ]; then
  exec ./mc tui
elif [ -t 0 ]; then
  bash "$MC_DIR/scripts/tui.sh"
  exec tmux attach -t mc
else
  bash "$MC_DIR/scripts/tui.sh"
fi
