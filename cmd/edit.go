package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/tui"

	"github.com/spf13/cobra"
)

var editAttachedFiles []string

var editCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "기존 노트 수정",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateAttachedFiles(editAttachedFiles); err != nil {
			return err
		}

		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}

		id, ok, err := resolveBoardID(client, args, "수정할 노트를 선택하세요", "수정할 노트가 없습니다.")
		if err != nil || !ok {
			return err
		}
		note, err := client.GetBoard(id)
		if err != nil {
			return fmt.Errorf("노트를 불러오지 못했습니다: %w", err)
		}

		title := note.Title
		category := note.Category
		if err := runNoteForm(&title, &category); err != nil {
			fmt.Println("수정이 취소되었습니다.")
			return nil
		}

		fmt.Println("노트 내용을 편집기에서 수정합니다...")
		content, err := tui.OpenEditor(note.Content)
		if err != nil {
			return fmt.Errorf("편집기를 열지 못했습니다: %w", err)
		}

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
	editCmd.Flags().StringSliceVarP(&editAttachedFiles, "file", "f", []string{}, "추가로 첨부할 파일 경로 (여러 개 지정 가능)")
	rootCmd.AddCommand(editCmd)
}
