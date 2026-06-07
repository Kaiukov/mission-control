---
created: 2026-06-07
source: YouTube https://youtu.be/hXWwqPgexZU
---

# RESEARCH — анализ аналогов

## TANK (Mike's Mission Control)

**Автор:** Mike (Creator Magic), видео: [youtu.be/hXWwqPgexZU](https://youtu.be/hXWwqPgexZU)
**Название тула:** TANK
**Статус:** Закрытый исходный код, доступен через Creator Magic Premium (coupon: `tank`)

### Что это

Веб-дашборд для управления Claude Code агентами. Работает на своём сервере. Ключевая идея — "я бутылочное горлышко, нужно mission control чтобы следить за всем сразу".

### Структура видео (главы)

1. **0:00** — Introduction: Mission Control for Claude Code
2. **0:34** — Starting a New Task & UI Overview
3. **1:16** — Anthropic Terms of Service & Terminal Hooks
4. **2:49** — Why Build "Tank" & Moving Away from GitHub (→ Forgejo)
5. **3:52** — Markdown To Do List Integration
6. **5:10** — Creating a Repo Flattening Feature (реальный пример)
7. **6:57** — Multitasking & General LLM Queries
8. **8:35** — Real Time Usage Tracking & Database Storage
9. **9:59** — Persistent Sessions with tmux & Drag and Drop Images
10. **11:39** — Completing & Testing the Feature
11. **13:23** — Visual Input Testing with Screenshots
12. **15:11** — Project Icon Customization
13. **15:36** — Security Warnings & Community Access

### Архитектура (из видео)

```
Web UI (React/Next.js)  ← hooks/WebSocket →  Backend (Node.js)
                                                   │
                              ┌────────────────────┼────────────────────┐
                              │                    │                    │
                         tmux sessions        Forgejo (git)         SQLite
                         (1 per agent)        (self-hosted)      (история чатов)
                              │
                    Claude Code CLI (интерактивный)
```

### Ключевые механики (из транскрипта)

#### 1. Новая задача (New Task button)
- Кнопка "New Task" в правом верхнем углу
- Вводишь промпт → локальный AI даёт имя задаче
- Claude Code стартует в pop-up терминале
- Терминал — это живое tmux окно, встроенное в веб

#### 2. Terminal Hooks (механика синхронизации)
> "I'm actually using hooks to talk back and forth between Claude Code. Everything that's happening here in the terminal is being echoed right here in my mission control."

- Claude Code работает в интерактивном режиме (не programmatic/API)
- Хуки перехватывают вывод терминала и шлют в UI через WebSocket
- Это позволяет Mike оставаться в "white area" Anthropic ToS (интерактивное использование)

#### 3. Кнопки действий
- **Open PR** — создать pull request в Forgejo
- **Merge** — смержить PR
- **Mark Done** — отметить задачу выполненной
- **Commit & Merge** — одной кнопкой

> "I've actually put in icons here that allow me to open pull requests and even mark the task as done."

#### 4. Статусные индикаторы
- **Жёлтая мигающая точка** — задача выполняется (running)
- **Синие** — выполненные задачи (completed)
- Статус обновляется в реальном времени

#### 5. Multi-agent (All View)
> "If I go to the all view here, I can see everything running at the same time."

- Можно запустить несколько Claude Code одновременно
- Например: repo flattening feature + поиск мексиканской еды
- Каждый в своём tmux окне, все видны в дашборде

#### 6. TODO.md интеграция
- TODO.md лежит в репозитории (Forgejo)
- Зеркалится в UI дашборда
- Формат: markdown с чекбоксами
- При изменении в UI → коммитится в репо

#### 7. Usage Tracking
> "I've actually created this dashboard to update my five hourly usage and weekly usage in real time."

- Показывает % от 5-часового лимита Claude Max
- Показывает % от недельного лимита
- В реальном времени

#### 8. SQLite база данных
> "I'm using a very simple SQLite database to make sure all the data is right there. I never lose a Claude Code chat again because it's recorded inside my dashboard."

- Все чаты сохраняются
- Можно resume любую сессию
- История не теряется при закрытии tmux

#### 9. tmux как основа
> "You may notice this green bar down the bottom. Yes, I'm running every Claude Code session in a tmux window."

- Зелёная полоса внизу терминала = tmux status bar
- Каждый Claude Code в своём tmux окне
- Не нужен VPS для персистентности — tmux на локальной машине

#### 10. Drag & Drop скриншотов
- Можно перетащить скриншот в поле ввода промпта
- Скриншот загружается на сервер → попадает в промпт Claude Code
- Claude Code видит и анализирует изображение

#### 11. Локальный CI/CD
> "I've actually got this setting up the job and installing the SSH key. This is all happening locally right now."

- После merge → запускается CI локально
- Передеплой самого TANK
- Smoke test после деплоя

#### 12. Repo Flattening — пример фичи построенной TANK'ом
- Mike показал как TANK сам себе написал фичу "flatten repo → text file"
- Полный цикл: задача → Claude Code → PR → merge → деплой → кнопка в UI

### UI (восстановлено по описанию)

```
┌──────────────────────────────────────────────┐
│  TANK                          [New Task] [⚙]│
├────────┬─────────────────────────────────────┤
│ Tasks  │                                     │
│        │  #1 Create app favicon     ✓ done   │
│ ● #2   │  #2 Add repo flattening    ⏳ running│
│ ✓ #1   │  ┌──────────────────────────────┐   │
│        │  │ Terminal: task-2              │   │
│        │  │ $ claude "add flattening..." │   │
│        │  │ Thinking...                   │   │
│        │  │ Writing flatten.go...         │   │
│        │  └──────────────────────────────┘   │
│        │                                     │
│        │  [Open PR] [Merge] [Mark Done]      │
│        │                                     │
│        │  Usage: 10% hourly | 3% weekly      │
├────────┴─────────────────────────────────────┤
│ Projects │ Chats │ Calendar │ Settings        │
└──────────────────────────────────────────────┘
```

**Навигация (левая панель):**
- Tasks — список задач
- Projects — переключение между проектами
- Chats — история всех чатов
- Calendar — календарь (упомянут в UI)

### Стек (реконструкция)

| Слой | Что | Почему |
|---|---|---|
| **Фронт** | Next.js / React | SPA с WebSocket |
| **Бэк** | Node.js | WebSocket + tmux control + Forgejo API |
| **Git** | Forgejo | Свой Git (после уязвимости GitHub private repos) |
| **БД** | SQLite | Чаты, история, задачи |
| **Терминалы** | tmux | Каждый Claude Code в своём окне |
| **Агенты** | Claude Code CLI | ИНТЕРАКТИВНЫЙ режим (ключевое!) |
| **Хуки** | Кастомные | Перехват вывода Claude Code → WebSocket |
| **CI/CD** | Локальный runner | После merge → деплой самого TANK |
| **Naming** | Local AI | Лёгкая модель даёт имена задачам |

### Что Mike говорит о безопасности

> "Every agent here runs with permissions skipped because I like to live dangerously. Run it isolated, have a lock down to test it and never run it as root."

- Claude Code запускается с `--yolo` (без подтверждений)
- Рекомендует: изолированная среда, не под root
- Сам TANK передеплоится через собственный CI

### Что беру из TANK

- ✅ tmux для терминалов — зелёный status bar, несколько сессий
- ✅ WebSocket для live-вывода терминала
- ✅ SQLite для истории чатов (resume сессии)
- ✅ Статусные индикаторы (running/completed/error)
- ✅ Кнопки действий (Open PR, Merge, Done) в дашборде
- ✅ TODO.md как source of truth
- ✅ Концепция "New Task" → промпт → агент
- ✅ Multi-agent All View
- ✅ Drag & drop для скриншотов (Phase 2)

### Что НЕ беру / адаптирую

- ❌ Forgejo → GitHub (уже есть)
- ❌ Claude Code only → Codex CLI + Hermes (агностик)
- ❌ Usage tracking → не надо (пока)
- ❌ Local AI для naming → YAGNI (Phase 1)
- ❌ Локальный CI/CD → GitHub Actions
- ⚠️ Interactive mode only → И для интерактива, и для автономного запуска

---

## cmux-todo-board (мой плагин)

**Репозиторий:** `Kaiukov/claude-code-cmux-todo-plugin`

### Что делает
- GitHub Issues → board.json + TODO.md
- Планирование задач в Claude Code
- Диспатч через cmux панели

### Что беру
- ✅ `pull-issues.sh` логику (gh → JSON)
- ✅ Формат board.json
- ✅ Статусную модель (ready, in_progress, blocked, done)

---

## Hermes Agent Kanban

Встроенная система в Hermes: `hermes kanban <verb>`

### Что умеет
- Доска задач (SQLite)
- Assign на профили/воркеров
- Auto-диспатч ready задач
- Heartbeat контроль воркеров

### Что беру
- ✅ Концепцию ready → dispatch → monitor
- ❌ Саму систему (слишком завязана на Hermes профили, не годится для внешних CLI-агентов)

---

## Вывод

Все три системы делают одно: задача → агент → результат. Разница в транспорте:

| | TANK | cmux | Мой Mission Control |
|---|---|---|---|
| **Агенты** | Claude Code only | Claude Code only | Любой CLI |
| **Git** | Forgejo | GitHub | GitHub |
| **UI** | Веб | Claude Code in-app | Веб (свой) |
| **Состояние** | SQLite | board.json | tasks.json + SQLite |
| **Терминалы** | tmux | cmux | tmux |
| **Оркестрация** | Claude Code hooks | Claude Code skills | Opus + скрипты |

Мой подход — взять лучшее из всех трёх: **tmux терминалы (TANK) + GitHub Issues (cmux) + простые скрипты вместо монолита**.
