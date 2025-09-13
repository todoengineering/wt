package worktree

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/todoengineering/wt/internal/editor"
	"github.com/todoengineering/wt/internal/git"
	"github.com/todoengineering/wt/internal/tmux"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout [branch]",
	Short: "Create worktree from existing branch",
	Long: `Lists all available branches (local and remote) using fzf,
creates worktree for selected branch,
handles special characters in branch names,
detects and opens existing worktrees instead of creating duplicates.`,
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

		// Fetch remote branches first
		fmt.Println("Fetching remote branches...")
		if err := git.FetchRemoteBranches(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch remote branches: %v\n", err)
		}

		var selectedBranch string

		if len(args) > 0 {
			// Branch name provided as argument
			selectedBranch = args[0]
		} else {
			// Interactive selection
			branches, err := git.ListAllBranches()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing branches: %v\n", err)
				os.Exit(1)
			}

			// Filter out current branch
			currentBranch, _ := git.GetCurrentBranch()
			var selectableBranches []git.Branch
			for _, branch := range branches {
				if branch.Name != currentBranch {
					selectableBranches = append(selectableBranches, branch)
				}
			}

			if len(selectableBranches) == 0 {
				fmt.Println("No other branches available")
				os.Exit(1)
			}

			selectedBranch, err = selectBranchInteractive(selectableBranches)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error selecting branch: %v\n", err)
				os.Exit(1)
			}
		}

		// Check if worktree already exists for this branch
		exists, existingWorktree := git.WorktreeExistsForBranch(repoName, selectedBranch)
		if exists {
			fmt.Printf("A worktree already exists for branch '%s' at:\n", selectedBranch)
			fmt.Printf("  %s\n\n", existingWorktree.Path)
			fmt.Println("Would you like to switch to it? (y/n)")
			
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
				// Switch to existing worktree
				switchToWorktree(repoName, *existingWorktree)
			} else {
				fmt.Println("Cancelled")
			}
			return
		}

		// Create new worktree for the branch
		sanitizedName := git.SanitizeBranchName(selectedBranch)
		fmt.Printf("Creating worktree '%s' for branch '%s'...\n", sanitizedName, selectedBranch)
		
		worktreePath, err := git.CreateWorktree(repoName, sanitizedName, selectedBranch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating worktree: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Worktree created at: %s\n", worktreePath)

		// Create/switch tmux session with editor
		sessionName := fmt.Sprintf("%s-%s", repoName, sanitizedName)
		sessionName = tmux.SanitizeSessionName(sessionName)
		
		if tmux.IsInstalled() {
			if err := tmux.CreateSessionWithCommand(sessionName, worktreePath, editor.GetEditorCommand(worktreePath)); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			} else {
				fmt.Printf("Tmux session '%s' created with editor\n", sessionName)
			}
		} else {
			// No tmux, just open editor normally
			if err := editor.OpenInEditor(worktreePath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			} else {
				fmt.Printf("Opened in editor\n")
			}
		}
	},
}

func selectBranchInteractive(branches []git.Branch) (string, error) {
	// Check if fzf is available
	if _, err := exec.LookPath("fzf"); err == nil {
		return selectBranchWithFzf(branches)
	}

	// Fallback to simple selection
	fmt.Println("Select a branch:")
	for i, branch := range branches {
		status := ""
		if branch.IsLocal && branch.IsRemote {
			status = " [local+remote]"
		} else if branch.IsLocal {
			status = " [local]"
		} else {
			status = " [remote]"
		}
		fmt.Printf("%d) %s%s\n", i+1, branch.Name, status)
	}

	var choice int
	fmt.Print("Enter choice (number): ")
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if choice < 1 || choice > len(branches) {
		return "", fmt.Errorf("invalid choice: %d", choice)
	}

	return branches[choice-1].Name, nil
}

func selectBranchWithFzf(branches []git.Branch) (string, error) {
	// Sort branches: local first, then by name
	sort.Slice(branches, func(i, j int) bool {
		if branches[i].IsLocal != branches[j].IsLocal {
			return branches[i].IsLocal
		}
		return branches[i].Name < branches[j].Name
	})

	// Prepare input for fzf
	var input bytes.Buffer
	for _, branch := range branches {
		status := ""
		if branch.IsLocal && branch.IsRemote {
			status = "[local+remote]"
		} else if branch.IsLocal {
			status = "[local]"
		} else {
			status = "[remote]"
		}
		line := fmt.Sprintf("%-40s %s", branch.Name, status)
		input.WriteString(line + "\n")
	}

	// Create fzf command
	cmd := exec.Command("fzf", 
		"--prompt=Select branch: ",
		"--height=40%",
		"--layout=reverse",
		"--header=Select a branch to checkout as worktree")
	cmd.Stdin = &input
	cmd.Stderr = os.Stderr

	// Run fzf and capture output
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("selection cancelled")
	}

	// Parse the selected line to get branch name
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return "", fmt.Errorf("no selection made")
	}

	// Extract branch name (first field)
	parts := strings.Fields(selected)
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", fmt.Errorf("could not parse selection")
}

func switchToWorktree(repoName string, worktree git.Worktree) {
	fmt.Printf("Switching to worktree: %s\n", worktree.Name)

	// Create or switch to tmux session
	sessionName := fmt.Sprintf("%s-%s", repoName, worktree.Name)
	sessionName = tmux.SanitizeSessionName(sessionName)

	if tmux.IsInstalled() {
		if tmux.SessionExists(sessionName) {
			// Open editor in the existing session before switching
			if err := tmux.SendCommandToSession(sessionName, editor.GetEditorCommand(worktree.Path)); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to open editor in session: %v\n", err)
			}
			if err := tmux.SwitchToSession(sessionName); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to switch tmux session: %v\n", err)
			} else {
				fmt.Printf("Switched to tmux session: %s\n", sessionName)
			}
		} else {
			if err := tmux.CreateSessionWithCommand(sessionName, worktree.Path, editor.GetEditorCommand(worktree.Path)); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create tmux session: %v\n", err)
			} else {
				fmt.Printf("Created new tmux session: %s\n", sessionName)
			}
		}
	} else {
		// No tmux, just open editor
		if err := editor.OpenInEditor(worktree.Path); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to open editor: %v\n", err)
		}
	}
}