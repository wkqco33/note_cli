package cmd

import (
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
