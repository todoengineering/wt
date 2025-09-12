package worktree

import (
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open existing worktree (interactive)",
	Long: `Interactive two-step selection process using fzf:
1. Select project from available repositories
2. Select specific worktree within that project
Opens selected worktree in configured editor and creates/switches to tmux session.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("'worktree open' command not yet implemented")
	},
}