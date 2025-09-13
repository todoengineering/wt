package worktree

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/todoengineering/wt/internal/editor"
	"github.com/todoengineering/wt/internal/git"
	"github.com/todoengineering/wt/internal/tmux"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open existing worktree (interactive)",
	Long: `Interactive two-step selection process using fzf:
1. Select project from available repositories
2. Select specific worktree within that project
Opens selected worktree in configured editor and creates/switches to tmux session.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if fzf is available
		if _, err := exec.LookPath("fzf"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: fzf is required for the open command\n")
			fmt.Fprintf(os.Stderr, "Please install fzf: https://github.com/junegunn/fzf\n")
			os.Exit(1)
		}

		// List all projects
		projects, err := git.ListAllProjects()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing projects: %v\n", err)
			os.Exit(1)
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
	// Prepare input for fzf
	var input bytes.Buffer
	for _, project := range projects {
		worktreeCount := len(project.Worktrees)
		line := fmt.Sprintf("%-30s (%d worktree%s)", project.Name, worktreeCount, 
			func() string { if worktreeCount != 1 { return "s" }; return "" }())
		input.WriteString(line + "\n")
	}

	// Create fzf command for project selection
	cmd := exec.Command("fzf",
		"--prompt=Select project: ",
		"--height=40%",
		"--layout=reverse",
		"--header=Select a project to browse worktrees")
	cmd.Stdin = &input
	cmd.Stderr = os.Stderr

	// Run fzf and capture output
	output, err := cmd.Output()
	if err != nil {
		return git.Project{}, fmt.Errorf("selection cancelled")
	}

	// Parse the selected line to get project name
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return git.Project{}, fmt.Errorf("no selection made")
	}

	// Extract project name (first field)
	parts := strings.Fields(selected)
	if len(parts) > 0 {
		projectName := parts[0]
		for _, project := range projects {
			if project.Name == projectName {
				return project, nil
			}
		}
	}

	return git.Project{}, fmt.Errorf("could not find selected project")
}

func selectWorktreeFromProject(project git.Project) (git.Worktree, error) {
	// Prepare input for fzf
	var input bytes.Buffer
	for _, worktree := range project.Worktrees {
		line := fmt.Sprintf("%-30s %s", worktree.Name, worktree.Path)
		input.WriteString(line + "\n")
	}

	// Create fzf command for worktree selection
	cmd := exec.Command("fzf",
		"--prompt=Select worktree: ",
		"--height=40%",
		"--layout=reverse",
		fmt.Sprintf("--header=Select a worktree from %s", project.Name))
	cmd.Stdin = &input
	cmd.Stderr = os.Stderr

	// Run fzf and capture output
	output, err := cmd.Output()
	if err != nil {
		return git.Worktree{}, fmt.Errorf("selection cancelled")
	}

	// Parse the selected line to get worktree name
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return git.Worktree{}, fmt.Errorf("no selection made")
	}

	// Extract worktree name (first field)
	parts := strings.Fields(selected)
	if len(parts) > 0 {
		worktreeName := parts[0]
		for _, worktree := range project.Worktrees {
			if worktree.Name == worktreeName {
				return worktree, nil
			}
		}
	}

	return git.Worktree{}, fmt.Errorf("could not find selected worktree")
}

func openWorktree(projectName string, worktree git.Worktree) {
	fmt.Printf("Opening worktree: %s/%s\n", projectName, worktree.Name)

	// Create or switch to tmux session
	sessionName := fmt.Sprintf("%s-%s", projectName, worktree.Name)
	sessionName = tmux.SanitizeSessionName(sessionName)

	if tmux.IsInstalled() {
		if tmux.SessionExists(sessionName) {
			fmt.Printf("Switching to existing tmux session: %s\n", sessionName)
			// Open editor in the existing session before switching
			if err := tmux.SendCommandToSession(sessionName, editor.GetEditorCommand(worktree.Path)); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to open editor in session: %v\n", err)
			}
			if err := tmux.SwitchToSession(sessionName); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to switch tmux session: %v\n", err)
			}
		} else {
			fmt.Printf("Creating new tmux session: %s\n", sessionName)
			if err := tmux.CreateSessionWithCommand(sessionName, worktree.Path, editor.GetEditorCommand(worktree.Path)); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create tmux session: %v\n", err)
			}
		}
	} else {
		// No tmux, just open editor
		if err := editor.OpenInEditor(worktree.Path); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to open editor: %v\n", err)
		} else {
			fmt.Printf("Opened in editor\n")
		}
	}
}