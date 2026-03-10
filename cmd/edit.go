package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"
	"note_cli/tui"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit an existing note",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid ID: must be an integer")
			return
		}

		cfg, err := config.Load()
		if err != nil || cfg.AccessToken == "" {
			fmt.Println("Please login first using 'note login'")
			return
		}

		client := api.NewClient(cfg)
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

		update := api.BoardUpdate{
			Title:    &title,
			Content:  &content,
			Category: &category,
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
	rootCmd.AddCommand(editCmd)
}
