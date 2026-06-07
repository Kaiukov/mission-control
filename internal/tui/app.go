package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kaiukov/mission-control/internal/github"
	"github.com/Kaiukov/mission-control/internal/worker"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model holds the TUI state.
type Model struct {
	tasks   []github.Issue
	running *worker.RunningInfo
	width   int
	height  int
	err     error
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

// --- Commands ---

func loadTasksCmd() tea.Cmd {
	return func() tea.Msg {
		dir := mcDir()
		path := filepath.Join(dir, "state", "tasks.json")
		data, err := os.ReadFile(path)
		if err != nil {
			return tasksLoadedMsg{err: fmt.Errorf("read tasks.json: %w", err)}
		}
		var tl github.TaskList
		if err := json.Unmarshal(data, &tl); err != nil {
			return tasksLoadedMsg{err: fmt.Errorf("parse tasks.json: %w", err)}
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

// mcDir returns the Mission Control root directory.
func mcDir() string {
	if d := os.Getenv("MC_DIR"); d != "" {
		return d
	}
	return "."
}

// --- Model ---

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(loadTasksCmd(), checkRunningCmd())
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			return m, tea.Batch(loadTasksCmd(), checkRunningCmd())

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1') // '1' → 0, '9' → 8
			if idx < len(m.tasks) {
				issue := m.tasks[idx]
				return m, spawnWorkerCmd(issue.Number)
			}

		case "0":
			// index 9 (the 10th task) if we want to support it
			if len(m.tasks) > 9 {
				issue := m.tasks[9]
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
			m.err = fmt.Errorf("spawn #%d: %w", msg.issueNum, msg.err)
		} else {
			m.err = nil
		}
		return m, checkRunningCmd()

	case runningCheckedMsg:
		m.running = msg.info
	}

	return m, nil
}

// --- View ---

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m Model) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🛸 MISSION CONTROL"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	// Error
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  ✗ %s", m.err.Error())))
		b.WriteString("\n\n")
	}

	// Task list
	if len(m.tasks) == 0 {
		b.WriteString("  (no issues loaded — press 'r' to refresh or run 'mc pull')\n")
	} else {
		for i, t := range m.tasks {
			key := fmt.Sprintf("%d", i+1)
			if i >= 9 {
				key = " "
			}
			checkbox := "[ ]"
			if m.running != nil && m.running.Number == t.Number {
				checkbox = "[⏳]"
			}
			line := fmt.Sprintf("  %s %s  #%-3d %s", key, checkbox, t.Number, truncate(t.Title, m.width-20))
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Divider
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n")

	// Help
	b.WriteString(helpStyle.Render("  1-9: spawn  r: refresh  q: quit"))
	b.WriteString("\n")

	// Running status
	if m.running != nil {
		b.WriteString(statusStyle.Render(fmt.Sprintf("  ⏳ worker #%d running (%s)", m.running.Number, m.running.Model)))
		b.WriteString("\n")
	} else {
		b.WriteString(helpStyle.Render("  no workers running"))
		b.WriteString("\n")
	}

	// Bottom border
	b.WriteString(borderStyle.Render(strings.Repeat("─", 50)))

	return b.String()
}

func truncate(s string, maxLen int) string {
	if maxLen < 4 {
		return s[:maxLen]
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Run starts the bubbletea TUI program.
func Run() error {
	// Verify state directory exists
	sd := filepath.Join(mcDir(), "state")
	if _, err := os.Stat(sd); os.IsNotExist(err) {
		return fmt.Errorf("state directory not found at %s — run 'mc pull' first", sd)
	}

	m := Model{}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}
