---
created: 2026-06-07
status: ready
---

# IMPLEMENTATION PLAN — Phase 1 (CLI прототип)

> **Goal:** CLI-скрипты + минимальный HTML дашборд. GitHub Issues → задачи → агенты в tmux.
> **Stack:** bash scripts + 1 HTML file + Next.js (позже)
> **Time:** ~2-3 часа на прототип

---

## ⚡ Task 0: Сетап проекта

**Objective:** Создать структуру проекта и установить зависимости

### Step 1: Создать директории

```bash
mkdir -p /root/mission-control/{scripts,state,public,logs}
cd /root/mission-control
git init
```

### Step 2: Создать .env

```bash
cat > /root/mission-control/.env << 'EOF'
GITHUB_TOKEN=$(gh auth token)
GITHUB_REPO=Kaiukov/my-portfolio
MISSION_CONTROL_DIR=/root/mission-control
EOF
```

### Step 3: Создать package.json (для Phase 2)

```bash
cd /root/mission-control
cat > package.json << 'EOF'
{
  "name": "mission-control",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev -p 3000",
    "build": "next build",
    "start": "next start"
  }
}
EOF
```

### Step 4: .gitignore

```bash
cat > /root/mission-control/.gitignore << 'EOF'
node_modules/
.next/
state/tasks.json
state/history.db
.env
logs/
*.log
EOF
```

**Done when:** директории созданы, `.env` заполнен, `package.json` есть.

---

## ⚡ Task 1: pull-issues.sh — GitHub Issues → tasks.json

**Objective:** Скрипт который читает GitHub Issues и записывает в локальный tasks.json

**Создать:** `/root/mission-control/scripts/pull-issues.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

REPO="${GITHUB_REPO:-Kaiukov/my-portfolio}"
OUTFILE="/root/mission-control/state/tasks.json"
NOW=$(date -Iseconds)

echo '{"updated":"'"$NOW"'","tasks":[' > "$OUTFILE"

gh issue list \
  --repo "$REPO" \
  --state open \
  --limit 50 \
  --json number,title,labels,state,createdAt,updatedAt,url \
  --jq '.[] | {
    number: .number,
    title: .title,
    labels: [.labels[].name],
    state: .state,
    url: .url,
    created: .createdAt,
    updated: .updatedAt
  }' | while IFS= read -r issue; do
    echo "$issue," >> "$OUTFILE"
done

# Remove trailing comma from last entry
sed -i '$ s/,$//' "$OUTFILE"
echo ']}' >> "$OUTFILE"

echo "Pulled $(jq '.tasks | length' "$OUTFILE") issues to $OUTFILE"
```

**Verification:**

```bash
cd /root/mission-control
source .env
bash scripts/pull-issues.sh
# Expected: "Pulled 5 issues to /root/mission-control/state/tasks.json"
cat state/tasks.json | jq '.tasks[0] | {number, title, labels}'
```

**Done when:** `tasks.json` содержит открытые issues из репо.

---

## ⚡ Task 2: HTML дашборд (1 файл)

**Objective:** Минимальная веб-страница, которая показывает статус задач

**Создать:** `/root/mission-control/public/index.html`

