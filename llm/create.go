package llm

import (
	"context"
	"fmt"
	"strings"

	llmapi "github.com/wkqco33/LLM_client_go"
)

// createSystemPrompt 노트 생성 전용 시스템 프롬프트.
// 사용자 요청은 곧 명령이므로 injection 경고가 Improve와 방향이 다르다.
const createSystemPrompt = "너는 노트 작성 도우미다. 사용자 요청에 따라 하나의 노트를 작성하라. " +
	"요청에 근거가 없는 사실, 날짜, 담당자를 추측하지 말고 부족한 내용은 일반적인 표현으로 채워라. " +
	koreanOutputPrompt +
	" content는 Markdown 문서로 작성하며 첫 줄에 제목 헤더를 쓰지 않는다. " +
	"category는 work, personal, idea, other 중 노트 성격에 가장 맞는 값을 소문자로 반환하라. " +
	"반드시 title, content, category 키를 가진 JSON만 반환하라."

// CreateResult 사용자 요청으로 생성한 노트 초안.
type CreateResult struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category,omitempty"`
}

// CreateNote는 사용자 요청을 노트 초안(제목/본문/카테고리)으로 변환한다.
func CreateNote(ctx context.Context, client llmapi.Client, model, request string) (CreateResult, error) {
	request = strings.TrimSpace(request)
	if request == "" {
		return CreateResult{}, fmt.Errorf("노트 생성 요청이 비어 있습니다")
	}

	resp := llmapi.ChatRequest{
		Model: model,
		Messages: []llmapi.Message{
			{Role: llmapi.RoleSystem, Content: createSystemPrompt},
			{Role: llmapi.RoleUser, Content: fmt.Sprintf("다음 요청에 따라 노트를 작성하라.\n<request>\n%s\n</request>", request)},
		},
		ResponseFormat: jsonResponseFormat("note_creation", map[string]any{
			"type": "object", "properties": map[string]any{
				"title":    map[string]any{"type": "string"},
				"content":  map[string]any{"type": "string"},
				"category": map[string]any{"type": "string", "enum": []string{"work", "personal", "idea", "other"}},
			}, "required": []string{"title", "content", "category"},
		}),
	}
	response, err := client.Complete(ctx, resp)
	if err != nil {
		return CreateResult{}, fmt.Errorf("노트 생성을 요청하지 못했습니다: %w", err)
	}
	var result CreateResult
	if err := decodeResult(response, &result); err != nil {
		return CreateResult{}, fmt.Errorf("노트 생성 응답을 해석하지 못했습니다: %w", err)
	}
	result.Title = strings.TrimSpace(result.Title)
	result.Content = strings.TrimSpace(result.Content)
	result.Category = strings.TrimSpace(result.Category)
	if result.Title == "" || result.Content == "" {
		return CreateResult{}, fmt.Errorf("생성된 노트의 제목 또는 본문이 비어 있습니다")
	}

	return result, nil
}
