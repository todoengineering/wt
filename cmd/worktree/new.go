package worktree

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dboroujerdi/wt/internal/editor"
	"github.com/dboroujerdi/wt/internal/git"
	"github.com/dboroujerdi/wt/internal/tmux"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create new worktree with new branch",
	Long: `Creates a new Git branch with the specified name,
creates a worktree for that branch in the standard location,
opens the worktree in the configured editor,
and creates/switches to a tmux session.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Check if we're in a git repository
		if !git.IsGitRepository() {
			fmt.Fprintf(os.Stderr, "Error: not in a git repository\n")
			os.Exit(1)
		}
		
		// Get repository name
		repoName, err := git.GetRepositoryName()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		// Get or prompt for worktree name
		var worktreeName string
		if len(args) > 0 {
			worktreeName = args[0]
		} else {
			// Interactive mode - prompt for name
			fmt.Print("Enter worktree name: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
				os.Exit(1)
			}
			worktreeName = strings.TrimSpace(input)
			if worktreeName == "" {
				fmt.Fprintf(os.Stderr, "Error: worktree name cannot be empty\n")
				os.Exit(1)
			}
		}
		
		// Save current branch to restore later
		originalBranch, err := git.GetCurrentBranch()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not get current branch: %v\n", err)
		}
		
		// Create new branch
		fmt.Printf("Creating branch '%s'...\n", worktreeName)
		if err := git.CreateBranch(worktreeName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		// Create worktree
		fmt.Printf("Creating worktree for '%s'...\n", worktreeName)
		worktreePath, err := git.CreateWorktree(repoName, worktreeName, worktreeName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			// Try to clean up the branch we created
			if originalBranch != "" {
				_ = git.CheckoutBranch(originalBranch)
			}
			os.Exit(1)
		}
		
		// Switch back to original branch
		if originalBranch != "" {
			if err := git.CheckoutBranch(originalBranch); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not switch back to original branch: %v\n", err)
			}
		}
		
		fmt.Printf("Worktree created at: %s\n", worktreePath)
		
		// Open in editor
		if err := editor.OpenInEditor(worktreePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		} else {
			fmt.Printf("Opened in editor\n")
		}
		
		// Create/switch tmux session
		sessionName := tmux.SanitizeSessionName(worktreeName)
		if err := tmux.CreateSession(sessionName, worktreePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		} else if tmux.IsInstalled() {
			fmt.Printf("Tmux session '%s' created/switched\n", sessionName)
		}
	},
}