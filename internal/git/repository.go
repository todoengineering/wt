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
	// First check if we're in a worktree by getting the common git dir
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	
	gitCommonDir := strings.TrimSpace(string(output))
	
	// Convert to absolute path if relative
	if !filepath.IsAbs(gitCommonDir) {
		// Get the current working directory to resolve relative path
		cmd = exec.Command("git", "rev-parse", "--show-toplevel")
		topLevel, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("not in a git repository")
		}
		topLevelPath := strings.TrimSpace(string(topLevel))
		gitCommonDir = filepath.Join(topLevelPath, gitCommonDir)
	}
	
	// If the common dir ends with .git, we're in a worktree or the main repo
	if strings.HasSuffix(gitCommonDir, ".git") {
		// Get the parent directory of .git to find the main repository path
		mainRepoPath := filepath.Dir(gitCommonDir)
		return filepath.Base(mainRepoPath), nil
	}
	
	// Fallback: shouldn't normally reach here
	return "", fmt.Errorf("unable to determine repository name")
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