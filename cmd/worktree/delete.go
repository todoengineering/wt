package worktree

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/todoengineering/wt/internal/git"
	"github.com/todoengineering/wt/internal/tmux"
)

var forceDelete bool

var deleteCmd = &cobra.Command{
	Use:   "delete [worktree-name]",
	Short: "Delete a worktree (interactive)",
	Long: `Interactive selection of worktree to delete using fzf.
Shows worktree name and associated branch,
requires explicit confirmation,
prevents deletion of main repository worktree.`,
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

		// Get list of worktrees
		worktrees, err := git.ListWorktrees(repoName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing worktrees: %v\n", err)
			os.Exit(1)
		}

		if len(worktrees) == 0 {
			fmt.Println("No worktrees found")
			os.Exit(1)
		}

		var selectedWorktree git.Worktree

		if len(args) > 0 {
			// Worktree name provided as argument
			worktreeName := args[0]
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
				os.Exit(1)
			}
		} else {
			// Interactive selection
			selected, err := selectWorktreeForDeletion(worktrees)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error selecting worktree: %v\n", err)
				os.Exit(1)
			}
			selectedWorktree = selected
		}

		// Check if it's the main worktree
		isMain, err := git.IsMainWorktree(selectedWorktree.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking if main worktree: %v\n", err)
			os.Exit(1)
		}

		if isMain {
			fmt.Fprintf(os.Stderr, "Error: cannot delete the main repository worktree\n")
			os.Exit(1)
		}

		if !forceDelete {
			fmt.Printf("\n🗑️  Worktree Deletion Confirmation\n")
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("Worktree:    %s\n", selectedWorktree.Name)
			fmt.Printf("Branch:      %s\n", selectedWorktree.Branch)
			fmt.Printf("Path:        %s\n", selectedWorktree.Path)
			fmt.Printf("\n⚠️  WARNING: This will permanently remove:\n")
			fmt.Printf("   • The worktree directory and all its contents\n")
			fmt.Printf("   • Any uncommitted changes in this worktree\n")
			fmt.Printf("   • Associated tmux session (if exists)\n")
			fmt.Printf("\nType 'yes' to confirm deletion: ")

			var confirmation string
			fmt.Scanln(&confirmation)

			if strings.ToLower(confirmation) != "yes" {
				fmt.Println("✅ Deletion cancelled")
				return
			}
		}

		// Check for and kill associated tmux session
		// Standard session name is <repo>-<worktree>. For backward compatibility,
		// also try just <worktree> (older versions of wt new).
		primarySession := tmux.SanitizeSessionName(fmt.Sprintf("%s-%s", repoName, selectedWorktree.Name))
		legacySession := tmux.SanitizeSessionName(selectedWorktree.Name)

		if tmux.IsInstalled() {
			killed := false
			if tmux.SessionExists(primarySession) {
				fmt.Printf("🔄 Killing tmux session: %s\n", primarySession)
				if err := tmux.KillSession(primarySession); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to kill tmux session: %v\n", err)
				} else {
					killed = true
				}
			}
			// Try legacy session name if primary wasn't found/killed
			if !killed && tmux.SessionExists(legacySession) {
				fmt.Printf("🔄 Killing tmux session: %s\n", legacySession)
				if err := tmux.KillSession(legacySession); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to kill tmux session: %v\n", err)
				}
			}
		}

		// Delete the worktree
		fmt.Printf("🔄 Deleting worktree '%s'...\n", selectedWorktree.Name)
		if err := git.RemoveWorktree(selectedWorktree.Path); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting worktree: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Worktree '%s' has been deleted successfully\n", selectedWorktree.Name)
	},
}

func init() {
	deleteCmd.Flags().BoolVar(&forceDelete, "force", false, "skip confirmation and delete the worktree")
}

func selectWorktreeForDeletion(worktrees []git.Worktree) (git.Worktree, error) {
	// Check if fzf is available
	if _, err := exec.LookPath("fzf"); err == nil {
		return selectWorktreeWithFzfForDeletion(worktrees)
	}

	// Fallback to simple selection
	fmt.Println("Select a worktree to delete:")
	for i, wt := range worktrees {
		fmt.Printf("%d) %-30s [%s] %s\n", i+1, wt.Name, wt.Branch, wt.Path)
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

func selectWorktreeWithFzfForDeletion(worktrees []git.Worktree) (git.Worktree, error) {
	// Prepare input for fzf
	var input bytes.Buffer
	for _, wt := range worktrees {
		line := fmt.Sprintf("%-30s %-20s %s", wt.Name, fmt.Sprintf("[%s]", wt.Branch), wt.Path)
		input.WriteString(line + "\n")
	}

	// Create fzf command
	cmd := exec.Command("fzf",
		"--prompt=🗑️  Select worktree to delete: ",
		"--height=50%",
		"--layout=reverse",
		"--header=⚠️  WARNING: Selected worktree will be permanently deleted\n📁 Format: NAME [BRANCH] PATH")
	cmd.Stdin = &input
	cmd.Stderr = os.Stderr

	// Run fzf and capture output
	output, err := cmd.Output()
	if err != nil {
		return git.Worktree{}, fmt.Errorf("selection cancelled")
	}

	// Parse the selected line to get worktree name
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return git.Worktree{}, fmt.Errorf("no selection made")
	}

	// Extract worktree name (first field)
	parts := strings.Fields(selected)
	if len(parts) > 0 {
		worktreeName := parts[0]
		for _, wt := range worktrees {
			if wt.Name == worktreeName {
				return wt, nil
			}
		}
	}

	return git.Worktree{}, fmt.Errorf("could not find selected worktree")
}
