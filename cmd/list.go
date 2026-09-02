package cmd

import (
	"fmt"

	"github.com/wkqco33/wcli"
)

var listCmd = &wcli.Command{
	Use:   "list",
	Short: "모든 노트 목록 조회",
	Run: func(ctx *wcli.Context) error {
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()
		notes, err := client.GetBoards()
		if err != nil {
			return fmt.Errorf("노트 목록을 불러오지 못했습니다: %w", err)
		}

		if len(notes) == 0 {
			fmt.Println("노트가 없습니다.")
			return nil
		}

		printBoardTable(notes)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
