package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kaiukov/mission-control/internal/github"
	"github.com/Kaiukov/mission-control/internal/worker"
	"github.com/charmbracelet/bubbletea"
)

// Model holds the TUI state.
type Model struct {
	tasks    []github.Issue
	running  *worker.RunningInfo
	workerID int        // which issue has a running worker
	width    int
	height   int
	err      error
	lastMsg  string      // feedback to user
	logTail  string      // last N lines of worker log
}

// --- Messages ---

type tasksLoadedMsg struct {
	tasks []github.Issue
	err   error
}

type workerSpawnedMsg struct {
	issueNum int
	err      error
}

type runningCheckedMsg struct {
	info *worker.RunningInfo
}

type logPollMsg struct {
	text string
}

// --- Commands ---

func loadTasksCmd() tea.Cmd {
	return func() tea.Msg {
		dir := mcDir()
		path := filepath.Join(dir, "state", "tasks.json")
		data, err := os.ReadFile(path)
		if err != nil {
			return tasksLoadedMsg{err: fmt.Errorf("no tasks — run 'mc pull' first")}
		}
		var tl github.TaskList
		if err := json.Unmarshal(data, &tl); err != nil {
			return tasksLoadedMsg{err: fmt.Errorf("bad tasks.json")}
		}
		return tasksLoadedMsg{tasks: tl.Tasks}
	}
}

func spawnWorkerCmd(issueNum int) tea.Cmd {
	return func() tea.Msg {
		err := worker.Spawn(issueNum)
		return workerSpawnedMsg{issueNum: issueNum, err: err}
	}
}

func checkRunningCmd() tea.Cmd {
	return func() tea.Msg {
		dir := mcDir()
		path := filepath.Join(dir, "state", "running.json")
		data, err := os.ReadFile(path)
		if err != nil {
			return runningCheckedMsg{info: nil}
		}
		var info worker.RunningInfo
		if err := json.Unmarshal(data, &info); err != nil {
			return runningCheckedMsg{info: nil}
		}
		return runningCheckedMsg{info: &info}
	}
}

// pollLogCmd returns a command that reads the last N lines of the worker log.
func pollLogCmd(taskNum int) tea.Cmd {
	return func() tea.Msg {
		dir := mcDir()
		logPath := filepath.Join(dir, "logs", fmt.Sprintf("task-%d.log", taskNum))
		data, err := os.ReadFile(logPath)
		if err != nil {
			return logPollMsg{text: "(worker starting...)"}
		}
		lines := strings.Split(string(data), "\n")
		start := len(lines) - 25
		if start < 0 {
			start = 0
		}
		return logPollMsg{text: strings.Join(lines[start:], "\n")}
	}
}

// logTicker returns a command that polls the log every 1 second.
func logTicker(taskNum int) tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		dir := mcDir()
		logPath := filepath.Join(dir, "logs", fmt.Sprintf("task-%d.log", taskNum))
		data, err := os.ReadFile(logPath)
		if err != nil {
			return logPollMsg{text: "(worker starting...)"}
		}
		lines := strings.Split(string(data), "\n")
		start := len(lines) - 25
		if start < 0 {
			start = 0
		}
		return logPollMsg{text: strings.Join(lines[start:], "\n")}
	})
}

func mcDir() string {
	if d := os.Getenv("MC_DIR"); d != "" {
		return d
	}
	// Find from executable location
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	// Try to find go module root
	cmd := exec.Command("sh", "-c", "cd \"$(dirname \"$0\")\" && git rev-parse --show-toplevel 2>/dev/null || pwd")
	cmd.Dir = "."
	if out, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(out))
	}
	return "."
}

