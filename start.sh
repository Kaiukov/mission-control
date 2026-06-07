#!/usr/bin/env bash
# start.sh — launch Mission Control
set -euo pipefail

export MC_DIR="${MC_DIR:-/root/mission-control}"
cd "$MC_DIR"

# Source .env if exists
[ -f "$MC_DIR/.env" ] && source "$MC_DIR/.env"

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║          🛸 MISSION CONTROL              ║"
echo "║          v0.1 — Phase 1                  ║"
echo "╚══════════════════════════════════════════╝"
echo ""
echo "→ Repo: ${GITHUB_REPO:-Kaiukov/my-portfolio}"
echo ""

# Pull fresh issues
echo "→ Pulling GitHub Issues..."
bash "$MC_DIR/scripts/pull-issues.sh"
echo ""

# Launch TUI
echo "→ Launching TUI..."
echo "   (Ctrl+b q to quit, Ctrl+b ↑↓ to switch panes)"
echo ""
sleep 1
bash "$MC_DIR/scripts/tui.sh"
