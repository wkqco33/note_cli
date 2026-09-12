package cmd

import (
	"bytes"
	"strings"
	"testing"

	"note_cli/api"
)

func TestPadRightPadsToWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{"ascii short", "ab", 5, "ab   "},
		{"ascii exact", "abcde", 5, "abcde"},
		{"korean counts double", "가", 5, "가   "},
		{"over width returns as-is", "abcdef", 4, "abcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := padRight(tt.input, tt.width); got != tt.want {
				t.Fatalf("padRight(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
			}
		})
	}
}

func TestFileExtensionFallsBackToStoredName(t *testing.T) {
	tests := []struct {
		name string
		file api.FileRead
		want string
	}{
		{"original present", api.FileRead{OriginalFilename: "report.PDF", Filename: "hash.bin"}, ".pdf"},
		{"original empty uses stored", api.FileRead{Filename: "data.txt"}, ".txt"},
		{"no extension", api.FileRead{OriginalFilename: "notes", Filename: "notes"}, ""},
		{"no ext but stored has", api.FileRead{Filename: "img.png"}, ".png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileExtension(tt.file); got != tt.want {
				t.Fatalf("fileExtension(%#v) = %q, want %q", tt.file, got, tt.want)
			}
		})
	}
}

func TestCategorySelectOptions(t *testing.T) {
	opts := categorySelectOptions()
	if len(opts) != 4 {
		t.Fatalf("expected 4 options, got %d", len(opts))
	}

	wantValues := []string{"work", "personal", "idea", "other"}
	for i, want := range wantValues {
		if opts[i].Value != want {
			t.Fatalf("option[%d] value = %q, want %q", i, opts[i].Value, want)
		}
	}
}

func TestBuildBoardTable(t *testing.T) {
	notes := []api.BoardRead{
		{ID: 1, Title: "할 일", Category: "work", UpdatedAt: "2024-01-02T10:00:00"},
		{ID: 2, Title: "아이디어 메모", Category: "idea", UpdatedAt: "2024-01-03T11:30:00"},
	}

	got := buildBoardTable(notes)
	if !strings.Contains(got, "ID") || !strings.Contains(got, "TITLE") {
		t.Fatalf("expected header in table, got:\n%s", got)
	}
	if !strings.Contains(got, "1") || !strings.Contains(got, "할 일") {
		t.Fatalf("expected first row in table, got:\n%s", got)
	}
	if !strings.Contains(got, "2024-01-02 10:00") {
		t.Fatalf("expected formatted timestamp, got:\n%s", got)
	}
}

func TestBuildBoardTableEmpty(t *testing.T) {
	if got := buildBoardTable(nil); got != "" {
		t.Fatalf("expected empty output for no notes, got: %q", got)
	}
}

// withStatusWriter statusWriter를 버퍼로 교체한다.
func withStatusWriter(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prev := statusWriter
	statusWriter = buf
	t.Cleanup(func() { statusWriter = prev })
	return buf
}

// withQuiet quietMode를 지정 값으로 바꾼다.
func withQuiet(t *testing.T, quiet bool) {
	t.Helper()
	prev := quietMode
	quietMode = quiet
	t.Cleanup(func() { quietMode = prev })
}

func TestStatusfWritesToStatusWriter(t *testing.T) {
	buf := withStatusWriter(t)
	withQuiet(t, false)

	statusf("노트 %d개 처리 중", 3)

	if got := buf.String(); !strings.Contains(got, "노트 3개 처리 중") {
		t.Fatalf("statusf() = %q, want 메시지 포함", got)
	}
}

func TestStatusfSuppressedInQuietMode(t *testing.T) {
	buf := withStatusWriter(t)
	withQuiet(t, true)

	statusf("보이면 안 됨")
	statusln("이것도 보이면 안 됨")

	if buf.Len() != 0 {
		t.Fatalf("quiet 모드에서 상태 메시지가 출력되었습니다: %q", buf.String())
	}
}

func TestStatusln(t *testing.T) {
	buf := withStatusWriter(t)
	withQuiet(t, false)

	statusln("줄바꿈 확인")

	if got := buf.String(); got != "줄바꿈 확인\n" {
		t.Fatalf("statusln() = %q, want %q", got, "줄바꿈 확인\n")
	}
}

func TestProgressBarVisible(t *testing.T) {
	tests := []struct {
		name      string
		quiet     bool
		stderrTTY bool
		want      bool
	}{
		{name: "일반 터미널", quiet: false, stderrTTY: true, want: true},
		{name: "quiet이면 숨김", quiet: true, stderrTTY: true, want: false},
		{name: "비TTY면 숨김", quiet: false, stderrTTY: false, want: false},
		{name: "둘 다면 숨김", quiet: true, stderrTTY: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := progressBarVisible(tt.quiet, tt.stderrTTY); got != tt.want {
				t.Fatalf("progressBarVisible(%v, %v) = %v, want %v", tt.quiet, tt.stderrTTY, got, tt.want)
			}
		})
	}
}

func TestResolveMarkdownStyle(t *testing.T) {
	if got := resolveMarkdownStyle(true); got != "notty" {
		t.Fatalf("resolveMarkdownStyle(true) = %q, want notty", got)
	}
	if got := resolveMarkdownStyle(false); got != "auto" {
		t.Fatalf("resolveMarkdownStyle(false) = %q, want auto", got)
	}
}
