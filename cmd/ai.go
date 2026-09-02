package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"note_cli/api"
	"note_cli/attachments"
	"note_cli/embedding"
	appLLM "note_cli/llm"
	"note_cli/utils"

	llmapi "github.com/wkqco33/LLM_client_go"
	"github.com/wkqco33/wcli"
)

var applyImprovement bool
var createTodoNote bool
var improveComment string

var aiCmd = &wcli.Command{
	Use:   "ai",
	Short: "LLM을 사용한 노트 기능",
}

var aiStatusCmd = &wcli.Command{
	Use:   "status",
	Short: "LLM 설정 상태 확인",
	Run: func(ctx *wcli.Context) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		if _, err := appLLM.NewClient(cfg); err != nil {
			return err
		}
		fmt.Printf("LLM provider: %s\n", cfg.LLM.Provider)
		fmt.Printf("LLM model: %s\n", cfg.LLM.Model)
		fmt.Printf("LLM base URL: %s\n", cfg.LLM.BaseURL)
		if cfg.LLM.Provider == "openai" {
			fmt.Printf("API key: %s\n", maskedValue(cfg.LLM.APIKey))
		}
		return nil
	},
}

var aiImproveCmd = &wcli.Command{
	Use:   "improve <id>",
	Short: "노트 문장과 Markdown 개선안 생성",
	Run: func(ctx *wcli.Context) error {
		if err := requireMaxArgs(ctx.Args, 1); err != nil {
			return err
		}
		store, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = store.Close() }()
		id, ok, err := resolveBoardID(store, ctx.Args, "개선할 노트를 선택하세요", "개선할 노트가 없습니다.")
		if err != nil || !ok {
			return err
		}
		note, err := store.GetBoard(id)
		if err != nil {
			return fmt.Errorf("노트를 불러오지 못했습니다: %w", err)
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		client, err := appLLM.NewClient(cfg)
		if err != nil {
			return err
		}
		var result appLLM.ImproveResult
		err = utils.WithSpinner("노트 개선안 생성 중...", func() error {
			var innerErr error
			result, innerErr = appLLM.ImproveWithComment(context.Background(), client, cfg.LLM.Model, note.Title, note.Content, improveComment)
			return innerErr
		})
		if err != nil {
			return err
		}
		printImprovement(note, result)
		if applyImprovement {
			updated, err := store.UpdateBoard(id, api.BoardUpdate{Title: &result.Title, Content: &result.Content})
			if err != nil {
				return fmt.Errorf("개선된 노트를 저장하지 못했습니다: %w", err)
			}
			fmt.Printf("개선된 노트를 저장했습니다. ID: %d\n", updated.ID)
		}
		return nil
	},
}

var aiTodosCmd = &wcli.Command{
	Use:   "todos <id>",
	Short: "노트에서 할 일 목록 추출",
	Run: func(ctx *wcli.Context) error {
		if err := requireMaxArgs(ctx.Args, 1); err != nil {
			return err
		}
		store, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = store.Close() }()
		id, ok, err := resolveBoardID(store, ctx.Args, "할 일을 추출할 노트를 선택하세요", "할 일을 추출할 노트가 없습니다.")
		if err != nil || !ok {
			return err
		}
		note, err := store.GetBoard(id)
		if err != nil {
			return fmt.Errorf("노트를 불러오지 못했습니다: %w", err)
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		client, err := appLLM.NewClient(cfg)
		if err != nil {
			return err
		}
		var result appLLM.TodoResult
		err = utils.WithSpinner("할 일 목록 추출 중...", func() error {
			var innerErr error
			result, innerErr = appLLM.ExtractTodos(context.Background(), client, cfg.LLM.Model, note.Title, note.Content)
			return innerErr
		})
		if err != nil {
			return err
		}
		printTodos(note, result)
		if createTodoNote {
			content := formatTodoNote(note, result)
			created, err := store.CreateBoard(api.BoardCreate{Title: "TODO: " + note.Title, Content: content, Category: "Other"})
			if err != nil {
				return fmt.Errorf("할 일 노트를 생성하지 못했습니다: %w", err)
			}
			fmt.Printf("할 일 노트를 생성했습니다. ID: %d\n", created.ID)
		}
		return nil
	},
}

