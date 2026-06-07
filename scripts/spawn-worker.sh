#!/usr/bin/env bash
# spawn-worker.sh — Codex CLI in tmux for a specific GitHub issue
set -euo pipefail

# Auto-detect MC_DIR from script location
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MC_DIR="${MC_DIR:-$(dirname "$SCRIPT_DIR")}"

# Source .env if exists
[ -f "$MC_DIR/.env" ] && source "$MC_DIR/.env"

GITHUB_REPO="${GITHUB_REPO:-Kaiukov/mission-control}"
TASK_NUM="${1:?Usage: spawn-worker.sh <issue-number> [model]}"
MODEL="${2:-deepseek-v3}"
SESSION="task-${TASK_NUM}"
REPO_DIR="$MC_DIR/repo"
PLAN_FILE="$MC_DIR/tasks/plan-${TASK_NUM}.md"

echo "🚀 Spawning worker for #${TASK_NUM} (model: $MODEL)..."

# Get issue details
ISSUE_JSON=$(gh issue view "$TASK_NUM" --repo "$GITHUB_REPO" --json title,body,state 2>/dev/null || echo '{"title":"unknown","body":"","state":"unknown"}')
TITLE=$(echo "$ISSUE_JSON" | jq -r '.title // "unknown"')
BODY=$(echo "$ISSUE_JSON" | jq -r '.body // ""')
STATE=$(echo "$ISSUE_JSON" | jq -r '.state // "unknown"')

if [ "$STATE" = "unknown" ]; then
  echo "❌ Issue #${TASK_NUM} not found in $GITHUB_REPO"
  exit 1
fi

# Plan from orchestrator (if exists)
PLAN_CTX=""
if [ -f "$PLAN_FILE" ]; then
  PLAN_CTX="
Orchestrator's plan:
$(cat "$PLAN_FILE")"
fi

# Clone repo if needed
if [ ! -d "$REPO_DIR/.git" ]; then
  echo "→ Cloning $GITHUB_REPO..."
  git clone "git@github.com:${GITHUB_REPO}.git" "$REPO_DIR" 2>&1 | tail -1
fi

# Update repo
cd "$REPO_DIR"
git fetch origin 2>/dev/null
git checkout main 2>/dev/null || git checkout master 2>/dev/null
git pull origin main 2>/dev/null || git pull origin master 2>/dev/null

# Create branch
BRANCH="fix/issue-${TASK_NUM}"
git checkout -b "$BRANCH" 2>/dev/null || git checkout "$BRANCH"

# Kill existing session
tmux kill-session -t "$SESSION" 2>/dev/null || true
mkdir -p "$MC_DIR/logs"

# Build prompt
PROMPT="GitHub Issue #${TASK_NUM}: ${TITLE}

${BODY}
${PLAN_CTX}

Your task:
1. Read the issue above and the orchestrator's plan if provided
2. Implement the fix or feature
3. Write/update tests
4. Commit with message: 'fix: ${TITLE} (#${TASK_NUM})'
5. Do NOT create a PR — the orchestrator will do that"

# Spawn worker in tmux
tmux new-session -d -s "$SESSION" -x 120 -y 40 -c "$REPO_DIR" \
  "echo '🔧 Worker #${TASK_NUM} starting (model: ${MODEL})...'; echo ''; codex exec --model ${MODEL} '${PROMPT}' 2>&1 | tee $MC_DIR/logs/${SESSION}.log; echo ''; echo '✅ Worker #${TASK_NUM} finished'"

sleep 1
echo "✓ Worker spawned: tmux attach -t $SESSION"
echo "  Log: logs/${SESSION}.log"
echo "  Status: tmux list-sessions | grep $SESSION"
