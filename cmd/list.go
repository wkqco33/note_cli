package cmd

import (
	"fmt"

	"note_cli/api"
	"note_cli/utils"

	"github.com/wkqco33/wcli"
)

var (
	listFormat string
	listJSON   bool
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
		var notes []api.BoardRead
		err = utils.WithSpinner("노트를 불러오는 중...", func() error {
			var innerErr error
			notes, innerErr = client.GetBoards()
			return innerErr
		})
		if err != nil {
			return fmt.Errorf("노트 목록을 불러오지 못했습니다: %w", err)
		}

		format, err := outputFormatFromFlags(listFormat, listJSON)
		if err != nil {
			return err
		}
		if len(notes) == 0 && format == formatText {
			fmt.Println("노트가 없습니다.")
			return nil
		}

		rendered, err := renderNotes(notes, format)
		if err != nil {
			return err
		}
		printRenderedNotes(rendered, format == formatText)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVar(&listFormat, "format", "", "", "출력 형식 지정 (text, json, yaml)")
	listCmd.Flags().BoolVar(&listJSON, "json", "", false, "JSON 형식으로 출력 (--format json과 동일)")
}
