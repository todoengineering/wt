package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func IsInstalled() bool {
	cmd := exec.Command("which", "tmux")
	err := cmd.Run()
	return err == nil
}

func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func SessionExists(sessionName string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	err := cmd.Run()
	return err == nil
}

func CreateSession(sessionName, workingDir string) error {
	if !IsInstalled() {
		// Silently skip if tmux is not installed
		return nil
	}
	
	// If session already exists, just switch to it
	if SessionExists(sessionName) {
		return SwitchToSession(sessionName)
	}
	
	// Create new detached session
	cmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", workingDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create tmux session: %s", string(output))
	}
	
	// Switch to the new session
	return SwitchToSession(sessionName)
}

func CreateSessionWithCommand(sessionName, workingDir, command string) error {
	if !IsInstalled() {
		// Silently skip if tmux is not installed
		return nil
	}
	
	// If session already exists, just switch to it
	if SessionExists(sessionName) {
		return SwitchToSession(sessionName)
	}
	
	// Create new session with the editor command
	var cmd *exec.Cmd
	if IsInsideTmux() {
		// Create detached session and then switch
		cmd = exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", workingDir, command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create tmux session: %s", string(output))
		}
		return SwitchToSession(sessionName)
	} else {
		// Create and attach to session directly with the command
		cmd = exec.Command("tmux", "new-session", "-s", sessionName, "-c", workingDir, command)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Start()
	}
}

func SendCommandToSession(sessionName, command string) error {
	if !IsInstalled() {
		return nil
	}
	
	// Send command to the first pane of the session
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName+":0.0", command, "Enter")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to send command to tmux session: %s", string(output))
	}
	
	return nil
}

func SwitchToSession(sessionName string) error {
	if !IsInstalled() {
		return nil
	}
	
	var cmd *exec.Cmd
	if IsInsideTmux() {
		// If inside tmux, switch client
		cmd = exec.Command("tmux", "switch-client", "-t", sessionName)
	} else {
		// If outside tmux, attach to session
		cmd = exec.Command("tmux", "attach-session", "-t", sessionName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Only return error if we're trying to switch from inside tmux
		// Attaching from outside tmux will block, so we don't wait for it
		if IsInsideTmux() {
			return fmt.Errorf("failed to switch to tmux session: %s", string(output))
		}
	}
	
	return nil
}

func SanitizeSessionName(name string) string {
	// Tmux session names can't contain certain characters
	// Replace them with underscores
	replacer := strings.NewReplacer(
		":", "_",
		".", "_",
		" ", "_",
		"/", "_",
		"\\", "_",
	)
	return replacer.Replace(name)
}