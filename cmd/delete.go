package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var deleteFile bool

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "노트 또는 첨부파일 삭제",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil || cfg.AccessToken == "" {
			fmt.Println("Please login first using 'note login'")
			return
		}

		client := api.NewClient(cfg)

		target := "note"
		if deleteFile {
			target = "file"
		}

		var id int

		if len(args) == 0 {
			err = huh.NewSelect[string]().
				Title("어떤 항목을 삭제하시겠습니까?").
				Options(
					huh.NewOption("노트 (문서)", "note"),
					huh.NewOption("첨부 파일", "file"),
				).
				Value(&target).
				Run()
			if err != nil {
				fmt.Println("취소되었습니다.")
				return
			}

			if target == "note" {
				notes, err := client.GetBoards()
				if err != nil {
					fmt.Printf("Failed to get notes: %v\n", err)
					return
				}
				if len(notes) == 0 {
					fmt.Println("삭제할 노트가 없습니다.")
					return
				}
	
				var options []huh.Option[int]
				for _, note := range notes {
					title := note.Title
					if len(title) > 40 {
						title = title[:37] + "..."
					}
					label := fmt.Sprintf("[%d] %s", note.ID, title)
					options = append(options, huh.NewOption(label, note.ID))
				}
	
				err = huh.NewSelect[int]().
					Title("삭제할 노트를 선택하세요").
					Options(options...).
					Value(&id).
					Run()
				if err != nil {
					fmt.Println("취소되었습니다.")
					return
				}
			} else {
				files, err := client.GetFiles()
				if err != nil {
					fmt.Printf("Failed to get files: %v\n", err)
					return
				}
				if len(files) == 0 {
					fmt.Println("삭제할 파일이 없습니다.")
					return
				}

				boards, err := client.GetBoards()
				var boardMap = make(map[string]string)
				if err == nil {
					for _, b := range boards {
						for _, img := range b.Images {
							idx := strings.LastIndex(img, "/")
							if idx != -1 {
								boardMap[img[idx+1:]] = b.Title
							}
						}
					}
				}
	
				var options []huh.Option[int]
				for _, file := range files {
					name := file.OriginalFilename
					if name == "" {
						name = file.Filename
					}
					
					label := fmt.Sprintf("[%d] %s", file.ID, name)
					noteTitle := boardMap[file.Filename]
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
	
				err = huh.NewSelect[int]().
					Title("삭제할 파일을 선택하세요").
					Options(options...).
					Value(&id).
					Run()
				if err != nil {
					fmt.Println("취소되었습니다.")
					return
				}
			}
		} else {
			id, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Invalid ID: must be an integer")
				return
			}
		}

		confirm := false
		err = huh.NewConfirm().
			Title(fmt.Sprintf("정말로 ID %d 항목을 삭제하시겠습니까?", id)).
			Affirmative("예 (Delete)").
			Negative("아니오 (Cancel)").
			Value(&confirm).
			Run()
		
		if err != nil || !confirm {
			fmt.Println("삭제가 취소되었습니다.")
			return
		}

		if target == "note" {
			err = client.DeleteBoard(id)
			if err != nil {
				fmt.Printf("Failed to delete note: %v\n", err)
				return
			}
			fmt.Printf("Successfully deleted note %d\n", id)
		} else {
			err = client.DeleteFile(id)
			if err != nil {
				fmt.Printf("Failed to delete file: %v\n", err)
				return
			}
			fmt.Printf("Successfully deleted file %d\n", id)
		}
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteFile, "file", "f", false, "파일 삭제 모드")
	rootCmd.AddCommand(deleteCmd)
}
