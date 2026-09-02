package llm

import (
	"context"
	"strings"
	"testing"

	llmapi "github.com/wkqco33/LLM_client_go"
)

type fakeClient struct {
	response *llmapi.ChatResponse
	request  llmapi.ChatRequest
}

func (f *fakeClient) Complete(_ context.Context, request llmapi.ChatRequest) (*llmapi.ChatResponse, error) {
	f.request = request
	return f.response, nil
}

func (*fakeClient) Stream(context.Context, llmapi.ChatRequest) (llmapi.Stream, error) {
	return nil, nil
}

func (*fakeClient) CreateEmbeddings(context.Context, llmapi.EmbeddingRequest) (*llmapi.EmbeddingResponse, error) {
	return nil, nil
}

func (*fakeClient) TokenCounter(string) any { return nil }

func responseWithContent(content string) *llmapi.ChatResponse {
	return &llmapi.ChatResponse{Choices: []llmapi.Choice{{Message: llmapi.Message{Content: content}}}}
}

func TestImproveBuildsStructuredRequest(t *testing.T) {
	client := &fakeClient{response: responseWithContent(`{"title":"개선 제목","content":"개선 본문","changes":["문장 정리"],"warnings":[]}`)}
	result, err := Improve(context.Background(), client, "test-model", "원래 제목", "원래 본문")
	if err != nil {
		t.Fatalf("Improve() error = %v", err)
	}
	if result.Title != "개선 제목" || result.Content != "개선 본문" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if client.request.Model != "test-model" || client.request.ResponseFormat == nil {
		t.Fatalf("request did not include model and response format: %+v", client.request)
	}
	if !strings.Contains(client.request.Messages[0].Content, "한국어") {
		t.Fatal("improvement prompt does not require Korean output")
	}
}

func TestImproveWithCommentIncludesUserRequest(t *testing.T) {
	client := &fakeClient{response: responseWithContent(`{"title":"간결한 제목","content":"보강된 본문","changes":[],"warnings":[]}`)}
	_, err := ImproveWithComment(context.Background(), client, "model", "제목", "본문", "제목을 간결하게 작성하고 설명을 보강해줘")
	if err != nil {
		t.Fatalf("ImproveWithComment() error = %v", err)
	}
	if !strings.Contains(client.request.Messages[1].Content, "제목을 간결하게 작성하고 설명을 보강해줘") {
		t.Fatalf("user comment was not included: %s", client.request.Messages[1].Content)
	}
}

func TestImproveAcceptsStringChangeList(t *testing.T) {
	client := &fakeClient{response: responseWithContent(`{"title":"제목","content":"본문 개선","changes":"문장을 간결하게 수정","warnings":"주의 사항 없음"}`)}
	result, err := Improve(context.Background(), client, "model", "제목", "본문")
	if err != nil {
		t.Fatalf("Improve() error = %v", err)
	}
	if len(result.Changes) != 1 || result.Changes[0] != "문장을 간결하게 수정" {
		t.Fatalf("unexpected changes: %#v", result.Changes)
	}
	if len(result.Warnings) != 1 || result.Warnings[0] != "주의 사항 없음" {
		t.Fatalf("unexpected warnings: %#v", result.Warnings)
	}
}

func TestImproveKeepsOriginalWhenModelReturnsEmptyFields(t *testing.T) {
	client := &fakeClient{response: responseWithContent(`{"title":"","content":"","changes":[],"warnings":[]}`)}
	result, err := Improve(context.Background(), client, "model", "title", "content")
	if err != nil {
		t.Fatalf("Improve() error = %v", err)
	}
	if result.Title != "title" || result.Content != "content" || len(result.Warnings) != 2 {
		t.Fatalf("unexpected fallback result: %+v", result)
	}
}

func TestImproveAcceptsBodyFromNonConformingModel(t *testing.T) {
	client := &fakeClient{response: responseWithContent(`{"title":"개선 제목","body":"개선 본문","changes":[],"warnings":[]}`)}
	result, err := Improve(context.Background(), client, "model", "title", "content")
	if err != nil || result.Content != "개선 본문" {
		t.Fatalf("Improve() = %+v, error = %v", result, err)
	}
}

func TestExtractTodosFiltersEmptyTasks(t *testing.T) {
	client := &fakeClient{response: responseWithContent(`{"items":[{"task":"테스트 작성","assignee":"","due_date":"","source_text":"테스트"},{"task":"","assignee":"누구","due_date":"내일","source_text":""}]}`)}
	result, err := ExtractTodos(context.Background(), client, "model", "title", "content")
	if err != nil {
		t.Fatalf("ExtractTodos() error = %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Task != "테스트 작성" {
		t.Fatalf("unexpected todos: %+v", result.Items)
	}
}

func TestDecodeResultAcceptsJSONCodeFence(t *testing.T) {
	client := &fakeClient{response: responseWithContent("```json\n{\"title\":\"제목\",\"content\":\"본문\",\"changes\":[],\"warnings\":[]}\n```")}
	result, err := Improve(context.Background(), client, "model", "title", "content")
	if err != nil || result.Title != "제목" {
		t.Fatalf("Improve() = %+v, error = %v", result, err)
	}
}

func TestSummarizeText(t *testing.T) {
	client := &fakeClient{response: responseWithContent(`{"summary":"요약","key_points":["핵심"],"warnings":[]}`)}
	result, err := SummarizeText(context.Background(), client, "model", "notes.txt", "내용")
	if err != nil || result.Summary != "요약" {
		t.Fatalf("SummarizeText() = %+v, error = %v", result, err)
	}
}
