package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"
	"strconv"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "ID로 노트 삭제",
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
		err = client.DeleteBoard(id)
		if err != nil {
			fmt.Printf("Failed to delete note: %v\n", err)
			return
		}

		fmt.Printf("Successfully deleted note %d\n", id)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
