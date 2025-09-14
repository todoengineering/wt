package worktree

import (
    "bufio"
    "bytes"
    "fmt"
    "os"
    "os/exec"
    "sort"
    "strings"

    "github.com/todoengineering/wt/internal/editor"
    "github.com/todoengineering/wt/internal/git"
    "github.com/todoengineering/wt/internal/tmux"
    "github.com/spf13/cobra"
)

var newFromBranch string

var newCmd = &cobra.Command{
    Use:   "new <name>",
    Short: "Create new worktree",
    Long: `Two modes:
1) Default: Creates a new Git branch named <name> and a worktree for it.
2) With --from <branch>: Creates a worktree for an existing branch, optionally named <name>.
In both modes, opens the worktree in the configured editor and creates/switches to a tmux session.`,
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
		
        // Mode selection: from existing branch vs new branch
        var worktreeName string
        var worktreePath string
        if newFromBranch != "" {
            // Create worktree from an existing branch
            // Always fetch remote branches for up-to-date list
            fmt.Println("Fetching remote branches...")
            if err := git.FetchRemoteBranches(); err != nil {
                fmt.Fprintf(os.Stderr, "Warning: failed to fetch remote branches: %v\n", err)
            }

            sourceBranch := newFromBranch
            if sourceBranch == ":pick" {
                // Interactive selection
                branches, err := git.ListAllBranches()
                if err != nil {
                    fmt.Fprintf(os.Stderr, "Error listing branches: %v\n", err)
                    os.Exit(1)
                }

                // Filter out current branch
                currentBranch, _ := git.GetCurrentBranch()
                var selectableBranches []git.Branch
                for _, b := range branches {
                    if b.Name != currentBranch {
                        selectableBranches = append(selectableBranches, b)
                    }
                }
                if len(selectableBranches) == 0 {
                    fmt.Println("No other branches available")
                    os.Exit(1)
                }

                picked, err := selectBranchInteractive(selectableBranches)
                if err != nil {
                    fmt.Fprintf(os.Stderr, "Error selecting branch: %v\n", err)
                    os.Exit(1)
                }
                sourceBranch = picked
            }

            // Determine worktree name
            if len(args) > 0 {
                worktreeName = args[0]
            } else {
                // Default to sanitized branch name
                worktreeName = git.SanitizeBranchName(sourceBranch)
            }

            // If a worktree already exists for this branch, offer to switch instead
            if exists, existing := git.WorktreeExistsForBranch(repoName, sourceBranch); exists {
                fmt.Printf("A worktree already exists for branch '%s' at:\n  %s\n\n", sourceBranch, existing.Path)
                fmt.Println("Would you like to switch to it? (y/n)")
                var response string
                fmt.Scanln(&response)
                if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
                    // Reuse open behavior for consistency
                    openWorktree(repoName, *existing)
                    return
                }
                fmt.Println("Cancelled")
                os.Exit(0)
            }

            fmt.Printf("Creating worktree '%s' for branch '%s'...\n", worktreeName, sourceBranch)
            p, err := git.CreateWorktree(repoName, worktreeName, sourceBranch)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
            }
            worktreePath = p
        } else {
            // Default behavior: create a new branch, then a worktree for it
            // Get or prompt for worktree/branch name
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

            // Create worktree for the new branch
            fmt.Printf("Creating worktree for '%s'...\n", worktreeName)
            p, err := git.CreateWorktree(repoName, worktreeName, worktreeName)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
            }
            worktreePath = p
        }
		
		fmt.Printf("Worktree created at: %s\n", worktreePath)
		
        // Create/switch tmux session and/or open editor according to flags
        // Standardize on session name: <repo>-<worktree>
        sessionName := tmux.SanitizeSessionName(fmt.Sprintf("%s-%s", repoName, worktreeName))
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

func init() {
    newCmd.Flags().StringVar(&newFromBranch, "from", "", "create a worktree from an existing branch (optionally provide <name> for worktree)")
}

// Branch selection helpers (ported from former checkout flow)
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
        "--header=Select a branch to create a worktree from")
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
