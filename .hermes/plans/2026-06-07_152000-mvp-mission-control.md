# Mission Control MVP — Minimum Working Version

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** TUI where you press a key → Codex works on that issue → you see output → you know when done.

**Architecture:** 3 bash scripts + tmux TUI. Pull issues from GitHub, spawn Codex workers in tmux, track status in tasks.json. Single machine (CT 113).

**Tech Stack:** bash, tmux, jq, gh CLI, codex CLI. No Node, no Python, no Docker.

**Test environment:** `root@hermes` (CT 113, 100.91.176.74), repo `Kaiukov/mission-control`.

---

## Current State (what exists)

```
/root/mission-control/
├── start.sh              ← launch (pulls issues + starts TUI)
├── scripts/
│   ├── pull-issues.sh    ← GitHub → tasks.json ✅ works
│   ├── spawn-worker.sh   ← Codex in tmux ✅ works (tested: issue #1)
│   └── tui.sh            ← tmux dashboard ⚠️ needs fixes
├── state/tasks.json      ← 1 issue (#1)
└── .env                  ← GITHUB_REPO, MC_DIR
```

**Works:** pull-issues.sh pulls real issues. spawn-worker.sh clones repo, creates branch, launches codex in tmux.

**Broken:**
1. TUI hotkeys (`r`, `q`) don't fire because top pane (watch) captures input
2. `r` hardcoded to issue #1 — no way to select which issue
3. Worker output not visible from TUI — user must manually `tmux attach -t task-1`
4. No status tracking — TUI doesn't know if worker is running/done
5. `spawn-worker.sh` reads issue `--json title,body,state` but state isn't in `tasks.json` format

---

## What MVP Must Do

1. **Launch:** `bash start.sh` → TUI opens, shows open issues
2. **Select & Run:** Press `1`-`9` key → spawn Codex for that issue
3. **See output:** Bottom pane shows live worker terminal
4. **Know when done:** Worker finishes → TUI shows `[✓]` instead of `[ ]`
5. **Quit:** `q` exits

---

## Task 1: Fix TUI — bottom pane active by default

**Objective:** User lands in bottom pane so hotkeys work immediately.

**File:** `scripts/tui.sh`

**Problem:** top pane runs `watch` which captures stdin. `bind-key -n` binds to root table but top pane's shell owns the TTY.

**Fix:** Remove `tmux select-pane -U` at end. Bottom pane should be active by default. Update help text.

```bash
# Line 68: remove this
# tmux select-pane -t "$SESSION" -U

# Replace with: leave bottom pane active
tmux select-pane -t "$SESSION" -D  # bottom pane active
```

Update help:
```
  1.  Press 1-9 to run that issue
  2.  Worker output shows below
  3.  Press q to quit
```

**Verification:** `bash start.sh` → immediately see `▸ HOW TO USE` with cursor. Press `q` → exits. Press `1` → no error (not wired yet).

---

## Task 2: Add numeric hotkeys (1-9) for each issue

**Objective:** Press `1` → runs spawn-worker.sh for issue #1. Press `2` → issue #2. Etc.

**File:** `scripts/tui.sh`

**Add after line 71:**

```bash
# Bind numeric keys 1-9 to spawn worker for corresponding issue
for i in $(seq 1 9); do
  tmux bind-key -n "$i" "run-shell 'bash $MC_DIR/scripts/spawn-worker.sh $i'"
done
```

Also update the `r` key to prompt for issue number:

```bash
# r = prompt for issue number
tmux bind-key -n r "command-prompt -p 'issue number:' 'run-shell \"bash $MC_DIR/scripts/spawn-worker.sh %%\"'"
```

**Update help text** to show: `1-9 = run issue, r = run with prompt`

**Verification:** Press `1` → spawns worker for #1. Press `r` → prompts for number → enter `1` → spawns worker.

---

## Task 3: Show worker output in bottom pane

**Objective:** When worker spawns, bottom pane shows its terminal output live.

