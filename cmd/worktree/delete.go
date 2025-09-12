package worktree

import (
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a worktree (interactive)",
	Long: `Interactive selection of worktree to delete using fzf.
Shows worktree name and associated branch,
requires explicit confirmation,
prevents deletion of main repository worktree.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("'worktree delete' command not yet implemented")
	},
}