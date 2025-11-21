package worktree

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/todoengineering/wt/internal/editor"
	"github.com/todoengineering/wt/internal/git"
	"github.com/todoengineering/wt/internal/tmux"
	"github.com/todoengineering/wt/internal/ui"
)

var (
	openAllFlag       bool
	openProjectFilter string
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open existing worktree (interactive)",
	Long: `Interactive two-step selection process using fzf:
1. Select project from available repositories
2. Select specific worktree within that project
Opens selected worktree in configured editor and creates/switches to tmux session.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Determine project scope
		projects, err := git.ListAllProjects()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing projects: %v\n", err)
			os.Exit(1)
		}

		// If filtering by project name
		if openProjectFilter != "" {
			var filtered []git.Project
			for _, p := range projects {
				if p.Name == openProjectFilter {
					filtered = append(filtered, p)
				}
			}
			projects = filtered
		} else if !openAllFlag && git.IsGitRepository() {
			// If inside a repo and --all not set, limit to current repo
			if repoName, err := git.GetRepositoryName(); err == nil {
				var filtered []git.Project
				for _, p := range projects {
					if p.Name == repoName {
						filtered = append(filtered, p)
						break
					}
				}
				projects = filtered
			}
		}

		if len(projects) == 0 {
			fmt.Println("No projects with worktrees found")
			fmt.Printf("Worktree base directory: %s\n", git.GetWorktreeBaseDir())
			os.Exit(1)
		}

		var selectedProject git.Project
		var selectedWorktree git.Worktree

		if len(projects) == 1 {
			// Only one project, skip project selection
			selectedProject = projects[0]
			fmt.Printf("Project: %s\n", selectedProject.Name)
		} else {
			// Multiple projects, select one
			selected, err := selectProjectInteractive(projects)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error selecting project: %v\n", err)
				os.Exit(1)
			}
			selectedProject = selected
		}

		// Select worktree from the chosen project
		if len(selectedProject.Worktrees) == 1 {
			// Only one worktree, skip worktree selection
			selectedWorktree = selectedProject.Worktrees[0]
			fmt.Printf("Worktree: %s\n", selectedWorktree.Name)
		} else {
			// Multiple worktrees, select one
			selected, err := selectWorktreeFromProject(selectedProject)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error selecting worktree: %v\n", err)
				os.Exit(1)
			}
			selectedWorktree = selected
		}

		// Open the selected worktree
		openWorktree(selectedProject.Name, selectedWorktree)
	},
}

func selectProjectInteractive(projects []git.Project) (git.Project, error) {
	var items []ui.Item
	for _, project := range projects {
		worktreeCount := len(project.Worktrees)
		desc := fmt.Sprintf("%d worktree%s", worktreeCount,
			func() string {
				if worktreeCount != 1 {
					return "s"
				}
				return ""
			}())

		items = append(items, ui.Item{
			TitleStr:       project.Name,
			DescriptionStr: desc,
			FilterStr:      project.Name,
			Value:          project,
		})
	}

	selected, err := ui.Select(items, "Select a project to browse worktrees")
	if err != nil {
		return git.Project{}, err
	}

	return selected.Value.(git.Project), nil
}

func selectWorktreeFromProject(project git.Project) (git.Worktree, error) {
	var items []ui.Item
	for _, worktree := range project.Worktrees {
		items = append(items, ui.Item{
			TitleStr:       worktree.Name,
			DescriptionStr: worktree.Path,
			FilterStr:      worktree.Name,
			Value:          worktree,
		})
	}

	selected, err := ui.Select(items, fmt.Sprintf("Select a worktree from %s", project.Name))
	if err != nil {
		return git.Worktree{}, err
	}

	return selected.Value.(git.Worktree), nil
}

func openWorktree(projectName string, worktree git.Worktree) {
	fmt.Printf("Opening worktree: %s/%s\n", projectName, worktree.Name)

	// Create or switch to tmux session
	sessionName := fmt.Sprintf("%s-%s", projectName, worktree.Name)
	sessionName = tmux.SanitizeSessionName(sessionName)

	if noTmux {
		if !noEditor {
			if err := editor.OpenInEditor(worktree.Path); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to open editor: %v\n", err)
			} else {
				fmt.Printf("Opened in editor\n")
			}
		}
		return
	}
	if tmux.IsInstalled() {
		if tmux.SessionExists(sessionName) {
			fmt.Printf("Switching to existing tmux session: %s\n", sessionName)
			if !noEditor {
				if err := tmux.SendCommandToSession(sessionName, editor.GetEditorCommand(worktree.Path)); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to open editor in session: %v\n", err)
				}
			}
			if err := tmux.SwitchToSession(sessionName); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to switch tmux session: %v\n", err)
			}
		} else {
			fmt.Printf("Creating new tmux session: %s\n", sessionName)
			if noEditor {
				if err := tmux.CreateSession(sessionName, worktree.Path); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to create tmux session: %v\n", err)
				}
			} else {
				if err := tmux.CreateSessionWithCommand(sessionName, worktree.Path, editor.GetEditorCommand(worktree.Path)); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to create tmux session: %v\n", err)
				}
			}
		}
	} else {
		if !noEditor {
			if err := editor.OpenInEditor(worktree.Path); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to open editor: %v\n", err)
			} else {
				fmt.Printf("Opened in editor\n")
			}
		}
	}
}

func init() {
	openCmd.Flags().BoolVar(&openAllFlag, "all", false, "list across all projects even when in a repository")
	openCmd.Flags().StringVar(&openProjectFilter, "project", "", "filter to a specific project name")
}
