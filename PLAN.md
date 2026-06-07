---
created: 2026-06-07
updated: 2026-06-07
status: ready — Phase 1
---

# PLAN.md — что строим и как

## 🎯 Две фазы

```
Phase 1: Bash (выходные)        Phase 2: Go/Meow (потом)
┌─────────────────────┐         ┌──────────────────────┐
│ 4 скрипта            │         │ Форк Kaiukov/meow    │
│ ~150 строк           │  ──▶   │ + GitHub adapter     │
│ 0 зависимостей       │         │ + bubbletea TUI      │
│ tmux-native TUI      │         │ + Opus orchestrator  │
└─────────────────────┘         └──────────────────────┘
```

## 📋 Phase 1: Bash (4 задачи)

### Task 0: Сетап

```bash
mkdir -p /root/mission-control/{scripts,state,tasks,logs}
cd /root/mission-control
git init
```

**Создать `.env`:**
```bash
GITHUB_TOKEN=ghp_...
GITHUB_REPO=Kaiukov/my-portfolio
MC_DIR=/root/mission-control
```

**Done when:** директории есть, `.env` заполнен.

---

### Task 1: `pull-issues.sh` — GitHub → tasks.json

**Что:** читает GitHub Issues API и пишет локальный JSON.

```bash
#!/usr/bin/env bash
# scripts/pull-issues.sh
set -euo pipefail
source /root/mission-control/.env

OUTFILE="$MC_DIR/state/tasks.json"

echo '{"updated":"'"$(date -Iseconds)"'","tasks":[' > "$OUTFILE"

gh issue list --repo "$GITHUB_REPO" --state open --limit 50 \
  --json number,title,labels,state,createdAt,url \
  --jq '.[] | {number,title,labels:[.labels[].name],state,url,created:.createdAt}' \
  | sed '$!s/$/,/' >> "$OUTFILE"

echo ']}' >> "$OUTFILE"
echo "✓ $(jq '.tasks | length' "$OUTFILE") issues pulled"
```

**Проверка:**
```bash
bash scripts/pull-issues.sh
cat state/tasks.json | jq '.tasks[0] | {number,title,labels}'
```

---

### Task 2: `tui.sh` — терминальный дашборд

**Что:** tmux сессия с двумя панелями — задачи сверху, терминал снизу.

```bash
#!/usr/bin/env bash
# tui.sh
set -euo pipefail
source /root/mission-control/.env
SESSION="mc"

tmux kill-session -t "$SESSION" 2>/dev/null || true
tmux new-session -d -s "$SESSION" -x 140 -y 40

# Верхняя панель: список задач (обновляется каждые 3с)
tmux send-keys -t "$SESSION" \
  "watch -n 3 'cat $MC_DIR/state/tasks.json | jq -r \".tasks[] | \\\"\(.state) #\(.number) \(.title)\\\"\"'" C-m

# Нижняя панель: терминал
tmux split-window -v -t "$SESSION" -p 40
tmux send-keys -t "$SESSION" "echo '🛸 Mission Control — выбери задачу ↑'" C-m

# Бинды
tmux bind-key -t "$SESSION" r "run-shell 'bash $MC_DIR/scripts/spawn-worker.sh'"
tmux bind-key -t "$SESSION" s "run-shell 'tmux kill-session -t \$(tmux display-message -p \"#S\")'"

tmux attach -t "$SESSION"
```

**Проверка:**
```bash
bash scripts/tui.sh
# Должен открыться tmux с задачами сверху и терминалом снизу
```

---

### Task 3: `spawn-worker.sh` — Codex в tmux

**Что:** запускает Codex CLI в отдельной tmux сессии для конкретного issue.

```bash
#!/usr/bin/env bash
# scripts/spawn-worker.sh
set -euo pipefail
source /root/mission-control/.env

TASK_NUM="${1:?Usage: spawn-worker.sh <issue-number>}"
SESSION="task-$TASK_NUM"
MODEL="${2:-deepseek-v3}"

# Читаем issue
ISSUE=$(gh issue view "$TASK_NUM" --repo "$GITHUB_REPO" --json title,body)
TITLE=$(echo "$ISSUE" | jq -r '.title')
BODY=$(echo "$ISSUE" | jq -r '.body')

# Клонируем репо если надо
REPO_DIR="$MC_DIR/repo"
[ -d "$REPO_DIR" ] || git clone "git@github.com:${GITHUB_REPO}.git" "$REPO_DIR"

# Убиваем старую сессию
tmux kill-session -t "$SESSION" 2>/dev/null || true

# План от оркестратора (если есть)
PLAN_FILE="$MC_DIR/tasks/plan-${TASK_NUM}.md"
PLAN_CTX=""
[ -f "$PLAN_FILE" ] && PLAN_CTX="Plan: $(cat "$PLAN_FILE")"

# Запускаем
tmux new-session -d -s "$SESSION" -x 120 -y 40 -c "$REPO_DIR" \
  "codex exec --model $MODEL '
GitHub Issue #${TASK_NUM}: ${TITLE}

${BODY}

${PLAN_CTX}
' 2>&1 | tee $MC_DIR/logs/${SESSION}.log"

echo "✓ Worker spawned: tmux attach -t $SESSION"
echo "  Log: logs/${SESSION}.log"
```

---

### Task 4: `start.sh` — запуск всего

```bash
#!/usr/bin/env bash
# start.sh
set -euo pipefail
cd /root/mission-control
source .env

echo "🛸 Mission Control"
echo ""

# Pull fresh issues
echo "→ Pulling issues..."
bash scripts/pull-issues.sh

# Launch TUI
echo "→ Launching TUI..."
bash scripts/tui.sh
```

```bash
chmod +x scripts/*.sh start.sh
```

---

## 🏗️ Phase 2: Meow fork (после MVP)

**Репозиторий:** `Kaiukov/meow` (форк `akatz-ai/meow`)

**Что допилить:**
- [ ] GitHub Issues adapter — сейчас Meow работает из workflow/TOML, а не из Issues
- [ ] Opus orchestrator adapter — уже есть `claude-opus` adapter
- [ ] Worker adapters — уже есть `codex`, `opencode`
- [ ] TUI improvements — bubbletea-based task board (сейчас Meow CLI-focused)

## 📁 Итоговая структура

```
/root/mission-control/
├── start.sh              ← одна команда запуска
├── scripts/
│   ├── pull-issues.sh    ← GitHub → tasks.json
│   ├── tui.sh            ← tmux дашборд
│   └── spawn-worker.sh   ← Codex в tmux
├── state/
│   └── tasks.json        ← состояние задач
├── tasks/                ← планы и ревью от Opus
├── logs/                 ← логи воркеров
├── repo/                 ← клон репозитория
└── .env                  ← токены и пути
```

**Phase 2 добавляет:**
```
/root/meow/               ← форк Meow (Go)
├── cmd/meow/cmd/adapters/  ← адаптеры моделей
└── internal/               ← workflow engine
```
