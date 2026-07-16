package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"note_cli/api"
)

func TestParseIDArg(t *testing.T) {
	id, err := parseIDArg([]string{"42"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
}

func TestParseIDArgRejectsInvalidValue(t *testing.T) {
	if _, err := parseIDArg([]string{"abc"}); err == nil {
		t.Fatal("expected error for invalid id")
	}
}

func TestValidateAttachedFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "memo.txt")
	if err := os.WriteFile(path, []byte("hello"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	if err := validateAttachedFiles([]string{path}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAttachedFilesRejectsMissingFile(t *testing.T) {
	err := validateAttachedFiles([]string{"/tmp/does-not-exist"})
	if err == nil {
		t.Fatal("expected missing file error")
	}
	if !strings.Contains(err.Error(), "파일을 찾을 수 없습니다") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTruncateText(t *testing.T) {
	got := truncateText("1234567890", 7)
	if got != "1234..." {
		t.Fatalf("unexpected truncated text: %q", got)
	}
}

func TestTruncateTextKorean(t *testing.T) {
	// 한글은 표시 폭 2칸. 바이트 기준으로 자르면 문자가 깨지므로 폭 기준으로 잘라야 한다.
	got := truncateText("가나다라마바사", 9)
	if got != "가나다..." {
		t.Fatalf("unexpected truncated text: %q", got)
	}
	if strings.Contains(got, "�") {
		t.Fatalf("truncated text contains replacement character: %q", got)
	}

	// 폭이 충분하면 그대로 반환
	if got := truncateText("가나다", 10); got != "가나다" {
		t.Fatalf("unexpected text: %q", got)
	}
}

func TestFormatTimestamp(t *testing.T) {
	got := formatTimestamp("2026-04-17T09:00:30Z", 16)
	if got != "2026-04-17 09:00" {
		t.Fatalf("unexpected timestamp: %q", got)
	}
}

func TestDisplayFileNamePrefersOriginal(t *testing.T) {
	file := api.FileRead{
		Filename:         "stored-name.png",
		OriginalFilename: "photo.png",
	}

	if got := displayFileName(file); got != "photo.png" {
		t.Fatalf("unexpected display filename: %q", got)
	}
}

func TestFileNameFromURL(t *testing.T) {
	got := fileNameFromURL("https://example.com/files/image.png")
	if got != "image.png" {
		t.Fatalf("unexpected filename: %q", got)
	}
}

func TestBoardTitleByFileName(t *testing.T) {
	boards := []api.BoardRead{
		{
			Title:  "회의 메모",
			Images: []string{"https://example.com/files/a.png"},
		},
	}

	got := boardTitleByFileName(boards)
	if got["a.png"] != "회의 메모" {
		t.Fatalf("unexpected mapping: %#v", got)
	}
}

func TestFindFileByID(t *testing.T) {
	files := []api.FileRead{
		{ID: 1, Filename: "a.txt"},
		{ID: 2, Filename: "b.txt"},
	}

	file, ok := findFileByID(files, 2)
	if !ok {
		t.Fatal("expected file to be found")
	}
	if file.Filename != "b.txt" {
		t.Fatalf("unexpected file: %#v", file)
	}
}
