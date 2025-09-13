package worktree

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/todoengineering/wt/internal/git"
)

var (
    listAll bool
    listJSON bool
    listPathOnly bool
)

type listEntry struct {
    Project string `json:"project"`
    Name    string `json:"name"`
    Path    string `json:"path"`
    Branch  string `json:"branch"`
}

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List worktrees",
    Long:  `Lists on-disk worktrees for the current repository or across all projects when --all is provided. Supports JSON and path-only output for scripting.`,
    Run: func(cmd *cobra.Command, args []string) {
        if listAll {
            projects, err := git.ListAllProjects()
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error listing projects: %v\n", err)
                os.Exit(1)
            }

            if listJSON {
                var out []listEntry
                for _, p := range projects {
                    for _, wt := range p.Worktrees {
                        out = append(out, listEntry{Project: p.Name, Name: wt.Name, Path: wt.Path, Branch: wt.Branch})
                    }
                }
                enc := json.NewEncoder(os.Stdout)
                enc.SetIndent("", "  ")
                if err := enc.Encode(out); err != nil {
                    fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
                    os.Exit(1)
                }
                return
            }

            if listPathOnly {
                for _, p := range projects {
                    for _, wt := range p.Worktrees {
                        fmt.Println(wt.Path)
                    }
                }
                return
            }

            if len(projects) == 0 {
                fmt.Println("No projects with worktrees found")
                fmt.Printf("Worktree base directory: %s\n", git.GetWorktreeBaseDir())
                return
            }

            for _, p := range projects {
                fmt.Printf("%s:\n", p.Name)
                if len(p.Worktrees) == 0 {
                    fmt.Println("  (no worktrees)")
                    continue
                }
                for _, wt := range p.Worktrees {
                    fmt.Printf("  %s -> %s\n", wt.Name, wt.Path)
                }
            }
            return
        }

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

        if listJSON {
            var out []listEntry
            for _, wt := range worktrees {
                out = append(out, listEntry{Project: repoName, Name: wt.Name, Path: wt.Path, Branch: wt.Branch})
            }
            enc := json.NewEncoder(os.Stdout)
            enc.SetIndent("", "  ")
            if err := enc.Encode(out); err != nil {
                fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
                os.Exit(1)
            }
            return
        }

        if listPathOnly {
            for _, wt := range worktrees {
                fmt.Println(wt.Path)
            }
            return
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

func init() {
    listCmd.Flags().BoolVar(&listAll, "all", false, "list across all projects")
    listCmd.Flags().BoolVar(&listJSON, "json", false, "output JSON for scripting")
    listCmd.Flags().BoolVar(&listPathOnly, "path-only", false, "output only worktree paths")
}
