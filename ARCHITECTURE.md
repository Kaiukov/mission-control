---
created: 2026-06-07
updated: 2026-06-07
---

# ARCHITECTURE — как это работает

## 🎯 Главный принцип: Opus дорогой → только оркестрирует

```
Opus (дорогой) → ДУМАЕТ: читает задачи, пишет план, назначает, ревьюит
DeepSeek/GPT-mini (дешёвые) → ДЕЛАЮТ: пишут код, тесты, документацию
```

**Почему:** Claude Opus стоит ~$15/1M токенов. DeepSeek V3 — ~$0.27/1M токенов. Разница в 55 раз. Opus нужен только для сложного планирования и review — всё остальное делают дешёвые модели.

## 🧩 Компоненты

```
SSH → tmux сессия "mission-control"
│
├─ Верхняя панель: список задач (fzf / watch + jq)
│     ⏳ #42 Fix login     deepseek-v3  2m
│     ○ #43 Add dark mode   ○ ready
│
├─ Нижняя панель: живой терминал выбранного агента
│     $ codex exec --model deepseek-v3 "Fix #42..."
│     > Analyzing issue...
│
└─ Горячие клавиши: r=run s=stop p=PR m=merge

────────────────────────────────────────────

Состояние:
  tasks.json  ←  pull-issues.sh (GitHub Issues)
  history.db  ←  SQLite (история чатов)

Агенты:
  Opus (оркестратор) → читает tasks.json → план → назначает воркера → review
  DeepSeek/GPT mini (воркеры) → Codex CLI в tmux → пишут код
```

## 💰 Экономика

| Роль | Модель | Цена за 1M токенов | На что тратит |
|---|---|---|---|
| **Оркестратор** | Claude Opus 4 | ~$15 input / $75 output | Чтение issue → план → review PR |
| **Воркер (код)** | DeepSeek V3 | ~$0.27 / $1.10 | Пишет код, тесты |
| **Воркер (лёгкий)** | GPT 5.4 mini | ~$0.15 / $0.60 | Мелкие правки, документация |
| **Воркер (сложный)** | DeepSeek R1 | ~$0.55 / $2.19 | Рефакторинг, архитектура |

**Пример задачи (типичный fix):**

| Этап | Кто | Токенов | Стоимость |
|---|---|---|---|
| Прочитать issue, написать план | Opus | 2K in / 1K out | ~$0.10 |
| Написать код | DeepSeek V3 | 3K in / 5K out | ~$0.006 |
| Review результата | Opus | 2K in / 0.5K out | ~$0.07 |
| Правки если надо | DeepSeek | 1K in / 2K out | ~$0.003 |
| **Итого** | | | **~$0.18** |

Если бы Opus делал всё сам: **~$1.50-3.00** за ту же задачу. Экономия: **в 10-15 раз**.

## 🔄 Поток: задача от Issue до Merge

### 1. New Issue → tasks.json

```
GitHub Issue #42 "Fix login bug"
        │
        ▼
pull-issues.sh → tasks.json
  {
    number: 42,
    title: "Fix login bug",
    labels: ["bug"],
    status: "open"
  }
```

### 2. Оркестратор берёт задачу

```
User (или cron) запускает:
  run-orchestrator.sh

Opus читает tasks.json:
  → Видит #42 (open)
  → Читает issue body: gh issue view 42
  → Пишет план: tasks/plan-42.md
  → Выбирает модель: bug → DeepSeek V3 (хорош для fixes)
  → Спавнит воркера: spawn-worker.sh 42 --model deepseek-v3
```

### 3. Воркер работает

```
spawn-worker.sh:
  → Clone repo если надо
  → Создать ветку: fix/issue-42
  → Запустить Codex CLI в tmux:
      tmux new -s task-42 "codex exec --model deepseek-v3
        'GitHub Issue #42: Fix login bug
         Plan: tasks/plan-42.md
         Steps: 1) Reproduce, 2) Fix, 3) Test, 4) Commit'
      "
  → Весь вывод → в лог + WebSocket в дашборд
```

### 4. Review оркестратором

```
Worker done → сигналит Opus

Opus смотрит результат:
  → Читает diff (git diff main...fix/issue-42)
  → Если ок:
      → gh pr create --title "Fix #42" --body "..."
      → tasks.json: status → "review"
  → Если не ок:
      → Пишет замечания в tasks/review-42.md
      → Возвращает воркеру с правками
```

### 5. Merge (ручной или авто)

```
PR создан → GitHub Actions CI
  → Если тесты зеленые:
      → Кнопка Merge в дашборде (или авто)
      → gh pr merge --auto
      → tasks.json: status → "done"
  → Если красные:
      → Задача возвращается воркеру
```

## 📁 Структура файлов

```
/root/mission-control/
├── start.sh                  # Запуск всего
│
├── scripts/                  # Shell скрипты
│   ├── pull-issues.sh        # GitHub Issues → tasks.json
│   ├── run-orchestrator.sh   # Opus читает + назначает
│   ├── spawn-worker.sh       # Codex CLI в tmux (с выбором модели)
│   ├── review-task.sh        # Opus review результата
│   └── ws-server.js          # WebSocket + HTTP API
│
├── tasks/                    # Планы и review (на задачу)
│   ├── plan-42.md            # План от Opus для #42
│   └── review-42.md          # Review замечания
│
├── state/                    # JSON + SQLite
│   ├── tasks.json            # Все задачи
│   └── history.db            # SQLite: чаты, история
│
├── public/                   # Статика дашборда
│   └── index.html
│
├── logs/                     # Логи агентов
│   └── task-42.log
│
├── repo/                     # Клон репозитория
│   └── (git clone)
│
└── .env                      # GitHub token, пути
```

## 🎭 Модели для разных типов задач

| Тип задачи | Модель воркера | Почему |
|---|---|---|
| `bug` | DeepSeek V3 | Хорош в отладке, дёшев |
| `enhancement` | DeepSeek V3 | Новая функциональность |
| `refactor` | DeepSeek R1 / Opus | Сложная логика, нужен ум |
| `docs` | GPT 5.4 mini | Простой текст, дёшев |
| `research` | DeepSeek R1 | Нужен reasoning |
| `tests` | DeepSeek V3 | Стандартная задача |

**Как выбирается:** Opus сам решает по labels + содержимому задачи. Можно форсировать через `/run --model deepseek-v3`.

## 🚦 Статусная модель

```
open ──→ ready ──→ running ──→ review ──→ done
  │                          │
  │                          ├──→ blocked (ждёт человека)
  │                          └──→ rework (правки)
  │
  └──→ archived (закрыли без работы)
```

- **open** — только создан в GitHub
- **ready** — Opus прочитал, составил план
- **running** — воркер работает в tmux
- **review** — Opus проверяет результат
- **blocked** — нужен человек (вопрос, решение)
- **rework** — Opus вернул на доработку
- **done** — смержен, закрыт

## 🔧 Почему tmux а не Docker

- **Проще** — не нужен Docker daemon, образы, volume mounting
- **Нативно** — агент работает как обычный CLI, видит всю файловую систему
- **Дешевле** — нет оверхеда контейнера
- **Mike тоже так сделал** — проверено на TANK

Но: **никогда не под root**. Каждый tmux под `kaiukov` или отдельным пользователем.
