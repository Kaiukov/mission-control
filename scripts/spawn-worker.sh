#!/usr/bin/env bash
# spawn-worker.sh — Codex CLI in tmux for a specific GitHub issue
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MC_DIR="${MC_DIR:-$(dirname "$SCRIPT_DIR")}"
[ -f "$MC_DIR/.env" ] && source "$MC_DIR/.env"

GITHUB_REPO="${GITHUB_REPO:-Kaiukov/mission-control}"
TASK_NUM="${1:?Usage: spawn-worker.sh <issue-number> [model]}"
MODEL="${2:-opencode-go/deepseek-v4-flash}"
SESSION="task-${TASK_NUM}"
REPO_DIR="$MC_DIR/repo"
PLAN_FILE="$MC_DIR/tasks/plan-${TASK_NUM}.md"
TUI_SESSION="mc"

echo "🚀 Spawning worker for #${TASK_NUM} (model: $MODEL)..."

# ── Mark as running ──
mkdir -p "$MC_DIR/state"
cat > "$MC_DIR/state/running.json" << RUNEOF
{"number":$TASK_NUM,"started":"$(date -Iseconds)","model":"$MODEL"}
RUNEOF

# Get issue details
ISSUE_JSON=$(gh issue view "$TASK_NUM" --repo "$GITHUB_REPO" --json title,body,state 2>/dev/null || echo '{"title":"unknown","body":"","state":"unknown"}')
TITLE=$(echo "$ISSUE_JSON" | jq -r '.title // "unknown"')
BODY=$(echo "$ISSUE_JSON" | jq -r '.body // ""')
STATE=$(echo "$ISSUE_JSON" | jq -r '.state // "unknown"')

if [ "$STATE" = "unknown" ]; then
  echo "❌ Issue #${TASK_NUM} not found in $GITHUB_REPO"
  rm -f "$MC_DIR/state/running.json"
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
1. Read the issue above and the plan if provided
2. Implement the fix or feature
3. Test it works
4. Commit with message referencing #${TASK_NUM}
5. Do NOT create a PR"

# Spawn worker in tmux
tmux new-session -d -s "$SESSION" -x 120 -y 40 -c "$REPO_DIR" \
  "echo '🔧 Worker #${TASK_NUM} (${MODEL}) — \$(date +%H:%M)'; echo ''; codex exec --model ${MODEL} '${PROMPT}' 2>&1 | tee $MC_DIR/logs/${SESSION}.log; EXIT=\${PIPESTATUS[0]}; echo ''; echo '✅ Worker #${TASK_NUM} done (exit='\${EXIT})'; rm -f $MC_DIR/state/running.json"

sleep 1
echo "✓ Worker spawned: tmux attach -t $SESSION"
echo "  Log: logs/${SESSION}.log"
echo "  Tail: tail -f $MC_DIR/logs/${SESSION}.log"

# ── Show worker output in TUI bottom pane ──
if tmux has-session -t "$TUI_SESSION" 2>/dev/null; then
  sleep 2
  tmux send-keys -t "${TUI_SESSION}.1" C-c Enter 2>/dev/null || true
  sleep 0.2
  tmux send-keys -t "${TUI_SESSION}.1" "clear && echo '🔧 Worker #${TASK_NUM} (${MODEL}) — live' && echo '' && tail -f $MC_DIR/logs/${SESSION}.log" Enter
fi
