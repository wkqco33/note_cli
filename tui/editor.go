package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// resolveEditorCommand EDITOR 환경변수를 실행 커맨드와 인자로 분해한다.
// 값 전체가 실행 파일 경로면 그대로, 아니면 공백 기준으로 분리한다.
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

// OpenEditor 편집기로 초기 내용을 열고 사용자가 수정한 내용을 반환한다.
func OpenEditor(initialContent string) (string, error) {
	editor, editorArgs := resolveEditorCommand()

	tempFile, err := os.CreateTemp("", "note_cli_*.md")
	if err != nil {
		return "", fmt.Errorf("임시 파일을 생성하지 못했습니다: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if initialContent != "" {
		if _, err := tempFile.WriteString(initialContent); err != nil {
			return "", fmt.Errorf("초기 내용을 쓰지 못했습니다: %w", err)
		}
	}
	tempFile.Close()

	cmd := exec.Command(editor, append(editorArgs, tempFile.Name())...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("편집기가 오류를 반환했습니다: %w", err)
	}

	content, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("임시 파일을 읽지 못했습니다: %w", err)
	}

	return string(content), nil
}
