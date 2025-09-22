package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func CreateBranch(branchName string) error {
	cmd := exec.Command("git", "branch", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %s", branchName, string(output))
	}
	return nil
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func CheckoutBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %s", branchName, string(output))
	}
	return nil
}

type Branch struct {
	Name     string
	IsRemote bool
	IsLocal  bool
}

func ListAllBranches() ([]Branch, error) {
	branches := make(map[string]*Branch)

	// Get local branches
	localCmd := exec.Command("git", "branch", "--format=%(refname:short)")
	localOutput, err := localCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list local branches: %w", err)
	}

	for _, line := range strings.Split(string(localOutput), "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			branches[name] = &Branch{
				Name:    name,
				IsLocal: true,
			}
		}
	}

	// Get remote branches
	remoteCmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	remoteOutput, err := remoteCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list remote branches: %w", err)
	}

	// Pattern to extract branch name from remote refs (e.g., origin/main -> main)
	remotePattern := regexp.MustCompile(`^[^/]+/(.+)$`)

	for _, line := range strings.Split(string(remoteOutput), "\n") {
		remoteName := strings.TrimSpace(line)
		if remoteName == "" || strings.HasSuffix(remoteName, "/HEAD") {
			continue
		}

		// Extract the branch name without the remote prefix
		matches := remotePattern.FindStringSubmatch(remoteName)
		if len(matches) > 1 {
			branchName := matches[1]
			if existing, ok := branches[branchName]; ok {
				existing.IsRemote = true
			} else {
				branches[branchName] = &Branch{
					Name:     branchName,
					IsRemote: true,
				}
			}
		}
	}

	// Convert map to slice
	result := make([]Branch, 0, len(branches))
	for _, branch := range branches {
		result = append(result, *branch)
	}

	return result, nil
}

func FetchRemoteBranches() error {
	cmd := exec.Command("git", "fetch", "--all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to fetch remote branches: %s", string(output))
	}
	return nil
}

func BranchExists(branchName string) (bool, error) {
	branches, err := ListAllBranches()
	if err != nil {
		return false, err
	}

	for _, branch := range branches {
		if branch.Name == branchName {
			return true, nil
		}
	}
	return false, nil
}
