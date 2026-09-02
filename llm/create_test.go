package llm

import (
	"context"
	"strings"
	"testing"
)

func TestCreateNoteBuildsStructuredRequest(t *testing.T) {
	client := &fakeClient{response: responseWithContent(`{"title":"회의 정리","content":"# 회의\n\n내용","category":"work"}`)}
	result, err := CreateNote(context.Background(), client, "test-model", "어제 회의 내용을 정리해서 노트로 만들어줘")
	if err != nil {
		t.Fatalf("CreateNote() error = %v", err)
	}
	if result.Title != "회의 정리" || result.Content != "# 회의\n\n내용" || result.Category != "work" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if client.request.Model != "test-model" || client.request.ResponseFormat == nil {
		t.Fatalf("request did not include model and response format: %+v", client.request)
	}
	if !strings.Contains(client.request.Messages[0].Content, "한국어") {
		t.Fatal("create prompt does not require Korean output")
	}
	if !strings.Contains(client.request.Messages[1].Content, "어제 회의 내용을 정리해서 노트로 만들어줘") {
		t.Fatalf("user request was not included: %s", client.request.Messages[1].Content)
	}
}

func TestCreateNoteRejectsEmptyResult(t *testing.T) {
	tests := []struct {
		name     string
		response string
	}{
		{name: "빈 제목", response: `{"title":"","content":"본문","category":"work"}`},
		{name: "빈 본문", response: `{"title":"제목","content":"","category":"work"}`},
		{name: "잘못된 JSON", response: "노트 내용입니다"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeClient{response: responseWithContent(tt.response)}
			if _, err := CreateNote(context.Background(), client, "model", "요청"); err == nil {
				t.Fatal("CreateNote() expected an error for invalid result")
			}
		})
	}
}

func TestCreateNoteAcceptsCodeFence(t *testing.T) {
	client := &fakeClient{response: responseWithContent("```json\n{\"title\":\"제목\",\"content\":\"본문\",\"category\":\"idea\"}\n```")}
	result, err := CreateNote(context.Background(), client, "model", "요청")
	if err != nil {
		t.Fatalf("CreateNote() error = %v", err)
	}
	if result.Title != "제목" || result.Category != "idea" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestCreateNoteDefaultsCategoryWhenMissing(t *testing.T) {
	client := &fakeClient{response: responseWithContent(`{"title":"제목","content":"본문"}`)}
	result, err := CreateNote(context.Background(), client, "model", "요청")
	if err != nil {
		t.Fatalf("CreateNote() error = %v", err)
	}
	if result.Category != "" {
		t.Fatalf("category should stay empty for cmd default handling, got %q", result.Category)
	}
}

func TestCreateNoteRejectsEmptyRequest(t *testing.T) {
	client := &fakeClient{}
	if _, err := CreateNote(context.Background(), client, "model", "   "); err == nil {
		t.Fatal("CreateNote() expected an error for empty request")
	}
}
