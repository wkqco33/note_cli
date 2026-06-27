package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var downloadCmd = &cobra.Command{
	Use:   "download [file_id]",
	Short: "파일 ID로 첨부파일 다운로드",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}

		var id int
		if len(args) == 1 {
			id, err = parseIDArg(args)
			if err != nil {
				return err
			}
		}

		var filename string
		if len(args) != 1 {
			files, err := client.GetFiles()
			if err != nil {
				return fmt.Errorf("파일 목록을 불러오지 못했습니다: %w", err)
			}

			if len(files) == 0 {
				fmt.Println("다운로드할 파일이 없습니다.")
				return nil
			}

			boards, err := client.GetBoards()
			if err != nil {
				return fmt.Errorf("노트 목록을 불러오지 못했습니다: %w", err)
			}

			id, err = selectFileID("다운로드할 파일을 선택하세요", files, boardTitleByFileName(boards))
			if err != nil {
				fmt.Println("취소되었습니다.")
				return nil
			}

			if file, ok := findFileByID(files, id); ok {
				filename = displayFileName(file)
			}
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("현재 작업 디렉토리를 가져올 수 없습니다: %w", err)
		}
		fmt.Printf("다운로드 중 (File ID: %d)...\n", id)
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
