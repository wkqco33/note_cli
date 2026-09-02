package cmd

import (
	"strings"
	"testing"
)

func TestJoinCreateRequest(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "단일 인자", args: []string{"회의 노트 만들어줘"}, want: "회의 노트 만들어줘"},
		{name: "여러 인자 결합", args: []string{"어제", "회의 내용", "정리"}, want: "어제 회의 내용 정리"},
		{name: "앞뒤 공백 제거", args: []string{"  요청  "}, want: "요청"},
		{name: "인자 없으면 빈 값", args: nil, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinCreateRequest(tt.args); got != tt.want {
				t.Fatalf("joinCreateRequest(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestResolveAICreateCategory(t *testing.T) {
	tests := []struct {
		name     string
		proposed string
		override string
		want     string
		wantErr  bool
	}{
		{name: "LLM 제안 유효", proposed: "work", want: "work"},
		{name: "LLM 제안 대소문자 무시", proposed: "Idea", want: "idea"},
		{name: "LLM 제안 무효는 기본값", proposed: "random-category", want: defaultNoteCategory},
		{name: "LLM 제안 비어 있으면 기본값", proposed: "", want: defaultNoteCategory},
		{name: "플래그 지정이 LLM 제안보다 우선", proposed: "work", override: "idea", want: "idea"},
		{name: "플래그 무효는 에러", proposed: "work", override: "invalid", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveAICreateCategory(tt.proposed, tt.override)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveAICreateCategory() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("resolveAICreateCategory() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderAICreateDryRun(t *testing.T) {
	draft := aiCreateDraft{Title: "제목", Content: "본문", Category: "work"}
	got := renderAICreateDryRun(draft)

	for _, want := range []string{"제목", "work", "본문", "--dry-run"} {
		if !strings.Contains(got, want) {
			t.Fatalf("renderAICreateDryRun()에 %q가 없습니다:\n%s", want, got)
		}
	}
}
