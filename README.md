# 🛸 Mission Control

Персональный веб-дашборд для управления AI-агентами при разработке.

**Принцип:** Claude Opus ($15/M токенов) думает, дешёвые модели ($0.15/M) пишут код. Экономия 10-15x.

## Быстрый старт (Phase 1)

```bash
bash start.sh
# Дашборд: http://localhost:3000
```

## Документация

- [WILL.md](./WILL.md) — философия, принципы, roadmap
- [ARCHITECTURE.md](./ARCHITECTURE.md) — архитектура, стек, поток данных
- [RESEARCH.md](./RESEARCH.md) — анализ аналогов (TANK, cmux, Hermes Kanban)
- [PLAN.md](./PLAN.md) — пошаговый план реализации (6 задач)

## Архитектура (5 секунд)

```
💰 Opus ($15/M) → ДУМАЕТ: план, assign, review
💸 DeepSeek/GPT mini ($0.15/M) → ДЕЛАЕТ: код, тесты
📱 HTML дашборд → WebSocket → живой tmux в браузере
🐙 GitHub: Issue → PR → merge
```

## Стек

| Слой | Технология |
|---|---|
| Дашборд | HTML + WebSocket (Next.js Phase 2) |
| Оркестратор | Claude Code CLI (Opus) |
| Воркеры | Codex CLI (DeepSeek V3, GPT 5.4 mini) |
| Терминалы | tmux |
| Git | GitHub |
| Состояние | tasks.json + SQLite |

## Статус

🟡 **Phase 1 — планирование.** Код не написан.

- [ARCHITECTURE.md](./ARCHITECTURE.md) — архитектура готова ✅
- [RESEARCH.md](./RESEARCH.md) — анализ аналогов готов ✅
- [PLAN.md](./PLAN.md) — 6 задач Phase 1 расписаны ✅
