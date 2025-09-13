package editor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GetEditorCommand(path string) string {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vi"  // Default fallback
	}
	
	// Return the full command with path
	return fmt.Sprintf("%s %s", editorCmd, path)
}

func OpenInEditor(path string) error {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		return fmt.Errorf("EDITOR environment variable not set")
	}
	
	// Split the editor command to handle flags (e.g., "code -n")
	parts := strings.Fields(editorCmd)
	if len(parts) == 0 {
		return fmt.Errorf("invalid EDITOR command")
	}
	
	// Append the path to the command
	args := append(parts[1:], path)
	
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}
	
	// Don't wait for the editor to close
	return nil
}