package worktree

import (
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Create worktree from existing branch",
	Long: `Lists all available branches (local and remote) using fzf,
creates worktree for selected branch,
handles special characters in branch names,
detects and opens existing worktrees instead of creating duplicates.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("'worktree checkout' command not yet implemented")
	},
}