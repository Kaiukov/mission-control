---
created: 2026-06-07
status: v2 — TUI native
---

# Реализация TUI (Terminal UI)

## Финальный дизайн

```
┌──────────────────────────────────────────────────┐
│  🛸 Mission Control    1 running │ 2 ready       │ ← tmux status bar
├──────────────────────────────────────────────────┤
│                                                  │
│  ⏳ #42 Fix login bug         deepseek-v3  2m   │ ← верх: список сессий
│  ○ #43 Add dark mode          ○ ready           │   стрелки ↑↓ выбор
│  ○ #44 Refactor auth          ○ ready           │   Enter → смотреть
│  ✓ #41 Update deps            ✓ done            │   r → запустить
│  ✓ #40 Fix typo               ✓ done            │   p → PR
│                                                  │
├──────────────────────────────────────────────────┤
│                                                  │
│  Terminal: task-42 (deepseek-v3)                │ ← низ: живой терминал
│  ─────────────────────────────────              │   выбранной сессии
│  $ codex exec --model deepseek-v3 ...           │
│  > Analyzing issue #42...                        │
│  > Found bug in src/auth.ts:142                  │
│  > Writing fix...                                │
│                                                  │
└──────────────────────────────────────────────────┘
```

## Как работает выбор сессии

### Вариант A: fzf (рекомендую)

```bash
# fzf показывает список, Enter → attach к tmux сессии
tmux list-sessions -F "#{session_name} #{?session_attached,●,○}" \
  | fzf --header="Select agent session" \
  | awk '{print $1}' \
  | xargs -I{} tmux switch-client -t {}
```

Плюсы: fuzzy search, красиво, 1 пакет (`fzf`)
Минусы: +1 зависимость

### Вариант B: select (zero deps)

```bash
# Встроенный bash select
select SESSION in $(tmux list-sessions -F "#{session_name}"); do
  tmux switch-client -t "$SESSION"
  break
done
```

Плюсы: 0 зависимостей
Минусы: менее удобно, нет поиска

### Вариант C: fzf + preview (максимально круто)

```bash
# fzf с превью последних строк терминала
tmux list-sessions -F "#{session_name}" \
  | fzf --preview="tmux capture-pane -t {} -p | tail -20" \
        --preview-window=bottom:50% \
        --bind="enter:execute(tmux switch-client -t {})+abort"
```

Плюсы: видно последние 20 строк терминала при наведении!
Минусы: +1 пакет

## Скрипт tui.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

SESSION_NAME="mc"
MC_DIR="/root/mission-control"

# Убить старую сессию если есть
tmux kill-session -t "$SESSION_NAME" 2>/dev/null || true

# Создать новую сессию с 2 панелями (верх/низ)
tmux new-session -d -s "$SESSION_NAME" -x 140 -y 40

# Верхняя панель (60%): список задач с fzf
tmux send-keys -t "$SESSION_NAME" \
  "watch -n 3 'cat $MC_DIR/state/tasks.json | jq -r \".tasks[] | \\\"\\(.status) #\\(.number) \\(.title)\\\"\"'" C-m

# Нижняя панель (40%): терминал (изначально пустой)
tmux split-window -v -t "$SESSION_NAME" -p 40
tmux send-keys -t "$SESSION_NAME" "echo 'Select a session above ↑'" C-m

# Биндим клавиши
# r — run task
tmux bind-key -t "$SESSION_NAME" r \
  "run-shell '$MC_DIR/scripts/spawn-worker.sh'"

# s — stop task (kill session)
tmux bind-key -t "$SESSION_NAME" s \
  "run-shell 'tmux kill-session -t \$(tmux display-message -p \"#S\")'"

# j/k — навигация по списку (стрелки и так работают)

tmux attach -t "$SESSION_NAME"
```

## Горячие клавиши

| Клавиша | Действие |
|---|---|
| `↑↓` / `j k` | Выбор задачи в списке |
| `Enter` | Открыть терминал выбранной сессии |
| `r` | Запустить воркера для выбранной задачи |
| `s` | Остановить текущего воркера |
| `p` | Создать PR (gh pr create) |
| `m` | Merge PR |
| `Ctrl+b ↑↓` | Переключение между панелями |
| `q` | Выйти из Mission Control |

## Что остаётся от Phase 1 плана

| Task | Статус | Изменения |
|---|---|---|
| 0. Сетап | ✅ почти | + `tui.sh`, убрать public/, ws-server.js |
| 1. pull-issues.sh | без изменений | GitHub → tasks.json |
| 2. ~~HTML дашборд~~ | ❌ выкинут | Заменён на tui.sh |
| 3. spawn-worker.sh | без изменений | Codex в tmux |
| 4. ~~ws-server.js~~ | ❌ выкинут | Не нужен — терминал нативный |
| 5. ~~Интеграция~~ | ❌ выкинут | Не нужна |
| **6. start.sh** | переписан | Запускает tui.sh |

**Итого: 4 задачи вместо 6. 0 внешних зависимостей.**
