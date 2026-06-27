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
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
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
				return nil
			}

			if target == "note" {
				notes, err := client.GetBoards()
				if err != nil {
					return fmt.Errorf("노트 목록을 불러오지 못했습니다: %w", err)
				}
				if len(notes) == 0 {
					fmt.Println("삭제할 노트가 없습니다.")
					return nil
				}

				id, err = selectBoardID("삭제할 노트를 선택하세요", notes)
				if err != nil {
					fmt.Println("취소되었습니다.")
					return nil
				}
			} else {
				files, err := client.GetFiles()
				if err != nil {
					return fmt.Errorf("파일 목록을 불러오지 못했습니다: %w", err)
				}
				if len(files) == 0 {
					fmt.Println("삭제할 파일이 없습니다.")
					return nil
				}

				boards, err := client.GetBoards()
				boardMap := map[string]string{}
				if err == nil {
					boardMap = boardTitleByFileName(boards)
				}

				id, err = selectFileID("삭제할 파일을 선택하세요", files, boardMap)
				if err != nil {
					fmt.Println("취소되었습니다.")
					return nil
				}
			}
		} else {
			id, err = parseIDArg(args)
			if err != nil {
				return err
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
			return nil
		}

		if target == "note" {
			if err := client.DeleteBoard(id); err != nil {
				return fmt.Errorf("노트를 삭제하지 못했습니다: %w", err)
			}
			fmt.Printf("노트를 삭제했습니다. ID: %d\n", id)
		} else {
			if err := client.DeleteFile(id); err != nil {
				return fmt.Errorf("파일을 삭제하지 못했습니다: %w", err)
			}
			fmt.Printf("파일을 삭제했습니다. ID: %d\n", id)
		}
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteFile, "file", "f", false, "파일 삭제 모드")
	rootCmd.AddCommand(deleteCmd)
}
