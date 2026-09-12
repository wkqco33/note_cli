package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/tui"
	"os"

	"github.com/wkqco33/wcli"
)

var (
	addAttachedFiles []string
	addTitle         string
	addContent       string
	addContentFile   string
	addCategory      string
)

var addCmd = &wcli.Command{
	Use:   "add",
	Short: "새 노트 추가",
	Long: `새 노트를 추가합니다.

플래그를 지정하면 해당 항목의 폼/편집기 입력을 건너뜁니다.
제목(--title)과 내용(--content/--content-file)을 지정하면 대화형 입력 없이 즉시 생성되며,
카테고리를 지정하지 않으면 기본값(other)을 사용합니다.

예시:
  ncli add --title "회의 메모" --content-file meeting.md --category work
  echo "본문" | ncli add -t "할 일" -c -
  ncli add -t "메모" -c "내용" --no-input   # 완전 비대화형
  ncli add --title "제목"                    # 내용만 편집기에서 작성`,
	Run: func(ctx *wcli.Context) error {
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

		if err := validateAttachedFiles(addAttachedFiles); err != nil {
			return err
		}

		// 플래그로 지정된 내용 해석 (미지정 시 편집기 폴백)
		var content string
		if contentFlagsChanged(addContent, addContentFile) {
			content, err = resolveNoteContent(addContent, addContentFile, os.Stdin)
			if err != nil {
				return err
			}
		}

		// 카테고리 검증/정규화 (미지정 시 기본값)
		category, err := normalizeNoteCategory(addCategory)
		if err != nil {
			return err
		}

		// 미지정 필드만 폼/편집기로 입력받는다
		// 제목은 필수이므로 대화형이 아니면 플래그를 요구한다.
		// 카테고리는 기본값이 있으므로 대화형이 아니면 기본값을 그대로 사용한다.
		askTitle, askCategory := resolveNoteFormPrompts(addTitle, addCategory, promptAllowed())
		if addTitle == "" && !askTitle {
			return interactionRequired("--title/-t 플래그로 제목을 지정하세요")
		}

		if err := runNoteForm(askTitle, askCategory, &addTitle, &category); err != nil {
			return handlePromptError(err, "작성이 취소되었습니다.")
		}

		if !contentFlagsChanged(addContent, addContentFile) {
			if !promptAllowed() {
				return interactionRequired("내용을 지정하세요: --content/-c (stdin은 -c -), --content-file")
			}
			statusf("노트 내용을 편집기에서 작성합니다...")
			content, err = tui.OpenEditor("")
			if err != nil {
				return fmt.Errorf("편집기를 열지 못했습니다: %w", err)
			}
			if content == "" {
				fmt.Println("노트 내용이 비어 있어 생성을 취소했습니다.")
				return nil
			}
		}

		imageUrls, err := uploadAttachedFiles(client, addAttachedFiles)
		if err != nil {
			return err
		}

		board := api.BoardCreate{
			Title:    addTitle,
			Content:  content,
			Category: category,
			Images:   imageUrls,
		}

		createdBoard, err := client.CreateBoard(board)
		if err != nil {
			return fmt.Errorf("노트를 생성하지 못했습니다: %w", err)
		}

		fmt.Printf("노트를 생성했습니다. ID: %d\n", createdBoard.ID)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addTitle, "title", "t", "", "노트 제목 (지정 시 제목 폼 생략)")
	addCmd.Flags().StringVar(&addContent, "content", "c", "", "노트 내용 ('-' 지정 시 stdin에서 읽음, 지정 시 편집기 생략)")
	addCmd.Flags().StringVar(&addContentFile, "content-file", "", "", "노트 내용을 읽을 파일 경로 (--content와 동시 지정 불가)")
	addCmd.Flags().StringVar(&addCategory, "category", "", "", fmt.Sprintf("노트 카테고리 (미지정 시 %s)", defaultNoteCategory))
	addCmd.Flags().StringSliceVar(&addAttachedFiles, "file", "f", []string{}, "첨부할 파일 경로 (여러 개 지정 가능)")
	rootCmd.AddCommand(addCmd)
}
