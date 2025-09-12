package worktree

import (
	"fmt"
	"os"

	"github.com/dboroujerdi/wt/internal/git"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List worktrees for current repository",
	Long:  `Lists all on-disk worktrees for the current repository from the standard directory structure.`,
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
		
		worktrees, err := git.ListWorktrees(repoName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing worktrees: %v\n", err)
			os.Exit(1)
		}
		
		if len(worktrees) == 0 {
			fmt.Printf("No worktrees found for repository '%s'\n", repoName)
			fmt.Printf("Worktree directory: %s\n", git.GetWorktreeDir(repoName))
			return
		}
		
		fmt.Printf("Worktrees for repository '%s':\n", repoName)
		for _, wt := range worktrees {
			fmt.Printf("  %s -> %s\n", wt.Name, wt.Path)
		}
	},
}