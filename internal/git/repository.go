package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dboroujerdi/wt/internal/config"
)

func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

func GetRepositoryName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	
	repoPath := strings.TrimSpace(string(output))
	return filepath.Base(repoPath), nil
}

func GetWorktreeBaseDir() string {
	if baseDir := os.Getenv("WORKTREE_BASE_DIR"); baseDir != "" {
		return baseDir
	}
	
	return config.GetWorktreesLocation()
}

func GetWorktreeDir(repoName string) string {
	return filepath.Join(GetWorktreeBaseDir(), repoName)
}