// --- Model ---

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadTasksCmd(), checkRunningCmd())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case "r":
			m.lastMsg = "Refreshing..."
			return m, tea.Batch(loadTasksCmd(), checkRunningCmd())

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx < len(m.tasks) {
				// Prevent double-spawn
				if m.running != nil {
					m.lastMsg = fmt.Sprintf("Worker already running for #%d. Press 'r' to refresh.", m.running.Number)
					return m, nil
				}
				issue := m.tasks[idx]
				m.lastMsg = fmt.Sprintf("Spawning worker for #%d...", issue.Number)
				m.workerID = issue.Number
				return m, spawnWorkerCmd(issue.Number)
			}

		case "0":
			if len(m.tasks) > 9 {
				if m.running != nil {
					m.lastMsg = "Worker already running."
					return m, nil
				}
				issue := m.tasks[9]
				m.lastMsg = fmt.Sprintf("Spawning worker for #%d...", issue.Number)
				m.workerID = issue.Number
				return m, spawnWorkerCmd(issue.Number)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tasksLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.tasks = nil
		} else {
			m.err = nil
			m.tasks = msg.tasks
		}

	case workerSpawnedMsg:
		if msg.err != nil {
			m.err = fmt.Errorf("spawn #%d: %v", msg.issueNum, msg.err)
			m.lastMsg = ""
			m.workerID = 0
		} else {
			m.err = nil
			m.lastMsg = fmt.Sprintf("✓ Worker #%d spawned — watching log...", msg.issueNum)
		}
		return m, tea.Batch(checkRunningCmd(), logTicker(msg.issueNum))

	case runningCheckedMsg:
		wasRunning := m.running != nil
		m.running = msg.info
		if msg.info == nil && wasRunning {
			// Worker finished
			m.lastMsg = "✓ Worker finished!"
			m.workerID = 0
			m.logTail = ""
			return m, loadTasksCmd()
		}
		if msg.info != nil {
			m.workerID = msg.info.Number
			return m, logTicker(msg.info.Number)
		}

	case logPollMsg:
		m.logTail = msg.text
		// Keep polling if worker still running
		if m.running != nil {
			return m, logTicker(m.running.Number)
		}
	}

	return m, nil
}

// --- View ---

func (m Model) View() string {
	var b strings.Builder

	div := strings.Repeat("─", min(m.width-2, 78))
	if div == "" {
		div = "──"
	}

	// Header
	b.WriteString("\n  🛸 MISSION CONTROL")
	if m.running != nil {
		b.WriteString(fmt.Sprintf("  ⏳ #%d running (%s)", m.running.Number, m.running.Model))
	}
	b.WriteString("\n  " + div + "\n\n")

	// Feedback message
	if m.lastMsg != "" {
		b.WriteString("  " + m.lastMsg + "\n\n")
	}

	// --- Task list (top half) ---
	taskHeight := m.height/2 - 4
	if taskHeight < 3 {
		taskHeight = 3
	}

	if len(m.tasks) == 0 {
		b.WriteString("  (no issues — run 'mc pull')\n")
	} else {
		for i, t := range m.tasks {
			if i >= taskHeight {
				b.WriteString(fmt.Sprintf("  ... and %d more\n", len(m.tasks)-taskHeight))
				break
			}
			key := fmt.Sprintf("%d", i+1)
			if i >= 9 {
				key = " "
			}

			checkbox := "[ ]"
			status := ""
			if m.running != nil && m.running.Number == t.Number {
				checkbox = "[⏳]"
				status = " ← worker running"
			}

			title := t.Title
			maxLen := m.width - 30
			if maxLen < 20 {
				maxLen = 20
			}
			if len(title) > maxLen {
				title = title[:maxLen-3] + "..."
			}

			b.WriteString(fmt.Sprintf("  %s %s  #%-3d %s%s\n", key, checkbox, t.Number, title, status))
		}
	}

	// --- Worker log (bottom half) ---
	b.WriteString("\n  " + div + "\n")

	if m.logTail != "" {
		b.WriteString("  ── Worker output (last 25 lines) ──\n\n")
		for _, line := range strings.Split(m.logTail, "\n") {
			if len(line) > m.width-4 && m.width > 10 {
				line = line[:m.width-7] + "..."
			}
			b.WriteString("  " + line + "\n")
		}
	} else {
		b.WriteString("\n  How to use:\n")
		b.WriteString("    1-9 = spawn worker for that issue\n")
		b.WriteString("    r   = refresh task list\n")
		b.WriteString("    q   = quit\n")
		b.WriteString("\n  Worker output will appear here when running.\n")
	}

	// Footer
	b.WriteString("\n  " + div + "\n")
	b.WriteString("  1-9: run   r: refresh   q: quit")

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Run starts the bubbletea TUI program.
func Run() error {
	m := Model{}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