var aiSummarizeFileCmd = &wcli.Command{
	Use:   "summarize-file [file-id]",
	Short: "텍스트 첨부파일 요약",
	Run: func(ctx *wcli.Context) error {
		if err := requireMaxArgs(ctx.Args, 1); err != nil {
			return err
		}
		store, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = store.Close() }()
		var file api.FileRead
		var id int
		if len(ctx.Args) == 1 {
			id, err = parseIDArg(ctx.Args)
			if err != nil {
				return err
			}
			files, filesErr := store.GetFiles()
			if filesErr != nil {
				return fmt.Errorf("첨부파일 목록을 불러오지 못했습니다: %w", filesErr)
			}
			var found bool
			file, found = findFileByID(files, id)
			if !found {
				return fmt.Errorf("첨부파일을 찾을 수 없습니다: %d", id)
			}
		} else {
			var files []api.FileRead
			var selected bool
			id, files, selected, err = resolveFileID(store, "요약할 첨부파일을 선택하세요", "요약할 첨부파일이 없습니다.")
			if err != nil || !selected {
				return err
			}
			file, _ = findFileByID(files, id)
		}
		stream, _, err := store.GetFileStream(id)
		if err != nil {
			return fmt.Errorf("첨부파일을 열지 못했습니다: %w", err)
		}
		defer func() { _ = stream.Close() }()
		content, err := extractAttachmentText(file, stream)
		if err != nil {
			return err
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		client, err := appLLM.NewClient(cfg)
		if err != nil {
			return err
		}
		var result appLLM.FileSummaryResult
		err = utils.WithSpinner("첨부파일 요약 중...", func() error {
			var innerErr error
			result, innerErr = appLLM.SummarizeText(context.Background(), client, cfg.LLM.Model, file.OriginalFilename, content)
			return innerErr
		})
		if err != nil {
			return err
		}
		fmt.Printf("첨부파일 #%d: %s\n\n%s\n", file.ID, file.OriginalFilename, result.Summary)
		if len(result.KeyPoints) > 0 {
			fmt.Println("핵심 내용:")
			for _, point := range result.KeyPoints {
				fmt.Printf("- %s\n", point)
			}
		}
		if len(result.Warnings) > 0 {
			fmt.Println("주의 사항:")
			for _, warning := range result.Warnings {
				fmt.Printf("- %s\n", warning)
			}
		}
		return nil
	},
}

var aiIndexCmd = &wcli.Command{
	Use:   "index [id]",
	Short: "노트 임베딩 생성 또는 갱신",
	Run: func(ctx *wcli.Context) error {
		if err := requireMaxArgs(ctx.Args, 1); err != nil {
			return err
		}
		store, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = store.Close() }()
		embStore, ok := store.(embeddingStore)
		if !ok {
			return fmt.Errorf("임베딩 인덱싱은 현재 로컬 모드에서만 지원합니다")
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		client, err := appLLM.NewClient(cfg)
		if err != nil {
			return err
		}
		notes, err := store.GetBoards()
		if err != nil {
			return fmt.Errorf("노트 목록을 불러오지 못했습니다: %w", err)
		}
		if len(ctx.Args) == 1 {
			id, err := parseIDArg(ctx.Args)
			if err != nil {
				return err
			}
			notes = filterNotesByID(notes, id)
		}
		var indexed int
		err = utils.WithSpinner("임베딩 인덱싱 중...", func() error {
			var innerErr error
			indexed, innerErr = indexNotes(context.Background(), client, embStore, cfg.LLM.EmbeddingModel, notes)
			return innerErr
		})
		if err != nil {
			return err
		}
		fmt.Printf("임베딩 인덱싱 완료: %d개 노트\n", indexed)
		return nil
	},
}

var aiSearchCmd = &wcli.Command{
	Use:   "search <query>",
	Short: "자연어로 노트 검색",
	Run: func(ctx *wcli.Context) error {
		if len(ctx.Args) == 0 {
			return fmt.Errorf("검색어를 입력해야 합니다")
		}
		store, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = store.Close() }()
		embStore, ok := store.(embeddingStore)
		if !ok {
			return fmt.Errorf("의미 검색은 현재 로컬 모드에서만 지원합니다")
		}
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		client, err := appLLM.NewClient(cfg)
		if err != nil {
			return err
		}
		var response *llmapi.EmbeddingResponse
		err = utils.WithSpinner("검색어 임베딩 생성 중...", func() error {
			var innerErr error
			response, innerErr = client.CreateEmbeddings(context.Background(), llmEmbeddingRequest(cfg.LLM.EmbeddingModel, strings.Join(ctx.Args, " ")))
			return innerErr
		})
		if err != nil {
			return fmt.Errorf("검색어 임베딩을 생성하지 못했습니다: %w", err)
		}
		if len(response.Data) == 0 {
			return fmt.Errorf("검색어 임베딩 응답이 비어 있습니다")
		}
		records, err := embStore.GetEmbeddings()
		if err != nil {
			return err
		}
		results, err := semanticResults(store, records, response.Data[0].Embedding)
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Println("검색 결과가 없습니다. 'ai index'를 먼저 실행하세요.")
			return nil
		}
		fmt.Println("관련 노트")
		for _, result := range results {
			fmt.Printf("%.2f  #%d  %s\n", result.Score, result.Note.ID, result.Note.Title)
		}
		return nil
	},
}

type semanticResult struct {
	Note  api.BoardRead
	Score float32
}

func llmEmbeddingRequest(model, text string) llmapi.EmbeddingRequest {
	return llmapi.EmbeddingRequest{Model: model, Input: []string{text}, EncodingFormat: "float"}
}

func extractAttachmentText(file api.FileRead, stream io.Reader) (string, error) {
	return attachments.ExtractText(file.OriginalFilename, file.ContentType, stream)
}

func filterNotesByID(notes []api.BoardRead, id int) []api.BoardRead {
	for _, note := range notes {
		if note.ID == id {
			return []api.BoardRead{note}
		}
	}
	return nil
}

