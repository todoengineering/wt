package worktree

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wt",
	Short: "A CLI tool for managing Git worktrees",
	Long: `wt is a Go-based CLI tool for managing Git worktrees 
with enhanced features including automatic tmux session management 
and editor integration.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(branchCmd)
	rootCmd.AddCommand(deleteCmd)
}