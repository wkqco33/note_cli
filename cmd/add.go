package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/tui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var attachedFiles []string

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "새 노트 추가",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := newAuthenticatedClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := validateAttachedFiles(attachedFiles); err != nil {
			fmt.Println(err)
			return
		}

		var title, category string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("제목").
					Value(&title).
					Validate(func(str string) error {
						if str == "" {
							return fmt.Errorf("제목을 입력해야 합니다")
						}
						return nil
					}),
				huh.NewSelect[string]().
					Title("카테고리").
					Options(categorySelectOptions()...).
					Value(&category),
			),
		)

		if err := form.Run(); err != nil {
			fmt.Println("작성이 취소되었습니다.")
			return
		}

		fmt.Println("노트 내용을 편집기에서 작성합니다...")
		content, err := tui.OpenEditor("")
		if err != nil {
			fmt.Printf("편집기를 열지 못했습니다: %v\n", err)
			return
		}

		if content == "" {
			fmt.Println("노트 내용이 비어 있어 생성을 취소했습니다.")
			return
		}

		var imageUrls []string
		for _, f := range attachedFiles {
			fmt.Printf("파일 첨부 중: %s\n", f)
			uploaded, err := client.UploadFile(f)
			if err != nil {
				fmt.Printf("업로드 실패 (%s): %v\n", f, err)
				return
			}
			imageUrls = append(imageUrls, uploaded.URL)
			fmt.Println("업로드 완료!")
		}

		board := api.BoardCreate{
			Title:    title,
			Content:  content,
			Category: category,
			Images:   imageUrls,
		}

		createdBoard, err := client.CreateBoard(board)
		if err != nil {
			fmt.Printf("노트를 생성하지 못했습니다: %v\n", err)
			return
		}

		fmt.Printf("노트를 생성했습니다. ID: %d\n", createdBoard.ID)
	},
}

func init() {
	addCmd.Flags().StringSliceVarP(&attachedFiles, "file", "f", []string{}, "첨부할 파일 경로 (여러 개 지정 가능)")
	rootCmd.AddCommand(addCmd)
}