func indexNotes(ctx context.Context, client llmapi.Client, store embeddingStore, model string, notes []api.BoardRead) (int, error) {
	if len(notes) == 0 {
		return 0, nil
	}
	existing, err := store.GetEmbeddings()
	if err != nil {
		return 0, err
	}
	byID := make(map[int]embedding.Record, len(existing))
	for _, record := range existing {
		byID[record.NoteID] = record
	}
	var pending []api.BoardRead
	for _, note := range notes {
		hash := embedding.ContentHash(note.Title, note.Content, note.Category)
		if record, ok := byID[note.ID]; ok && record.Model == model && record.ContentHash == hash {
			continue
		}
		pending = append(pending, note)
	}
	if len(pending) == 0 {
		return 0, nil
	}
	inputs := make([]string, len(pending))
	for i, note := range pending {
		inputs[i] = note.Title + "\n" + note.Category + "\n" + note.Content
	}
	response, err := client.CreateEmbeddings(ctx, llmapi.EmbeddingRequest{Model: model, Input: inputs, EncodingFormat: "float"})
	if err != nil {
		return 0, fmt.Errorf("노트 임베딩을 생성하지 못했습니다: %w", err)
	}
	if len(response.Data) != len(pending) {
		return 0, fmt.Errorf("노트 임베딩 응답 개수가 일치하지 않습니다")
	}
	for i, data := range response.Data {
		note := pending[i]
		if err := store.SaveEmbedding(embedding.Record{NoteID: note.ID, Model: model, Dimensions: len(data.Embedding), Vector: data.Embedding, ContentHash: embedding.ContentHash(note.Title, note.Content, note.Category), UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano)}); err != nil {
			return 0, err
		}
	}
	return len(pending), nil
}

func semanticResults(store NoteStore, records []embedding.Record, query []float32) ([]semanticResult, error) {
	results := make([]semanticResult, 0, len(records))
	for _, record := range records {
		note, err := store.GetBoard(record.NoteID)
		if err != nil {
			continue
		}
		score, err := embedding.CosineSimilarity(query, record.Vector)
		if err != nil {
			continue
		}
		results = append(results, semanticResult{Note: *note, Score: score})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if len(results) > 10 {
		results = results[:10]
	}
	return results, nil
}

func printImprovement(note *api.BoardRead, result appLLM.ImproveResult) {
	fmt.Printf("노트 #%d 개선안\n\n", note.ID)
	fmt.Printf("제목: %s\n\n%s\n", result.Title, result.Content)
	if len(result.Changes) > 0 {
		fmt.Println("변경 사항:")
		for _, change := range result.Changes {
			fmt.Printf("- %s\n", change)
		}
	}
	if len(result.Warnings) > 0 {
		fmt.Println("주의 사항:")
		for _, warning := range result.Warnings {
			fmt.Printf("- %s\n", warning)
		}
	}
	if !applyImprovement {
		fmt.Println("저장하려면 --apply 옵션을 사용하세요.")
	}
}

func printTodos(note *api.BoardRead, result appLLM.TodoResult) {
	fmt.Printf("노트 #%d 할 일 목록\n", note.ID)
	if len(result.Items) == 0 {
		fmt.Println("추출된 할 일이 없습니다.")
		return
	}
	for _, item := range result.Items {
		fmt.Printf("[ ] %s\n", item.Task)
		if item.Assignee != "" {
			fmt.Printf("    담당: %s\n", item.Assignee)
		}
		if item.DueDate != "" {
			fmt.Printf("    기한: %s\n", item.DueDate)
		}
	}
	if !createTodoNote {
		fmt.Println("노트로 저장하려면 --create-note 옵션을 사용하세요.")
	}
}

func formatTodoNote(note *api.BoardRead, result appLLM.TodoResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# TODO: %s\n\n원본 노트: #%d\n\n", note.Title, note.ID)
	for _, item := range result.Items {
		fmt.Fprintf(&b, "- [ ] %s\n", item.Task)
		if item.Assignee != "" {
			fmt.Fprintf(&b, "  - 담당: %s\n", item.Assignee)
		}
		if item.DueDate != "" {
			fmt.Fprintf(&b, "  - 기한: %s\n", item.DueDate)
		}
	}
	return b.String()
}

func maskedValue(value string) string {
	if value == "" {
		return "미설정"
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}

func init() {
	aiImproveCmd.Flags().StringVar(&improveComment, "comment", "c", "", "개선 방향 또는 추가 요청")
	aiImproveCmd.Flags().BoolVar(&applyImprovement, "apply", "", false, "개선 결과를 노트에 저장")
	aiTodosCmd.Flags().BoolVar(&createTodoNote, "create-note", "", false, "추출 결과를 새 노트로 저장")
	aiCmd.AddCommand(aiStatusCmd)
	aiCmd.AddCommand(aiImproveCmd)
	aiCmd.AddCommand(aiTodosCmd)
	aiCmd.AddCommand(aiSummarizeFileCmd)
	aiCmd.AddCommand(aiIndexCmd)
	aiCmd.AddCommand(aiSearchCmd)
	rootCmd.AddCommand(aiCmd)
}
