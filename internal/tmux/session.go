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

func KillSession(sessionName string) error {
	if !IsInstalled() {
		return nil
	}

	// Check if session exists before trying to kill it
	if !SessionExists(sessionName) {
		return nil // Session doesn't exist, nothing to do
	}

	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to kill tmux session: %s", string(output))
	}

	return nil
}

type TmuxWindow struct {
	Name    string
	Command string
}

func CreateSessionWithNamedWindows(sessionName, workingDir string, windows []TmuxWindow) error {
	if !IsInstalled() {
		// Silently skip if tmux is not installed
		return nil
	}

	// If session already exists, just switch to it
	if SessionExists(sessionName) {
		return SwitchToSession(sessionName)
	}

	if len(windows) == 0 {
		// Fallback to regular session creation if no windows configured
		return CreateSession(sessionName, workingDir)
	}

	// Create new detached session with first window
	firstWindow := windows[0]
	var cmd *exec.Cmd
	if firstWindow.Command != "" {
		cmd = exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", workingDir, "-n", firstWindow.Name, firstWindow.Command)
	} else {
		cmd = exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", workingDir, "-n", firstWindow.Name)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create tmux session: %s", string(output))
	}

	// Create additional windows
	for _, window := range windows[1:] {
		if window.Command != "" {
			cmd = exec.Command("tmux", "new-window", "-t", sessionName, "-n", window.Name, "-c", workingDir, window.Command)
		} else {
			cmd = exec.Command("tmux", "new-window", "-t", sessionName, "-n", window.Name, "-c", workingDir)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create tmux window '%s': %s", window.Name, string(output))
		}
	}

	// Select the first window (tmux default behavior)
	cmd = exec.Command("tmux", "select-window", "-t", sessionName+":0")
	cmd.Run() // Ignore errors for this command

	// Switch to the new session
	return SwitchToSession(sessionName)
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