```html
<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Mission Control</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: 'SF Mono', 'Fira Code', monospace; background: #0d1117; color: #c9d1d9; padding: 20px; }
    h1 { font-size: 18px; color: #58a6ff; margin-bottom: 20px; }
    .task { display: flex; align-items: center; padding: 12px 16px; border: 1px solid #30363d; border-radius: 6px; margin-bottom: 8px; }
    .task:hover { border-color: #58a6ff; }
    .status { width: 10px; height: 10px; border-radius: 50%; margin-right: 12px; flex-shrink: 0; }
    .status.open { background: #3fb950; }
    .status.running { background: #d29922; animation: pulse 1s infinite; }
    .status.done { background: #58a6ff; }
    @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.3; } }
    .task-info { flex: 1; }
    .task-title { font-size: 14px; }
    .task-meta { font-size: 12px; color: #8b949e; margin-top: 4px; }
    .task-actions button { background: #21262d; border: 1px solid #30363d; color: #c9d1d9; padding: 4px 12px; border-radius: 6px; cursor: pointer; font-size: 12px; margin-left: 8px; }
    .task-actions button:hover { background: #30363d; }
    .terminal { background: #161b22; border: 1px solid #30363d; border-radius: 6px; padding: 12px; margin-top: 20px; min-height: 200px; font-size: 13px; white-space: pre-wrap; overflow-y: auto; max-height: 400px; }
    .terminal-header { font-size: 12px; color: #8b949e; margin-bottom: 8px; padding-bottom: 8px; border-bottom: 1px solid #30363d; }
    .empty { color: #484f58; text-align: center; padding: 60px 0; }
    .label { display: inline-block; padding: 1px 6px; border-radius: 10px; font-size: 11px; margin-left: 6px; }
    .label.enhancement { background: #a2eeef33; color: #a2eeef; }
    .label.bug { background: #d73a4a33; color: #f85149; }
    .label.refactor { background: #ededed33; color: #d2a8ff; }
    button.run { background: #238636; border-color: #238636; }
    button.run:hover { background: #2ea043; }
    .refresh { position: fixed; top: 20px; right: 20px; background: #21262d; border: 1px solid #30363d; color: #8b949e; padding: 6px 12px; border-radius: 6px; cursor: pointer; }
  </style>
</head>
<body>
  <button class="refresh" onclick="loadTasks()">⟳ Refresh</button>
  <h1>🛸 Mission Control</h1>
  <div id="tasks"></div>
  <div id="terminal" class="terminal" style="display:none">
    <div class="terminal-header">Terminal: <span id="termName"></span></div>
    <div id="termOutput"></div>
  </div>

  <script>
    async function loadTasks() {
      try {
        const resp = await fetch('/state/tasks.json');
        const data = await resp.json();
        renderTasks(data.tasks || []);
      } catch (e) {
        document.getElementById('tasks').innerHTML = 
          '<div class="empty">No tasks loaded. Run pull-issues.sh first.</div>';
      }
    }

    function renderTasks(tasks) {
      const el = document.getElementById('tasks');
      if (!tasks.length) {
        el.innerHTML = '<div class="empty">✅ All tasks done!</div>';
        return;
      }
      el.innerHTML = tasks.map(t => `
        <div class="task">
          <div class="status ${t.state === 'open' ? 'open' : 'done'}"></div>
          <div class="task-info">
            <div class="task-title">
              #${t.number} ${t.title}
              ${(t.labels||[]).map(l => `<span class="label ${l}">${l}</span>`).join('')}
            </div>
            <div class="task-meta">${t.state} · created ${new Date(t.created).toLocaleDateString()}</div>
          </div>
          <div class="task-actions">
            <button class="run" onclick="runTask(${t.number})">▶ Run</button>
            <a href="${t.url}" target="_blank"><button>GH</button></a>
          </div>
        </div>
      `).join('');
    }

    function runTask(num) {
      fetch('/api/run-task', { method: 'POST', body: JSON.stringify({number: num}) })
        .then(r => r.json())
        .then(d => {
          document.getElementById('terminal').style.display = 'block';
          document.getElementById('termName').textContent = `task-${num}`;
          document.getElementById('termOutput').textContent = d.output;
        });
    }

    loadTasks();
    setInterval(loadTasks, 10000); // refresh every 10s
  </script>
</body>
</html>
```

**Verification:**

```bash
# Serve the HTML
cd /root/mission-control/public
python3 -m http.server 3000 &
# Open http://192.168.1.131:3000
# Expected: видно список задач из tasks.json
```

**Done when:** страница открывается и показывает задачи из `state/tasks.json`.

---

## ⚡ Task 3: spawn-worker.sh — Codex в tmux

**Objective:** Запуск Codex CLI в tmux-сессии для конкретной задачи

**Создать:** `/root/mission-control/scripts/spawn-worker.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

TASK_NUM="${1:-}"
REPO="${GITHUB_REPO:-Kaiukov/my-portfolio}"
SESSION="task-${TASK_NUM}"
WORKDIR="/root/mission-control"

if [ -z "$TASK_NUM" ]; then
  echo "Usage: spawn-worker.sh <issue-number>"
  exit 1
fi

# Get issue details
ISSUE=$(gh issue view "$TASK_NUM" --repo "$REPO" --json title,body --jq '{title: .title, body: .body}')
TITLE=$(echo "$ISSUE" | jq -r '.title')
BODY=$(echo "$ISSUE" | jq -r '.body')

# Build prompt
PROMPT="GitHub Issue #${TASK_NUM}: ${TITLE}

${BODY}

Your task:
1. Read the issue above
2. Plan what needs to change
3. Implement the fix/feature
4. Test it
5. Commit with message referencing #${TASK_NUM}
6. Create a pull request"

# Kill existing session if any
tmux kill-session -t "$SESSION" 2>/dev/null || true

# Clone repo if needed
if [ ! -d "$WORKDIR/repo" ]; then
  git clone "git@github.com:${REPO}.git" "$WORKDIR/repo"
fi

# Start worker in tmux
tmux new-session -d -s "$SESSION" -x 120 -y 40 -c "$WORKDIR/repo" \
  "codex exec --model deepseek-v4 '$PROMPT' 2>&1 | tee /root/mission-control/logs/${SESSION}.log"

echo "Worker spawned: tmux attach -t $SESSION"
echo "Log: logs/${SESSION}.log"
```

**Verification:**

```bash
cd /root/mission-control
source .env
bash scripts/spawn-worker.sh 275
# Expected: "Worker spawned: tmux attach -t task-275"
sleep 3
tmux capture-pane -t task-275 -p | head -5
# Expected: output from Codex CLI
```

**Done when:** Codex запускается в tmux и видно вывод.

---

## ⚡ Task 4: WebSocket сервер для терминала

**Objective:** Простой Node.js сервер, который читает tmux вывод и шлёт в браузер

**Создать:** `/root/mission-control/scripts/ws-server.js`

