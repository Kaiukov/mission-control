package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ghLabel is the raw label object returned by gh CLI.
type ghLabel struct {
	Name string `json:"name"`
}

// Issue represents a GitHub issue.
type Issue struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	URL     string `json:"url"`
	Created string `json:"created"`

	// Raw labels from gh (objects), unmarshalled separately.
	RawLabels []ghLabel `json:"labels"`
}

// TaskList is the envelope for the tasks.json file.
type TaskList struct {
	Updated string  `json:"updated"`
	Tasks   []Issue `json:"tasks"`
}

// Pull fetches open issues from a GitHub repository using `gh issue list`.
func Pull(repo string) (*TaskList, error) {
	// Run: gh issue list --repo <repo> --state open --limit 50 --json number,title,labels,state,createdAt,url
	args := []string{
		"issue", "list",
		"--repo", repo,
		"--state", "open",
		"--limit", "50",
		"--json", "number,title,labels,state,createdAt,url",
	}

	out, err := exec.Command("gh", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("gh issue list: %w (is gh CLI installed and authenticated?)", err)
	}

	// gh outputs a JSON array directly when using --json
	var issues []Issue
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, fmt.Errorf("parse gh output: %w", err)
	}

	// Normalize: extract label names from ghLabel objects
	for i := range issues {
		issues[i].State = strings.ToUpper(issues[i].State)
	}

	return &TaskList{
		Updated: time.Now().Format(time.RFC3339),
		Tasks:   issues,
	}, nil
}

// Labels returns the label names as strings.
func (i *Issue) Labels() []string {
	names := make([]string, len(i.RawLabels))
	for j, l := range i.RawLabels {
		names[j] = l.Name
	}
	return names
}
