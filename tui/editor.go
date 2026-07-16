package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// resolveEditorCommand EDITOR 환경변수를 실행 커맨드와 인자로 분해한다.
// 값 전체가 실행 파일 경로면 그대로 사용하고 (공백 포함 경로 지원),
// 아니면 "code --wait"처럼 인자가 붙은 형태로 보고 공백으로 분리한다.
func resolveEditorCommand() (string, []string) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			return "notepad", nil
		}
		return "vim", nil
	}

	if _, err := exec.LookPath(editor); err == nil {
		return editor, nil
	}

	fields := strings.Fields(editor)
	if len(fields) == 0 {
		return editor, nil
	}

	return fields[0], fields[1:]
}

// OpenEditor opens the user's preferred editor with initial content
// and returns the user's updated content.
func OpenEditor(initialContent string) (string, error) {
	editor, editorArgs := resolveEditorCommand()

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

	cmd := exec.Command(editor, append(editorArgs, tempFile.Name())...)
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
