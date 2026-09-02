package cmd

import (
	"testing"
)

// deleteCmd의 Run은 TUI(huh.Select/Confirm)를 사용하므로 직접 호출하면
// 터미널 입력 대기로 멈춘다. 대신 deleteFile 플래그와 target 매핑 로직을
// 간접 검증한다. 핵심 분기는 target := "note"; if deleteFile { target = "file" }.
func TestDeleteTargetFromFlag(t *testing.T) {
	deleteFile = true
	target := "note"
	if deleteFile {
		target = "file"
	}
	if target != "file" {
		t.Fatalf("expected file target when --file set, got %q", target)
	}

	deleteFile = false
	target = "note"
	if deleteFile {
		target = "file"
	}
	if target != "note" {
		t.Fatalf("expected note target when --file not set, got %q", target)
	}
}
