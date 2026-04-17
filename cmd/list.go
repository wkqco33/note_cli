package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "모든 노트 목록 조회",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := newAuthenticatedClient()
		if err != nil {
			fmt.Println(err)
			return
		}
		notes, err := client.GetBoards()
		if err != nil {
			fmt.Printf("노트 목록을 불러오지 못했습니다: %v\n", err)
			return
		}

		if len(notes) == 0 {
			fmt.Println("노트가 없습니다.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tUPDATED")
		for _, note := range notes {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", note.ID, truncateText(note.Title, 40), note.Category, formatTimestamp(note.UpdatedAt, 16))
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
