package tui

import (
	"fmt"
	"os"
	"os/exec"
)

// OpenEditor opens the user's preferred editor with initial content
// and returns the user's updated content.
func OpenEditor(initialContent string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	tempFile, err := os.CreateTemp("", "note_cli_*.md")
	if err != nil {
		return "", fmt.Errorf("could not create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if initialContent != "" {
		if _, err := tempFile.WriteString(initialContent); err != nil {
			return "", fmt.Errorf("could not write initial content: %w", err)
		}
	}
	// Close file to allow editor to open and save it properly
	tempFile.Close()

	cmd := exec.Command(editor, tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor returned error: %w", err)
	}

	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("could not read temp file: %w", err)
	}

	return string(content), nil
}
