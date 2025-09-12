package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Worktree struct {
	Name string
	Path string
}

func ListWorktrees(repoName string) ([]Worktree, error) {
	worktreeDir := GetWorktreeDir(repoName)
	
	if _, err := os.Stat(worktreeDir); os.IsNotExist(err) {
		return []Worktree{}, nil
	}
	
	entries, err := os.ReadDir(worktreeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read worktree directory: %w", err)
	}
	
	var worktrees []Worktree
	for _, entry := range entries {
		if entry.IsDir() {
			worktrees = append(worktrees, Worktree{
				Name: entry.Name(),
				Path: filepath.Join(worktreeDir, entry.Name()),
			})
		}
	}
	
	sort.Slice(worktrees, func(i, j int) bool {
		return strings.ToLower(worktrees[i].Name) < strings.ToLower(worktrees[j].Name)
	})
	
	return worktrees, nil
}

func CreateWorktree(repoName, worktreeName, branchName string) (string, error) {
	worktreeDir := GetWorktreeDir(repoName)
	worktreePath := filepath.Join(worktreeDir, worktreeName)
	
	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(worktreeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree directory: %w", err)
	}
	
	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		return "", fmt.Errorf("worktree '%s' already exists at %s", worktreeName, worktreePath)
	}
	
	// Create the worktree
	cmd := exec.Command("git", "worktree", "add", worktreePath, branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create worktree: %s", string(output))
	}
	
	return worktreePath, nil
}