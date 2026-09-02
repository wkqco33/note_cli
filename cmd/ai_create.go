package cmd

import (
	"context"
	"fmt"
	"strings"

	"note_cli/api"
	appLLM "note_cli/llm"
	"note_cli/utils"

	"github.com/charmbracelet/huh"
	"github.com/wkqco33/wcli"
)

var (
	aiCreateDryRun        bool
	aiCreateCategoryFlage string
)

// aiCreateDraft LLM이 생성한 노트 초안
type aiCreateDraft struct {
	Title    string
	Content  string
	Category string
}

// joinCreateRequest 가변 인자를 공백으로 결합해 단일 요청 문자열로 만든다
func joinCreateRequest(args []string) string {
	return strings.TrimSpace(strings.Join(args, " "))
}

// resolveAICreateCategory LLM 제안 카테고리와 --category 플래그를 반영해 최종 카테고리를 정한다.
// 플래그가 우선하며 플래그 값이 무효면 에러, LLM 제안이 무효/빈 값이면 기본값을 쓴다.
func resolveAICreateCategory(proposed, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return normalizeNoteCategory(override)
	}
	if category, err := normalizeNoteCategory(proposed); err == nil && strings.TrimSpace(proposed) != "" {
		return category, nil
	}
	return defaultNoteCategory, nil
}

// renderAICreateDryRun --dry-run 미리보기 출력을 구성한다
func renderAICreateDryRun(draft aiCreateDraft) string {
	var b strings.Builder
	fmt.Fprintf(&b, "제목: %s\n", draft.Title)
	fmt.Fprintf(&b, "카테고리: %s\n", draft.Category)
	fmt.Fprintf(&b, "\n%s\n", draft.Content)
	fmt.Fprintf(&b, "\n--dry-run 모드이므로 저장하지 않았습니다. 저장하려면 --dry-run 옵션을 빼고 다시 실행하세요.")
	return b.String()
}

var aiCreateCmd = &wcli.Command{
	Use:   "create [request...]",
	Short: "사용자 요청으로 LLM이 노트를 생성해 저장",
	Long: `사용자 요청에 따라 LLM이 노트 제목/본문/카테고리를 생성하고 저장합니다.

인자로 요청을 전달하거나, 인자 없이 실행하면 편집 폼으로 요청을 입력받습니다.

예시:
  ncli ai create "어제 회의 내용을 회의록으로 정리해줘"
  ncli ai create --dry-run "아이디어 브레인스톰 노트"
  ncli ai create --category work "오늘 한 일 정리"
  ncli ai create        # 요청을 폼에서 입력`,
	Run: func(ctx *wcli.Context) error {
		request := joinCreateRequest(ctx.Args)
		if request == "" {
			var input string
			if err := huh.NewText().Title("노트 생성 요청").Value(&input).Run(); err != nil {
				fmt.Println("작성이 취소되었습니다.")
				return nil
			}
			request = strings.TrimSpace(input)
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		client, err := appLLM.NewClient(cfg)
		if err != nil {
			return err
		}

		var result appLLM.CreateResult
		err = utils.WithSpinner("노트 생성 중...", func() error {
			var innerErr error
			result, innerErr = appLLM.CreateNote(context.Background(), client, cfg.LLM.Model, request)
			return innerErr
		})
		if err != nil {
			return err
		}

		category, err := resolveAICreateCategory(result.Category, aiCreateCategoryFlage)
		if err != nil {
			return err
		}

		draft := aiCreateDraft{Title: result.Title, Content: result.Content, Category: category}
		if aiCreateDryRun {
			fmt.Println(renderAICreateDryRun(draft))
			return nil
		}

		store, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = store.Close() }()

		created, err := store.CreateBoard(api.BoardCreate{Title: draft.Title, Content: draft.Content, Category: draft.Category})
		if err != nil {
			return fmt.Errorf("생성된 노트를 저장하지 못했습니다: %w", err)
		}

		fmt.Printf("노트를 생성했습니다. ID: %d\n", created.ID)
		return nil
	},
}

func init() {
	aiCreateCmd.Flags().BoolVar(&aiCreateDryRun, "dry-run", "", false, "노트를 저장하지 않고 생성 결과만 미리보기")
	aiCreateCmd.Flags().StringVar(&aiCreateCategoryFlage, "category", "", "", fmt.Sprintf("노트 카테고리 강제 지정 (미지정 시 LLM 제안 또는 %s)", defaultNoteCategory))
	aiCmd.AddCommand(aiCreateCmd)
}
