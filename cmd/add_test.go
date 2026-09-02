package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"note_cli/api"
)

func TestResolveNoteContent(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "note.md")
	if err := os.WriteFile(filePath, []byte("파일 내용\n"), 0o644); err != nil {
		t.Fatalf("테스트 파일 생성 실패: %v", err)
	}
	emptyPath := filepath.Join(dir, "empty.md")
	if err := os.WriteFile(emptyPath, []byte("\n"), 0o644); err != nil {
		t.Fatalf("빈 테스트 파일 생성 실패: %v", err)
	}

	tests := []struct {
		name        string
		content     string
		contentFile string
		stdin       string
		want        string
		wantErr     bool
	}{
		{name: "직접 지정한 내용 반환", content: "본문 내용", want: "본문 내용"},
		{name: "직접 지정 내용 끝 개행 제거", content: "본문 내용\n", want: "본문 내용"},
		{name: "파일에서 내용 읽기", contentFile: filePath, want: "파일 내용"},
		{name: "존재하지 않는 파일은 에러", contentFile: filepath.Join(dir, "missing.md"), wantErr: true},
		{name: "빈 파일은 에러", contentFile: emptyPath, wantErr: true},
		{name: "stdin 하이픈 지정", content: "-", stdin: "stdin 내용\n", want: "stdin 내용"},
		{name: "stdin이 비면 에러", content: "-", stdin: "", wantErr: true},
		{name: "빈 내용은 에러", content: "", wantErr: true},
		{name: "content와 content-file 동시 지정은 에러", content: "a", contentFile: filePath, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveNoteContent(tt.content, tt.contentFile, strings.NewReader(tt.stdin))
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveNoteContent() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("resolveNoteContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeNoteCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		want     string
		wantErr  bool
	}{
		{name: "빈 값은 기본값", category: "", want: defaultNoteCategory},
		{name: "유효한 값", category: "work", want: "work"},
		{name: "대소문자 무시", category: "Idea", want: "idea"},
		{name: "공백 제거", category: "  work  ", want: "work"},
		{name: "유효하지 않은 값은 에러", category: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeNoteCategory(tt.category)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeNoteCategory() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("normalizeNoteCategory() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCategoryValues(t *testing.T) {
	values := categoryValues()
	if len(values) == 0 {
		t.Fatal("categoryValues()가 빈 목록을 반환했습니다")
	}

	seen := map[string]bool{}
	for _, v := range values {
		if v == "" {
			t.Fatalf("categoryValues()에 빈 값이 포함되어 있습니다: %v", values)
		}
		if seen[v] {
			t.Fatalf("categoryValues()에 중복 값이 있습니다: %v", values)
		}
		seen[v] = true
	}

	if _, ok := seen[defaultNoteCategory]; !ok {
		t.Fatalf("기본 카테고리 %q가 목록에 없습니다: %v", defaultNoteCategory, values)
	}
}

func TestMergeNoteFields(t *testing.T) {
	note := &api.BoardRead{
		Title:    "기존 제목",
		Category: "work",
		Content:  "기존 내용",
	}

	newTitle := "새 제목"
	newCategory := "idea"
	newContent := "새 내용"

	tests := []struct {
		name         string
		provided     providedNoteFields
		wantTitle    string
		wantCategory string
		wantContent  string
	}{
		{name: "아무것도 지정하지 않으면 기존 값 유지", provided: providedNoteFields{}, wantTitle: note.Title, wantCategory: note.Category, wantContent: note.Content},
		{name: "제목만 변경", provided: providedNoteFields{Title: &newTitle}, wantTitle: newTitle, wantCategory: note.Category, wantContent: note.Content},
		{name: "카테고리만 변경", provided: providedNoteFields{Category: &newCategory}, wantTitle: note.Title, wantCategory: newCategory, wantContent: note.Content},
		{name: "내용만 변경", provided: providedNoteFields{Content: &newContent}, wantTitle: note.Title, wantCategory: note.Category, wantContent: newContent},
		{name: "모두 변경", provided: providedNoteFields{Title: &newTitle, Category: &newCategory, Content: &newContent}, wantTitle: newTitle, wantCategory: newCategory, wantContent: newContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTitle, gotCategory, gotContent := mergeNoteFields(note, tt.provided)
			if gotTitle != tt.wantTitle || gotCategory != tt.wantCategory || gotContent != tt.wantContent {
				t.Fatalf("mergeNoteFields() = (%q, %q, %q), want (%q, %q, %q)",
					gotTitle, gotCategory, gotContent, tt.wantTitle, tt.wantCategory, tt.wantContent)
			}
		})
	}
}
