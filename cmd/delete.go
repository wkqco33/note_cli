package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var deleteFile bool

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "노트 또는 첨부파일 삭제",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := newAuthenticatedClient()
		if err != nil {
			fmt.Println(err)
			return
		}

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
					fmt.Printf("노트 목록을 불러오지 못했습니다: %v\n", err)
					return
				}
				if len(notes) == 0 {
					fmt.Println("삭제할 노트가 없습니다.")
					return
				}

				id, err = selectBoardID("삭제할 노트를 선택하세요", notes)
				if err != nil {
					fmt.Println("취소되었습니다.")
					return
				}
			} else {
				files, err := client.GetFiles()
				if err != nil {
					fmt.Printf("파일 목록을 불러오지 못했습니다: %v\n", err)
					return
				}
				if len(files) == 0 {
					fmt.Println("삭제할 파일이 없습니다.")
					return
				}

				boards, err := client.GetBoards()
				boardMap := map[string]string{}
				if err == nil {
					boardMap = boardTitleByFileName(boards)
				}

				id, err = selectFileID("삭제할 파일을 선택하세요", files, boardMap)
				if err != nil {
					fmt.Println("취소되었습니다.")
					return
				}
			}
		} else {
			id, err = parseIDArg(args)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		confirm := false
		err = huh.NewConfirm().
			Title(fmt.Sprintf("정말로 ID %d 항목을 삭제하시겠습니까?", id)).
			Affirmative("예 (삭제)").
			Negative("아니오 (취소)").
			Value(&confirm).
			Run()

		if err != nil || !confirm {
			fmt.Println("삭제가 취소되었습니다.")
			return
		}

		if target == "note" {
			err = client.DeleteBoard(id)
			if err != nil {
				fmt.Printf("노트를 삭제하지 못했습니다: %v\n", err)
				return
			}
			fmt.Printf("노트를 삭제했습니다. ID: %d\n", id)
		} else {
			err = client.DeleteFile(id)
			if err != nil {
				fmt.Printf("파일을 삭제하지 못했습니다: %v\n", err)
				return
			}
			fmt.Printf("파일을 삭제했습니다. ID: %d\n", id)
		}
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteFile, "file", "f", false, "파일 삭제 모드")
	rootCmd.AddCommand(deleteCmd)
}
