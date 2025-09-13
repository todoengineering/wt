package worktree

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dboroujerdi/wt/internal/editor"
	"github.com/dboroujerdi/wt/internal/git"
	"github.com/dboroujerdi/wt/internal/tmux"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [worktree-name]",
	Short: "Switch to a worktree",
	Long: `Switch to a worktree by name or interactively select one.
Opens the worktree in your editor and switches to or creates a tmux session.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !git.IsGitRepository() {
			fmt.Fprintf(os.Stderr, "Error: not in a git repository\n")
			os.Exit(1)
		}

		repoName, err := git.GetRepositoryName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var selectedWorktree git.Worktree

		if len(args) == 0 {
			// Interactive selection
			worktrees, err := git.ListWorktrees(repoName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing worktrees: %v\n", err)
				os.Exit(1)
			}

			if len(worktrees) == 0 {
				fmt.Printf("No worktrees found for repository '%s'\n", repoName)
				fmt.Printf("Use 'wt new' to create a worktree\n")
				os.Exit(1)
			}

			selected, err := selectWorktreeInteractive(worktrees)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error selecting worktree: %v\n", err)
				os.Exit(1)
			}
			selectedWorktree = selected
		} else {
			// Direct selection by name
			worktreeName := args[0]
			worktrees, err := git.ListWorktrees(repoName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing worktrees: %v\n", err)
				os.Exit(1)
			}

			found := false
			for _, wt := range worktrees {
				if wt.Name == worktreeName {
					selectedWorktree = wt
					found = true
					break
				}
			}

			if !found {
				fmt.Fprintf(os.Stderr, "Error: worktree '%s' not found\n", worktreeName)
				fmt.Printf("Available worktrees:\n")
				for _, wt := range worktrees {
					fmt.Printf("  %s\n", wt.Name)
				}
				os.Exit(1)
			}
		}

		// Switch to the worktree
		fmt.Printf("Switching to worktree: %s\n", selectedWorktree.Name)

		// Create or switch to tmux session
		sessionName := fmt.Sprintf("%s-%s", repoName, selectedWorktree.Name)
		sessionName = tmux.SanitizeSessionName(sessionName)

		if tmux.IsInstalled() {
			if tmux.SessionExists(sessionName) {
				fmt.Printf("Switching to existing tmux session: %s\n", sessionName)
				// Open editor in the existing session before switching
				if err := tmux.SendCommandToSession(sessionName, editor.GetEditorCommand(selectedWorktree.Path)); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to open editor in session: %v\n", err)
				}
				if err := tmux.SwitchToSession(sessionName); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to switch tmux session: %v\n", err)
				}
			} else {
				fmt.Printf("Creating new tmux session: %s\n", sessionName)
				if err := tmux.CreateSessionWithCommand(sessionName, selectedWorktree.Path, editor.GetEditorCommand(selectedWorktree.Path)); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to create tmux session: %v\n", err)
				}
			}
		} else {
			// No tmux, just open editor normally
			if err := editor.OpenInEditor(selectedWorktree.Path); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to open editor: %v\n", err)
			}
		}
	},
}

func selectWorktreeInteractive(worktrees []git.Worktree) (git.Worktree, error) {
	// Check if fzf is available
	if _, err := exec.LookPath("fzf"); err == nil {
		return selectWorktreeWithFzf(worktrees)
	}

	// Fallback to simple selection
	fmt.Println("Select a worktree:")
	for i, wt := range worktrees {
		fmt.Printf("%d) %s (%s)\n", i+1, wt.Name, wt.Path)
	}

	var choice int
	fmt.Print("Enter choice (number): ")
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		return git.Worktree{}, fmt.Errorf("invalid input: %w", err)
	}

	if choice < 1 || choice > len(worktrees) {
		return git.Worktree{}, fmt.Errorf("invalid choice: %d", choice)
	}

	return worktrees[choice-1], nil
}

func selectWorktreeWithFzf(worktrees []git.Worktree) (git.Worktree, error) {
	// Prepare input for fzf
	var input bytes.Buffer
	worktreeMap := make(map[string]git.Worktree)
	for _, wt := range worktrees {
		display := fmt.Sprintf("%s\t%s", wt.Name, wt.Path)
		input.WriteString(display + "\n")
		worktreeMap[display] = wt
	}
	
	// Create fzf command
	cmd := exec.Command("fzf", "--prompt=Select worktree: ", "--height=40%", "--layout=reverse", "--with-nth=1", "--delimiter=\t")
	cmd.Stdin = &input
	cmd.Stderr = os.Stderr
	
	// Run fzf and capture output
	output, err := cmd.Output()
	if err != nil {
		// User cancelled or error
		return git.Worktree{}, fmt.Errorf("selection cancelled")
	}
	
	// Parse the selected line
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return git.Worktree{}, fmt.Errorf("no selection made")
	}
	
	// Find the matching worktree
	if wt, ok := worktreeMap[selected]; ok {
		return wt, nil
	}
	
	// Fallback: try to match by name only
	selectedName := strings.Split(selected, "\t")[0]
	for _, wt := range worktrees {
		if wt.Name == selectedName {
			return wt, nil
		}
	}
	
	return git.Worktree{}, fmt.Errorf("could not find selected worktree")
}