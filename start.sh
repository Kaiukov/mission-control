#!/usr/bin/env bash
# start.sh — launch Mission Control
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export MC_DIR="${MC_DIR:-$SCRIPT_DIR}"
cd "$MC_DIR"

[ -f "$MC_DIR/.env" ] && source "$MC_DIR/.env"

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║          🛸 MISSION CONTROL              ║"
echo "║          v0.1 — Phase 1                  ║"
echo "╚══════════════════════════════════════════╝"
echo ""
echo "→ Repo: ${GITHUB_REPO:-Kaiukov/mission-control}"
echo ""

# Pull fresh issues
echo "→ Pulling GitHub Issues..."
bash "$MC_DIR/scripts/pull-issues.sh"
echo ""

# Check if we have a real terminal
if [ -t 0 ]; then
  echo "→ Launching TUI..."
  bash "$MC_DIR/scripts/tui.sh"
  echo ""
  echo "Attaching..."
  sleep 0.5
  exec tmux attach -t mc
else
  echo "→ No TTY — launching detached. Run: tmux attach -t mc"
  bash "$MC_DIR/scripts/tui.sh"
fi
