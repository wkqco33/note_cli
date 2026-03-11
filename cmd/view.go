package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view [id]",
	Short: "ID로 노트 조회",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil || cfg.AccessToken == "" {
			fmt.Println("Please login first using 'note login'")
			return
		}

		client := api.NewClient(cfg)

		var id int
		if len(args) == 1 {
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
				fmt.Println("No notes found to view.")
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
						Title("조회할 노트를 선택하세요").
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

		updatedStr := note.UpdatedAt
		if len(updatedStr) >= 19 {
			updatedStr = strings.Replace(updatedStr[:19], "T", " ", 1)
		}
		fmt.Printf("=== %s ===\n", note.Title)
		fmt.Printf("Category: %s | Updated: %s\n", note.Category, updatedStr)
		fmt.Println("--------------------------------------------------")
		fmt.Println(note.Content)
		fmt.Println("--------------------------------------------------")
		
		if len(note.Images) > 0 {
			fmt.Println("[첨부파일 목록]")
			for i, urlStr := range note.Images {
				idx := strings.LastIndex(urlStr, "/")
				filename := urlStr
				if idx != -1 {
					filename = urlStr[idx+1:]
				}
				fmt.Printf("%d. %s (%s)\n", i+1, filename, urlStr)
			}
			fmt.Println("--------------------------------------------------")
		}
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
