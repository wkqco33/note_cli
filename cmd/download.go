package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [file_id]",
	Short: "파일 ID로 첨부파일 다운로드",
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
		}

		var filename string
		if len(args) != 1 {
			files, err := client.GetFiles()
			if err != nil {
				fmt.Printf("Failed to get files: %v\n", err)
				return
			}

			if len(files) == 0 {
				fmt.Println("No files found to download.")
				return
			}

			boards, err := client.GetBoards()
			if err != nil {
				fmt.Printf("Failed to get notes: %v\n", err)
				return
			}

			fileNoteMap := make(map[string]string)
			for _, board := range boards {
				for _, imgUrl := range board.Images {
					idx := strings.LastIndex(imgUrl, "/")
					fname := imgUrl
					if idx != -1 {
						fname = imgUrl[idx+1:]
					}
					fileNoteMap[fname] = board.Title
				}
			}

			var options []huh.Option[int]
			for _, file := range files {
				name := file.OriginalFilename
				if name == "" {
					name = file.Filename
				}
				
				label := fmt.Sprintf("[%d] %s", file.ID, name)
				noteTitle := fileNoteMap[file.Filename]
				if noteTitle != "" {
					if len(noteTitle) > 15 {
						noteTitle = noteTitle[:12] + "..."
					}
					label += fmt.Sprintf(" (Note: %s, Size: %s)", noteTitle, humanize.Bytes(uint64(file.FileSize)))
				} else {
					label += fmt.Sprintf(" (Size: %s)", humanize.Bytes(uint64(file.FileSize)))
				}
				
				options = append(options, huh.NewOption(label, file.ID))
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[int]().
						Title("다운로드할 파일을 선택하세요").
						Options(options...).
						Value(&id),
				),
			)
			if err := form.Run(); err != nil {
				fmt.Println("취소되었습니다.")
				return
			}

			for _, f := range files {
				if f.ID == id {
					filename = f.OriginalFilename
					if filename == "" {
						filename = f.Filename
					}
					break
				}
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
