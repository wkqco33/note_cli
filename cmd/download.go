package cmd

import (
	"fmt"
	"os"

	"note_cli/api"

	"github.com/wkqco33/wcli"
)

var downloadCmd = &wcli.Command{
	Use:   "download [file_id]",
	Short: "파일 ID로 첨부파일 다운로드",
	Run: func(ctx *wcli.Context) error {
		if err := requireMaxArgs(ctx.Args, 1); err != nil {
			return err
		}
		args := ctx.Args
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

		var id int
		var filename string
		if len(args) == 1 {
			id, err = parseIDArg(args)
			if err != nil {
				return err
			}
		} else {
			var files []api.FileRead
			var ok bool
			id, files, ok, err = resolveFileID(client, "다운로드할 파일을 선택하세요", "다운로드할 파일이 없습니다.")
			if err != nil || !ok {
				return err
			}

			if file, found := findFileByID(files, id); found {
				filename = displayFileName(file)
			}
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("현재 작업 디렉토리를 가져올 수 없습니다: %w", err)
		}
		statusf("다운로드 중 (File ID: %d)...", id)
		destPath, err := client.DownloadFile(id, cwd, filename)
		if err != nil {
			return fmt.Errorf("다운로드 실패: %w", err)
		}

		fmt.Printf("다운로드 성공! 위치: %s\n", destPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
