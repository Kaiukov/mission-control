---
created: 2026-06-07
status: discussion
---

# Option B: Terminal UI вместо Web UI

## Суть

Вместо HTML дашборда в браузере — всё работает в терминале. Зашёл по SSH → увидел доску задач и терминалы агентов.

## Два подхода к TUI

### B1: tmux-native (самый простой)

```
┌─────────────────────┬──────────────────────────┐
│                     │                          │
│   Task Board        │   Active Terminal        │
│   ─────────         │   ───────────────        │
│   ✓ #41 Done        │   $ codex exec ...      │
│   ⏳ #42 Running     │   Thinking...            │
│   ○ #43 Ready       │   Writing code...        │
│   ○ #44 Open        │                          │
│                     │                          │
│   [r]un [s]top      │                          │
│   [p]r [m]erge      │                          │
├─────────────────────┴──────────────────────────┤
│  Status: 2 done | 1 running | 2 ready           │
└────────────────────────────────────────────────┘
```

**Как:** один bash-скрипт, который:
1. Создаёт tmux layout с разделёнными панелями
2. Левая панель — `watch`-скрипт, обновляющий список задач из tasks.json
3. Правая панель — живой терминал выбранного агента

**Плюсы:**
- 0 зависимостей — только bash + tmux (уже есть)
- 0 строк кода на фронтенд
- Работает через SSH
- Клавиатурное управление естественное

**Минусы:**
- Нельзя мышкой
- Меньше визуальной гибкости

### B2: Полноценный TUI (ratatui/textual/bubbletea)

Как `lazygit` но для AI-агентов. Табы, цветные карточки, скроллинг.

**Плюсы:** красиво, интерактивно
**Минусы:** +1 язык/фреймворк, дольше писать

---

## Сравнение: Web UI vs TUI

| | Web UI | TUI (tmux-native) |
|---|---|---|
| **Зависимости** | Node.js, npm, Next.js, ws | 0 (bash + tmux) |
| **Строк кода** | ~300+ (HTML, JS, сервер) | ~50 (один bash-скрипт) |
| **Доступ** | Браузер (:3000) | SSH |
| **Мобильный** | Да (но не нужно) | Нет |
| **Drag & drop** | Да (Phase 2) | Нет |
| **Латентность** | WebSocket hop | Прямой вывод в терминал |
| **Интеграция с tmux** | Через WebSocket | Нативно |
| **Красота** | Можно стилизовать | ASCII-таблица |
| **Время до MVP** | 4-6 часов | **30 минут** |

---

## Что это меняет в архитектуре

```
Было (Web UI):
  Дашборд (HTML/Next.js) ← WebSocket → ws-server.js → tmux

Стало (TUI):
  SSH → tmux сессия → bash-скрипт → tasks.json + агенты в панелях
```

**Выбрасываем:**
- ❌ `public/index.html`
- ❌ `ws-server.js`
- ❌ `package.json` / Next.js
- ❌ WebSocket слой

**Остаётся:**
- ✅ `pull-issues.sh` — GitHub → tasks.json
- ✅ `spawn-worker.sh` — Codex в tmux
- ✅ `run-orchestrator.sh` — Opus оркестрирует
- ✅ `state/tasks.json` — состояние
- ✅ `start.sh` — запуск TUI сессии

**Новое:**
- ✅ `tui.sh` — tmux layout + watch-скрипт

---

## Как это выглядит (tmux-native)

```bash
#!/usr/bin/env bash
# tui.sh — запуск Mission Control в терминале

SESSION="mission-control"

# Убить старую сессию если есть
tmux kill-session -t "$SESSION" 2>/dev/null

# Создать сессию
tmux new-session -d -s "$SESSION" -x 140 -y 40

# Панель 1 (левая, 40%): список задач
tmux send-keys -t "$SESSION" \
  "watch -n 5 'cat /root/mission-control/state/tasks.json | jq -r \"\"\"\""
# [сложный jq формат вывода таблицы]

# Панель 2 (правая, 60%): приветствие
tmux split-window -h -t "$SESSION"
tmux send-keys -t "$SESSION" "echo 'Нажми [r] для запуска задачи'"
```

Горячие клавиши прямо в tmux:
- `Ctrl+b r` — запустить выбранную задачу
- `Ctrl+b s` — остановить воркера
- `Ctrl+b p` — создать PR
- `Ctrl+b m` — merge

---

## Рекомендация

**B1 (tmux-native) для Phase 1.** 

Причины:
1. 30 минут до рабочего прототипа вместо 4-6 часов
2. 0 новых зависимостей
3. Ты всё равно работаешь через SSH на CT 113
4. Если потом захочется Web UI — `tui.sh` останется как fallback
5. Mike (TANK) тоже использует tmux, просто обернул в веб

Web UI (Phase 2) имеет смысл когда:
- Будут скриншоты (drag & drop)
- Захочется смотреть с телефона
- Нужны красивые графики usage
