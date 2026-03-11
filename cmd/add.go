package cmd

import (
	"fmt"
	"os"
	"note_cli/api"
	"note_cli/config"
	"note_cli/tui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var attachedFiles []string

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "새 노트 추가",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil || cfg.AccessToken == "" {
			fmt.Println("Please login first using 'note login'")
			return
		}

		for _, f := range attachedFiles {
			stat, err := os.Stat(f)
			if err != nil {
				fmt.Printf("파일을 찾을 수 없습니다: %s\n", f)
				return
			}
			if stat.Size() > 500*1024*1024 {
				fmt.Printf("500MB 제한 초과 파일이 포함되어 있습니다: %s\n", f)
				return
			}
		}

		var title, category string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Title").
					Value(&title).
					Validate(func(str string) error {
						if str == "" {
							return fmt.Errorf("title is required")
						}
						return nil
					}),
				huh.NewSelect[string]().
					Title("Category").
					Options(
						huh.NewOption("Work", "work"),
						huh.NewOption("Personal", "personal"),
						huh.NewOption("Idea", "idea"),
						huh.NewOption("Other", "other"),
					).
					Value(&category),
			),
		)

		if err := form.Run(); err != nil {
			fmt.Println("Cancelled.")
			return
		}

		fmt.Println("Opening editor for note content...")
		content, err := tui.OpenEditor("")
		if err != nil {
			fmt.Printf("Error opening editor: %v\n", err)
			return
		}

		if content == "" {
			fmt.Println("Note content is empty, cancelling.")
			return
		}

		client := api.NewClient(cfg)
		
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
			fmt.Printf("Failed to create note: %v\n", err)
			return
		}

		fmt.Printf("Successfully created note with ID %d!\n", createdBoard.ID)
	},
}

func init() {
	addCmd.Flags().StringSliceVarP(&attachedFiles, "file", "f", []string{}, "첨부할 파일 경로 (여러 개 지정 가능)")
	rootCmd.AddCommand(addCmd)
}
