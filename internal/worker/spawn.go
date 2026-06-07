package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// RunningInfo represents a currently running worker session.
type RunningInfo struct {
	Number  int    `json:"number"`
	Started string `json:"started"`
	Model   string `json:"model"`
}

// Spawn launches a worker for the given issue number.
// It creates a tmux session running codex exec, and writes state/running.json.
func Spawn(issueNum int) error {
	dir := mcDir()
	repo := os.Getenv("GITHUB_REPO")
	if repo == "" {
		repo = "Kaiukov/mission-control"
	}
	model := os.Getenv("MC_MODEL")
	if model == "" {
		model = "opencode-go/deepseek-v4-flash"
	}

	session := fmt.Sprintf("task-%d", issueNum)
	repoDir := filepath.Join(dir, "repo")
	logDir := filepath.Join(dir, "logs")

	// Ensure logs directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("create logs dir: %w", err)
	}

	// Get issue details for the prompt
	issueJSON, err := exec.Command("gh", "issue", "view",
		fmt.Sprintf("%d", issueNum),
		"--repo", repo,
		"--json", "title,body,state",
	).Output()
	if err != nil {
		return fmt.Errorf("gh issue view #%d: %w (is the issue number correct?)", issueNum, err)
	}

	var issue struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		State string `json:"state"`
	}
	if err := json.Unmarshal(issueJSON, &issue); err != nil {
		return fmt.Errorf("parse issue JSON: %w", err)
	}

	prompt := fmt.Sprintf(`GitHub Issue #%d: %s

%s

Your task:
1. Read the issue above
2. Implement the fix or feature in %s
3. Write/update tests
4. Commit with message: 'fix: %s (#%d)'
5. Do NOT create a PR — the orchestrator will do that`, issueNum, issue.Title, issue.Body, repoDir, issue.Title, issueNum)

	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", session))

	// Kill existing session if any
	exec.Command("tmux", "kill-session", "-t", session).Run()

	// Spawn tmux session
	tmuxCmd := exec.Command("tmux",
		"new-session", "-d",
		"-s", session,
		"-x", "120", "-y", "40",
		"-c", repoDir,
		"bash", "-c",
		fmt.Sprintf(
			"echo '🔧 Worker #%d starting (model: %s)...'; echo ''; "+
				"codex exec --model %s '%s' 2>&1 | tee %s; "+
				"echo ''; echo '✅ Worker #%d finished'",
			issueNum, model, model, prompt, logFile, issueNum,
		),
	)
	if err := tmuxCmd.Run(); err != nil {
		return fmt.Errorf("tmux new-session: %w", err)
	}

	// Write running.json
	info := RunningInfo{
		Number:  issueNum,
		Started: time.Now().Format(time.RFC3339),
		Model:   model,
	}
	stateDir := filepath.Join(dir, "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal running info: %w", err)
	}
	runningFile := filepath.Join(stateDir, "running.json")
	if err := os.WriteFile(runningFile, data, 0644); err != nil {
		return fmt.Errorf("write running.json: %w", err)
	}

	fmt.Printf("✓ Worker spawned: tmux attach -t %s\n", session)
	fmt.Printf("  Log: logs/%s.log\n", session)
	return nil
}

func mcDir() string {
	if d := os.Getenv("MC_DIR"); d != "" {
		return d
	}
	return "."
}
