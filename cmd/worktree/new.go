package worktree

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/todoengineering/wt/internal/editor"
	"github.com/todoengineering/wt/internal/git"
	"github.com/todoengineering/wt/internal/tmux"
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
			os.Exit(1)
		}
		
		fmt.Printf("Worktree created at: %s\n", worktreePath)
		
        // Create/switch tmux session and/or open editor according to flags
        sessionName := tmux.SanitizeSessionName(worktreeName)
        if noTmux {
            if !noEditor {
                if err := editor.OpenInEditor(worktreePath); err != nil {
                    fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
                } else {
                    fmt.Printf("Opened in editor\n")
                }
            }
            return
        }
        if tmux.IsInstalled() {
            if noEditor {
                if err := tmux.CreateSession(sessionName, worktreePath); err != nil {
                    fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
                } else {
                    fmt.Printf("Tmux session '%s' created\n", sessionName)
                }
            } else {
                if err := tmux.CreateSessionWithCommand(sessionName, worktreePath, editor.GetEditorCommand(worktreePath)); err != nil {
                    fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
                } else {
                    fmt.Printf("Tmux session '%s' created with editor\n", sessionName)
                }
            }
        } else {
            if !noEditor {
                if err := editor.OpenInEditor(worktreePath); err != nil {
                    fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
                } else {
                    fmt.Printf("Opened in editor\n")
                }
            }
        }
	},
}
