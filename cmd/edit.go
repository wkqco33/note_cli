package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/tui"
	"os"

	"github.com/wkqco33/wcli"
)

var (
	editAttachedFiles []string
	editTitle         string
	editContent       string
	editContentFile   string
	editCategory      string
)

var editCmd = &wcli.Command{
	Use:   "edit [id]",
	Short: "기존 노트 수정",
	Long: `기존 노트를 수정합니다.

플래그를 지정하면 해당 항목의 폼/편집기 입력을 건너뛰고 지정한 값으로 수정합니다.
지정하지 않은 필드는 기존 값이 유지됩니다. 비대화형 환경에서는 지정하지 않은 필드를
묻지 않고 기존 값을 그대로 유지합니다.

예시:
  ncli edit 3 --title "수정된 제목"          # 제목만 변경
  ncli edit 3 --content-file new_content.md  # 내용만 변경
  echo "새 내용" | ncli edit 3 -c -          # stdin으로 내용 변경 (완전 비대화형)`,
	Run: func(ctx *wcli.Context) error {
		if err := requireMaxArgs(ctx.Args, 1); err != nil {
			return err
		}
		args := ctx.Args
		if err := validateAttachedFiles(editAttachedFiles); err != nil {
			return err
		}

		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

		id, ok, err := resolveBoardID(client, args, "수정할 노트를 선택하세요", "수정할 노트가 없습니다.")
		if err != nil || !ok {
			return err
		}
		note, err := client.GetBoard(id)
		if err != nil {
			return fmt.Errorf("노트를 불러오지 못했습니다: %w", err)
		}

		// 플래그로 지정된 내용 해석 (미지정 시 편집기 폴백)
		var flagContent string
		hasContentFlag := contentFlagsChanged(editContent, editContentFile)
		if hasContentFlag {
			flagContent, err = resolveNoteContent(editContent, editContentFile, os.Stdin)
			if err != nil {
				return err
			}
		}

		// 카테고리 검증/정규화 (미지정 시 기존 값 유지)
		var category string
		if editCategory == "" {
			category = note.Category
		} else {
			category, err = normalizeNoteCategory(editCategory)
			if err != nil {
				return err
			}
		}

		title := note.Title
		if editTitle != "" {
			title = editTitle
		}

		// 미지정 필드만 폼/편집기로 입력받는다.
		// 비대화형이면 미지정 필드는 기존 값을 유지한다 (플래그 미지정 = 기존 값 유지).
		askTitle, askCategory := resolveNoteFormPrompts(editTitle, editCategory, promptAllowed())

		if err := runNoteForm(askTitle, askCategory, &title, &category); err != nil {
			return handlePromptError(err, "수정이 취소되었습니다.")
		}

		content := note.Content
		if hasContentFlag {
			content = flagContent
		} else {
			if !promptAllowed() {
				return interactionRequired("내용을 지정하세요: --content/-c (stdin은 -c -), --content-file")
			}
			statusf("노트 내용을 편집기에서 수정합니다...")
			content, err = tui.OpenEditor(note.Content)
			if err != nil {
				return fmt.Errorf("편집기를 열지 못했습니다: %w", err)
			}
		}

		provided := providedNoteFields{
			Title:    &title,
			Category: &category,
			Content:  &content,
		}
		title, category, content = mergeNoteFields(note, provided)

		var imageUrls []string
		if len(note.Images) > 0 {
			imageUrls = append(imageUrls, note.Images...)
		}

		newUrls, err := uploadAttachedFiles(client, editAttachedFiles)
		if err != nil {
			return err
		}
		imageUrls = append(imageUrls, newUrls...)

		update := api.BoardUpdate{
			Title:    &title,
			Content:  &content,
			Category: &category,
			Images:   &imageUrls,
		}

		updatedBoard, err := client.UpdateBoard(id, update)
		if err != nil {
			return fmt.Errorf("노트를 수정하지 못했습니다: %w", err)
		}

		fmt.Printf("노트를 수정했습니다. ID: %d\n", updatedBoard.ID)
		return nil
	},
}

func init() {
	editCmd.Flags().StringVar(&editTitle, "title", "t", "", "수정할 노트 제목 (지정 시 제목 폼 생략)")
	editCmd.Flags().StringVar(&editContent, "content", "c", "", "수정할 노트 내용 ('-' 지정 시 stdin에서 읽음, 지정 시 편집기 생략)")
	editCmd.Flags().StringVar(&editContentFile, "content-file", "", "", "수정할 내용을 읽을 파일 경로 (--content와 동시 지정 불가)")
	editCmd.Flags().StringVar(&editCategory, "category", "", "", "수정할 노트 카테고리 (미지정 시 기존 값 유지)")
	editCmd.Flags().StringSliceVar(&editAttachedFiles, "file", "f", []string{}, "추가로 첨부할 파일 경로 (여러 개 지정 가능)")
	rootCmd.AddCommand(editCmd)
}
