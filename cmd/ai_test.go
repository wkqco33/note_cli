package cmd

import (
	"strings"
	"testing"

	"note_cli/api"
	appLLM "note_cli/llm"
)

func TestFormatTodoNote(t *testing.T) {
	note := &api.BoardRead{ID: 12, Title: "프로젝트 회의"}
	content := formatTodoNote(note, appLLM.TodoResult{Items: []appLLM.TodoItem{{Task: "API 구현", Assignee: "김철수", DueDate: "금요일"}}})
	want := "# TODO: 프로젝트 회의\n\n원본 노트: #12\n\n- [ ] API 구현\n  - 담당: 김철수\n  - 기한: 금요일\n"
	if content != want {
		t.Fatalf("formatTodoNote() = %q, want %q", content, want)
	}
}

func TestMaskedValue(t *testing.T) {
	for _, test := range []struct {
		input string
		want  string
	}{
		{"", "미설정"},
		{"abc", "****"},
		{"abcdef", "ab****ef"},
	} {
		if got := maskedValue(test.input); got != test.want {
			t.Errorf("maskedValue(%q) = %q, want %q", test.input, got, test.want)
		}
	}
	if strings.Contains(maskedValue("secret-key"), "secret") {
		t.Fatal("maskedValue leaked the API key")
	}
}
