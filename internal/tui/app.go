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

type Model struct {
	tasks    []github.Issue
	running  *worker.RunningInfo
	workerID int
	width    int
	height   int
	err      error
	lastMsg  string
	logTail  string
}

type tasksLoadedMsg struct{ tasks []github.Issue; err error }
type workerSpawnedMsg struct{ issueNum int; err error }
type runningCheckedMsg struct{ info *worker.RunningInfo }
type logPollMsg struct{ text string }

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
		json.Unmarshal(data, &info)
		return runningCheckedMsg{info: &info}
	}
}

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
	exe, _ := os.Executable()
	if exe != "" {
		return filepath.Dir(exe)
	}
	cmd := exec.Command("sh", "-c", "git rev-parse --show-toplevel 2>/dev/null || pwd")
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

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
				if m.running != nil {
					m.lastMsg = fmt.Sprintf("Worker already running #%d. Press r.", m.running.Number)
					return m, nil
				}
				issue := m.tasks[idx]
				m.lastMsg = fmt.Sprintf("Spawning worker #%d...", issue.Number)
				m.workerID = issue.Number
				return m, spawnWorkerCmd(issue.Number)
			}
		case "0":
			if len(m.tasks) > 9 && m.running == nil {
				issue := m.tasks[9]
				m.lastMsg = fmt.Sprintf("Spawning worker #%d...", issue.Number)
				m.workerID = issue.Number
				return m, spawnWorkerCmd(issue.Number)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tasksLoadedMsg:
		m.tasks = msg.tasks
		m.err = msg.err
	case workerSpawnedMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.lastMsg = fmt.Sprintf("Worker #%d spawned", msg.issueNum)
		}
		return m, tea.Batch(checkRunningCmd(), logTicker(msg.issueNum))
	case runningCheckedMsg:
		was := m.running != nil
		m.running = msg.info
		if msg.info == nil && was {
			m.lastMsg = "Worker finished!"
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
		if m.running != nil {
			return m, logTicker(m.running.Number)
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	div := strings.Repeat("─", 60)

	b.WriteString("\n  🛸 MISSION CONTROL")
	if m.running != nil {
		b.WriteString(fmt.Sprintf("  ⏳ #%d (%s)", m.running.Number, m.running.Model))
	}
	b.WriteString("\n  " + div + "\n\n")

	if m.lastMsg != "" {
		b.WriteString("  " + m.lastMsg + "\n\n")
	}

	if len(m.tasks) == 0 {
		b.WriteString("  (no issues)\n")
	} else {
		for i, t := range m.tasks {
			key := fmt.Sprintf("%d", i+1)
			if i >= 9 {
				key = " "
			}
			cb := "[ ]"
			status := ""
			if m.running != nil && m.running.Number == t.Number {
				cb = "[⏳]"
				status = " ← running"
			}
			title := t.Title
			if len(title) > 50 {
				title = title[:47] + "..."
			}
			b.WriteString(fmt.Sprintf("  %s %s  #%-3d %s%s\n", key, cb, t.Number, title, status))
		}
	}

	b.WriteString("\n  " + div + "\n")

	if m.logTail != "" {
		b.WriteString("  ── Worker output ──\n\n")
		for _, line := range strings.Split(m.logTail, "\n") {
			b.WriteString("  " + line + "\n")
		}
	} else {
		b.WriteString("\n  HOW TO USE\n")
		b.WriteString("    1-9 = spawn worker for that issue\n")
		b.WriteString("    r   = refresh\n")
		b.WriteString("    q   = quit\n\n")
		b.WriteString("  Worker output appears here.\n")
	}

	b.WriteString("\n  " + div + "\n")
	b.WriteString("  1-9: run   r: refresh   q: quit")
	return b.String()
}

func Run() error {
	m := Model{}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
