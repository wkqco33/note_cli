package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view [id]",
	Short: "View a note by ID",
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

		updatedStr := note.UpdatedAt
		if len(updatedStr) >= 19 {
			updatedStr = strings.Replace(updatedStr[:19], "T", " ", 1)
		}
		fmt.Printf("=== %s ===\n", note.Title)
		fmt.Printf("Category: %s | Updated: %s\n", note.Category, updatedStr)
		fmt.Println("--------------------------------------------------")
		fmt.Println(note.Content)
		fmt.Println("--------------------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
