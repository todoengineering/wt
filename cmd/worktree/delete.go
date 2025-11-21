package worktree

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/todoengineering/wt/internal/git"
	"github.com/todoengineering/wt/internal/tmux"
	"github.com/todoengineering/wt/internal/ui"
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
	var items []ui.Item
	for _, wt := range worktrees {
		items = append(items, ui.Item{
			TitleStr:       wt.Name,
			DescriptionStr: fmt.Sprintf("[%s] %s", wt.Branch, wt.Path),
			FilterStr:      wt.Name,
			Value:          wt,
		})
	}

	selected, err := ui.Select(items, "Select a worktree to delete")
	if err != nil {
		return git.Worktree{}, err
	}

	return selected.Value.(git.Worktree), nil
}
