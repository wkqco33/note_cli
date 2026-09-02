package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/tui"

	"github.com/wkqco33/wcli"
)

var attachedFiles []string

var addCmd = &wcli.Command{
	Use:   "add",
	Short: "새 노트 추가",
	Run: func(ctx *wcli.Context) error {
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

		if err := validateAttachedFiles(attachedFiles); err != nil {
			return err
		}

		var title, category string
		if err := runNoteForm(&title, &category); err != nil {
			fmt.Println("작성이 취소되었습니다.")
			return nil
		}

		fmt.Println("노트 내용을 편집기에서 작성합니다...")
		content, err := tui.OpenEditor("")
		if err != nil {
			return fmt.Errorf("편집기를 열지 못했습니다: %w", err)
		}

		if content == "" {
			fmt.Println("노트 내용이 비어 있어 생성을 취소했습니다.")
			return nil
		}

		imageUrls, err := uploadAttachedFiles(client, attachedFiles)
		if err != nil {
			return err
		}

		board := api.BoardCreate{
			Title:    title,
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
	addCmd.Flags().StringSliceVar(&attachedFiles, "file", "f", []string{}, "첨부할 파일 경로 (여러 개 지정 가능)")
	rootCmd.AddCommand(addCmd)
}
