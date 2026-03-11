package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"
	"note_cli/tui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "새 노트 추가",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil || cfg.AccessToken == "" {
			fmt.Println("Please login first using 'note login'")
			return
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
		board := api.BoardCreate{
			Title:    title,
			Content:  content,
			Category: category,
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
	rootCmd.AddCommand(addCmd)
}
