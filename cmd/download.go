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
	Run: func(cmd *cobra.Command, args []string) {
		client, err := newAuthenticatedClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		var id int
		if len(args) == 1 {
			id, err = parseIDArg(args)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		var filename string
		if len(args) != 1 {
			files, err := client.GetFiles()
			if err != nil {
				fmt.Printf("파일 목록을 불러오지 못했습니다: %v\n", err)
				return
			}

			if len(files) == 0 {
				fmt.Println("다운로드할 파일이 없습니다.")
				return
			}

			boards, err := client.GetBoards()
			if err != nil {
				fmt.Printf("노트 목록을 불러오지 못했습니다: %v\n", err)
				return
			}

			id, err = selectFileID("다운로드할 파일을 선택하세요", files, boardTitleByFileName(boards))
			if err != nil {
				fmt.Println("취소되었습니다.")
				return
			}

			if file, ok := findFileByID(files, id); ok {
				filename = displayFileName(file)
			}
		}

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("현재 작업 디렉토리를 가져올 수 없습니다: %v\n", err)
			return
		}
		fmt.Printf("다운로드 중 (File ID: %d)...\n", id)
		destPath, err := client.DownloadFile(id, cwd, filename)
		if err != nil {
			fmt.Printf("다운로드 실패: %v\n", err)
			return
		}

		fmt.Printf("다운로드 성공! 위치: %s\n", destPath)
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