**File:** `scripts/spawn-worker.sh`

**Current problem:** worker spawns in SEPARATE tmux session (`task-1`). User can't see it from TUI.

**Fix:** Instead of new tmux session, spawn worker in a new WINDOW of the same session. Or: use `tmux split-window` to create a third pane, or use `tmux new-window`.

**Better approach:** After spawning, attach bottom pane to worker's tmux session:

```bash
# In spawn-worker.sh, after spawning worker:
tmux select-pane -t mc.1
tmux send-keys -t mc.1 C-c  # clear any running command
sleep 0.2

# Pipe worker log to bottom pane
WATCH_CMD="tail -f $MC_DIR/logs/task-${TASK_NUM}.log"
tmux send-keys -t mc.1 "$WATCH_CMD" Enter
```

**Alternative (cleaner):** Use `tmux pipe-pane` or `tmux split-window` inside the mc session to create a worker pane below the help pane.

**Simplest working version:**

```bash
# After spawning worker, update bottom pane
tmux send-keys -t mc.1 C-c Enter  # kill current command
sleep 0.2
tmux send-keys -t mc.1 "clear && echo '🔧 Worker #${TASK_NUM} — $(date +%H:%M)' && echo '' && tail -f $MC_DIR/logs/task-${TASK_NUM}.log" Enter
```

**Verification:** Press `1` → bottom pane shows `🔧 Worker #1` then live Codex output.

---

## Task 4: Track worker status in tasks.json

**Objective:** When worker starts → TUI shows `⏳`. When done → TUI shows `✓`.

**File:** `scripts/spawn-worker.sh` (add at end), `scripts/tui.sh` (update watch script)

**Step 1: Update tasks.json on spawn:**

Add to `spawn-worker.sh` after spawning worker:

```bash
# Mark as running in tasks.json
RUNNING_FILE="$MC_DIR/state/running.json"
echo '{"number":'"$TASK_NUM"',"started":"'"$(date -Iseconds)"'","model":"'"$MODEL"'"}' > "$RUNNING_FILE"
```

**Step 2: Check running status in watch script:**

Update the watch script in `tui.sh` to check `running.json`:

```bash
# In the watch loop, after reading tasks.json:
RUNNING_NUM=""
[ -f "$MC_DIR/state/running.json" ] && RUNNING_NUM=$(jq -r '.number' "$MC_DIR/state/running.json" 2>/dev/null || echo "")

# Then in the task display:
jq -r --arg rn "$RUNNING_NUM" '.tasks[] | 
  if (.number | tostring) == $rn then "  [⏳]  #\(.number)  \(.title)"
  else "  [ ]  #\(.number)  \(.title)"
  end' "$MC_DIR/state/tasks.json"
```

**Step 3: Mark done (worker wrapper):**

Create `scripts/worker-wrapper.sh` that spawn-worker calls, which:
1. Runs codex
2. Removes `running.json` when done

**Verification:** Press `1` → TUI shows `[⏳] #1`. Wait for worker to finish → TUI shows `[ ] #1` again (or check log).

---

## Task 5: End-to-end test

**Objective:** Run the full flow with a real issue.

**Test scenario:**

```bash
# 1. Create test issue
gh issue create --repo Kaiukov/mission-control \
  --title "test: add hello world script" \
  --body "Create scripts/hello.sh that prints 'Hello from Mission Control'" \
  --label "enhancement"

# 2. Pull issues
bash scripts/pull-issues.sh

# 3. Start TUI (manually, involves tmux attach)
bash scripts/tui.sh

# 4. In TUI: press 2 (for issue #2)
# 5. Watch bottom pane for Codex output
# 6. Wait for "✅ Worker finished"
# 7. Check: tmux attach -t task-2 → see if hello.sh was created
```

**Acceptance criteria:**
- [ ] Issue #2 visible in TUI with `[ ]`
- [ ] Press `2` → bottom pane shows worker output
- [ ] Worker creates `scripts/hello.sh` in repo
- [ ] Worker commits with message referencing #2
- [ ] Worker does NOT create PR (should only commit)

