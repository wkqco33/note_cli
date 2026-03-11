package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"
	"note_cli/tui"
	"os"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var editAttachedFiles []string

var editCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "기존 노트 수정",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, f := range editAttachedFiles {
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

		cfg, err := config.Load()
		if err != nil || cfg.AccessToken == "" {
			fmt.Println("Please login first using 'note_cli login'")
			return
		}

		client := api.NewClient(cfg)

		var id int
		if len(args) == 1 {
			var err error
			id, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Invalid ID: must be an integer")
				return
			}
		} else {
			notes, err := client.GetBoards()
			if err != nil {
				fmt.Printf("Failed to get notes: %v\n", err)
				return
			}
			if len(notes) == 0 {
				fmt.Println("No notes found to edit.")
				return
			}

			var options []huh.Option[int]
			for _, note := range notes {
				title := note.Title
				if len(title) > 40 {
					title = title[:37] + "..."
				}
				label := fmt.Sprintf("[%d] %s", note.ID, title)
				options = append(options, huh.NewOption(label, note.ID))
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[int]().
						Title("수정할 노트를 선택하세요").
						Options(options...).
						Value(&id),
				),
			)
			if err := form.Run(); err != nil {
				fmt.Println("취소되었습니다.")
				return
			}
		}
		note, err := client.GetBoard(id)
		if err != nil {
			fmt.Printf("Failed to get note: %v\n", err)
			return
		}

		title := note.Title
		category := note.Category

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
		content, err := tui.OpenEditor(note.Content)
		if err != nil {
			fmt.Printf("Error opening editor: %v\n", err)
			return
		}

		var imageUrls []string
		if len(note.Images) > 0 {
			imageUrls = append(imageUrls, note.Images...)
		}
		
		if len(editAttachedFiles) > 0 {
			for _, f := range editAttachedFiles {
				fmt.Printf("파일 첨부 중: %s\n", f)
				uploaded, err := client.UploadFile(f)
				if err != nil {
					fmt.Printf("업로드 실패 (%s): %v\n", f, err)
					return
				}
				imageUrls = append(imageUrls, uploaded.URL)
				fmt.Println("업로드 완료!")
			}
		}

		update := api.BoardUpdate{
			Title:    &title,
			Content:  &content,
			Category: &category,
			Images:   &imageUrls,
		}

		updatedBoard, err := client.UpdateBoard(id, update)
		if err != nil {
			fmt.Printf("Failed to update note: %v\n", err)
			return
		}

		fmt.Printf("Successfully updated note %d\n", updatedBoard.ID)
	},
}

func init() {
	editCmd.Flags().StringSliceVarP(&editAttachedFiles, "file", "f", []string{}, "추가로 첨부할 파일 경로 (여러 개 지정 가능)")
	rootCmd.AddCommand(editCmd)
}