```javascript
#!/usr/bin/env node
const { WebSocketServer } = require('ws');
const { execSync } = require('child_process');
const http = require('http');
const fs = require('fs');

const PORT = 3001;
const STATE_DIR = '/root/mission-control/state';

// Simple HTTP server
const server = http.createServer((req, res) => {
  if (req.url === '/api/run-task' && req.method === 'POST') {
    let body = '';
    req.on('data', c => body += c);
    req.on('end', () => {
      const { number } = JSON.parse(body);
      execSync(`bash /root/mission-control/scripts/spawn-worker.sh ${number}`, { encoding: 'utf8' });
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ ok: true, session: `task-${number}` }));
    });
  } else if (req.url === '/state/tasks.json') {
    try {
      const data = fs.readFileSync(`${STATE_DIR}/tasks.json`, 'utf8');
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(data);
    } catch {
      res.writeHead(404);
      res.end('{}');
    }
  } else {
    res.writeHead(404);
    res.end();
  }
});

// WebSocket for terminal output
const wss = new WebSocketServer({ server });

wss.on('connection', (ws) => {
  console.log('Client connected');
  
  ws.on('message', (data) => {
    const { session } = JSON.parse(data);
    console.log(`Watching tmux session: ${session}`);
    
    // Poll tmux every 500ms
    const interval = setInterval(() => {
      try {
        const output = execSync(`tmux capture-pane -t ${session} -p -S -50`, { encoding: 'utf8' });
        ws.send(JSON.stringify({ session, output }));
      } catch (e) {
        // Session ended
        ws.send(JSON.stringify({ session, output: '', done: true }));
        clearInterval(interval);
      }
    }, 500);
    
    ws.on('close', () => clearInterval(interval));
  });
});

server.listen(PORT, () => {
  console.log(`Mission Control API on :${PORT}`);
});
```

**Verification:**

```bash
cd /root/mission-control
node scripts/ws-server.js &
# Expected: "Mission Control API on :3001"

# Test with curl
curl http://localhost:3001/state/tasks.json | jq '.tasks | length'
```

**Done when:** WebSocket сервер работает и отдаёт tasks.json.

---

## ⚡ Task 5: Интеграция — дашборд + WebSocket

**Objective:** Обновить HTML дашборд, чтобы он подключался к WebSocket и показывал живой терминал

**Modify:** `/root/mission-control/public/index.html`

Добавить в `<script>` после функции `runTask`:

```javascript
let ws;
function runTask(num) {
  ws = new WebSocket('ws://localhost:3001');
  ws.onopen = () => {
    ws.send(JSON.stringify({ session: `task-${num}` }));
  };
  ws.onmessage = (evt) => {
    const { output, done } = JSON.parse(evt.data);
    const terminal = document.getElementById('terminal');
    terminal.style.display = 'block';
    document.getElementById('termName').textContent = `task-${num}${done ? ' (done)' : ''}`;
    document.getElementById('termOutput').textContent = output;
    if (done) {
      document.getElementById('termOutput').style.borderLeft = '3px solid #3fb950';
    }
  };
  // Also call API
  fetch('http://localhost:3001/api/run-task', {
    method: 'POST',
    body: JSON.stringify({ number: num })
  });
}
```

**Done when:** нажатие "Run" запускает воркера и показывает живой терминал в дашборде.

---

## ⚡ Task 6: Final — скрипт запуска всего

**Objective:** Один скрипт который поднимает всё

**Создать:** `/root/mission-control/start.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
cd /root/mission-control
source .env

echo "🚀 Mission Control starting..."

# 1. Pull latest issues
echo "→ Pulling GitHub Issues..."
bash scripts/pull-issues.sh

# 2. Start WebSocket server (background)
echo "→ Starting API server..."
node scripts/ws-server.js &
API_PID=$!
sleep 1

# 3. Start static file server for dashboard
echo "→ Starting dashboard on :3000..."
cd public
python3 -m http.server 3000 &
HTTP_PID=$!
cd ..

echo ""
echo "✅ Mission Control ready!"
echo "   Dashboard: http://localhost:3000"
echo "   API:       ws://localhost:3001"
echo ""
echo "Press Ctrl+C to stop"

# Cleanup on exit
trap "kill $API_PID $HTTP_PID 2>/dev/null; exit" INT TERM
wait
```

```bash
chmod +x /root/mission-control/start.sh
chmod +x /root/mission-control/scripts/*.sh
```

**Done when:** `bash start.sh` запускает всё одной командой.

---

## 📋 Итог Phase 1

После всех 6 задач:

```bash
bash /root/mission-control/start.sh
# → http://localhost:3000 — дашборд
# → список задач из GitHub Issues
# → кнопка Run запускает Codex в tmux
# → живой терминал в браузере
```

**Артефакты:**
```
/root/mission-control/
├── scripts/
│   ├── pull-issues.sh        # GitHub → tasks.json
│   ├── spawn-worker.sh       # Codex → tmux
│   └── ws-server.js          # WebSocket + API
├── public/
│   └── index.html            # Дашборд
├── state/
│   └── tasks.json            # Текущие задачи
├── start.sh                  # Запуск всего
└── .env                      # Secrets
```
