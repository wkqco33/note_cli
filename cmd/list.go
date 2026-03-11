package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "모든 노트 목록 조회",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil || cfg.AccessToken == "" {
			fmt.Println("Please login first using 'note_cli login'")
			return
		}

		client := api.NewClient(cfg)
		notes, err := client.GetBoards()
		if err != nil {
			fmt.Printf("Failed to get notes: %v\n", err)
			return
		}

		if len(notes) == 0 {
			fmt.Println("No notes found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tUPDATED")
		for _, note := range notes {
			updatedStr := note.UpdatedAt
			if len(updatedStr) >= 16 {
				updatedStr = strings.Replace(updatedStr[:16], "T", " ", 1)
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", note.ID, note.Title, note.Category, updatedStr)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