**Spawn real worker (skip if Codex unavailable):**

```bash
# Test spawn-worker standalone
bash scripts/spawn-worker.sh 2
# Wait...
tmux capture-pane -t task-2 -p | tail -20
```

---

## Task 6: Polish — help text, status bar, error handling

**Objective:** TUI looks professional, handles errors gracefully.

**Files:** `scripts/tui.sh`, `scripts/spawn-worker.sh`

**Sub-tasks:**

### 6a. Dynamic help text in bottom pane

Show which issues can be run:

```bash
# In help file: list available issues
ISSUES_COUNT=$(jq '.tasks | length' "$MC_DIR/state/tasks.json" 2>/dev/null || echo 0)
AVAILABLE=""
if [ "$ISSUES_COUNT" -gt 0 ]; then
  AVAILABLE=$(jq -r '.tasks[] | "    \(.number) = #\(.number) \(.title[:40])..."' "$MC_DIR/state/tasks.json" 2>/dev/null)
fi
```

### 6b. Error handling

- `spawn-worker.sh`: if codex not found → show "❌ codex not installed" in bottom pane
- `pull-issues.sh`: if gh not authenticated → show "❌ gh not logged in"
- `tui.sh`: if no tasks.json → show "Run pull-issues.sh first"

### 6c. Green status bar

Use tmux status bar:

```bash
tmux set -t "$SESSION" status-style "bg=green,fg=black"
tmux set -t "$SESSION" status-left "🛸 MC"
tmux set -t "$SESSION" status-right "#(cat $MC_DIR/state/running.json 2>/dev/null | jq -r 'if .number then \"⏳ #\\(.number)\" else \"idle\" end') | %H:%M"
```

---

## Files Summary

| File | Action | What changes |
|---|---|---|
| `scripts/tui.sh` | **Modify** | Bottom pane active, 1-9 keys, running status, green bar, dynamic help |
| `scripts/spawn-worker.sh` | **Modify** | Write running.json, pipe output to TUI bottom pane |
| `scripts/worker-wrapper.sh` | **Create** | Wraps codex call, removes running.json when done |
| `state/running.json` | **New file** | Tracks which worker is active |

---

## Risks & Open Questions

1. **Codex prompt escaping** — single quotes inside PROMPT variable may break. Mitigation: use heredoc or base64 encode prompt.
2. **tmux nesting** — spawning worker in same session may cause nesting issues. Mitigation: separate tmux session (current approach), show output via `tail -f log`.
3. **Multiple workers** — what if user presses `1` then `2` while `1` is running? Current: second spawn kills first (same session name). Desired: queue or warn. MVP: just warn "worker already running".
4. **Orchestrator** — Opus not involved yet. Phase 2.

---

## Verification Checklist

- [ ] `bash start.sh` → TUI opens with task list
- [ ] Press `1` → spawns worker, bottom pane shows output
- [ ] TUI shows `[⏳]` while running
- [ ] Press `q` → exits TUI
- [ ] Worker finishes → `logs/task-1.log` has output, code committed
- [ ] Second `bash start.sh` → shows updated task list from GitHub

---

## Task 7: Use deepseek-v4-flash for testing

**Objective:** Test worker uses `deepseek-v4-flash` (fastest/cheapest) instead of default `deepseek-v3`.

**File:** `scripts/spawn-worker.sh`, default model override

**Change default model for test:**

```bash
# Line 14: change default from deepseek-v3 to deepseek-v4-flash
MODEL="${2:-deepseek-v4-flash}"
```

**Also update tui.sh keybindings** to pass model:

```bash
# In tui.sh, bind with model override
tmux bind-key -n 1 "run-shell 'bash $MC_DIR/scripts/spawn-worker.sh 1 deepseek-v4-flash'"
```

**Verification:** Press `1` → spawns Codex with `--model deepseek-v4-flash`. Check `logs/task-1.log` → first line shows model.
