package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Kaiukov/mission-control/internal/github"
	"github.com/Kaiukov/mission-control/internal/tui"
	"github.com/Kaiukov/mission-control/internal/worker"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mc",
		Short: "Mission Control — autonomous development orchestration",
		Long: `Mission Control manages GitHub issues and spawns AI workers to solve them.

Commands:
  pull     Fetch GitHub issues and save to state/tasks.json
  run <N>  Spawn a worker for issue N
  tui      Launch the terminal dashboard
  status   Show currently running workers`,
		SilenceUsage: true,
	}

	rootCmd.AddCommand(pullCmd())
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(tuiCmd())
	rootCmd.AddCommand(statusCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// mcDir returns the Mission Control root directory.
// Uses MC_DIR env var if set, otherwise defaults to current directory.
func mcDir() string {
	if d := os.Getenv("MC_DIR"); d != "" {
		return d
	}
	return "."
}

// stateDir returns the state subdirectory, creating it if needed.
func stateDir() (string, error) {
	dir := filepath.Join(mcDir(), "state")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating state dir: %w", err)
	}
	return dir, nil
}

func pullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Pull GitHub issues → state/tasks.json",
		Long: `Fetch open issues from the configured GitHub repository and save
them to state/tasks.json for the orchestrator and TUI to consume.

Uses GITHUB_REPO env var (default: Kaiukov/mission-control).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			repo := os.Getenv("GITHUB_REPO")
			if repo == "" {
				repo = "Kaiukov/mission-control"
			}

			tasks, err := github.Pull(repo)
			if err != nil {
				return fmt.Errorf("pull: %w", err)
			}

			sd, err := stateDir()
			if err != nil {
				return err
			}
			outfile := filepath.Join(sd, "tasks.json")

			data, err := json.MarshalIndent(tasks, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal tasks: %w", err)
			}
			if err := os.WriteFile(outfile, data, 0644); err != nil {
				return fmt.Errorf("write tasks.json: %w", err)
			}

			fmt.Printf("✓ Pulled %d issues from %s → %s\n", len(tasks.Tasks), repo, outfile)
			return nil
		},
	}
}

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <issue-number>",
		Short: "Spawn a worker for an issue",
		Long: `Spawn an AI worker (via codex exec in tmux) to work on the given issue number.
The worker checks out a branch, reads the issue, and starts coding.

Example:
  mc run 1    # spawn worker for issue #1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			num, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid issue number: %s", args[0])
			}
			if num < 1 {
				return fmt.Errorf("issue number must be positive, got %d", num)
			}

			fmt.Printf("🚀 Spawning worker for issue #%d...\n", num)
			return worker.Spawn(num)
		},
	}
}

func tuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch terminal dashboard",
		Long: `Launch the interactive Mission Control terminal dashboard.

Keys:
  1-9   spawn worker for the corresponding task
  r     refresh task list from GitHub
  q     quit`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run()
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show running workers",
		Long:  "Read state/running.json and display any currently active worker sessions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			sd, err := stateDir()
			if err != nil {
				return err
			}
			runningFile := filepath.Join(sd, "running.json")

			data, err := os.ReadFile(runningFile)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No workers running")
					return nil
				}
				return fmt.Errorf("read running.json: %w", err)
			}

			var info worker.RunningInfo
			if err := json.Unmarshal(data, &info); err != nil {
				return fmt.Errorf("parse running.json: %w", err)
			}

			fmt.Printf("⏳ Worker #%d running\n", info.Number)
			fmt.Printf("   Model:   %s\n", info.Model)
			fmt.Printf("   Started: %s\n", info.Started)
			fmt.Printf("   Session: tmux attach -t task-%d\n", info.Number)
			return nil
		},
	}
}
