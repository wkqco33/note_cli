package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/wkqco33/wcli"
)

var deleteFile bool

var deleteCmd = &wcli.Command{
	Use:   "delete [id]",
	Short: "노트 또는 첨부파일 삭제",
	Long: `노트 또는 첨부파일을 삭제합니다.

ID를 지정하면 해당 항목을, 생략하면 대화형 목록에서 선택합니다.
삭제는 확인 프롬프트를 거치며, 비대화형 환경에서는 --yes/-y가 필요합니다.

예시:
  ncli delete 3
  ncli delete 3 --yes        # 확인 없이 삭제 (비대화형/CI)
  ncli delete --file 2 --yes # 첨부파일 삭제`,
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

		target := "note"
		if deleteFile {
			target = "file"
		}

		var id int

		if len(args) == 0 {
			if err := requirePrompt("삭제 대상을 지정하세요: 노트는 '" + binaryName() + " delete <ID>', 파일은 '" + binaryName() + " delete --file <ID>'"); err != nil {
				return err
			}

			err = huh.NewSelect[string]().
				Title("어떤 항목을 삭제하시겠습니까?").
				Options(
					huh.NewOption("노트 (문서)", "note"),
					huh.NewOption("첨부 파일", "file"),
				).
				Value(&target).
				Run()
			if err != nil {
				return handlePromptError(err, "취소되었습니다.")
			}

			if target == "note" {
				var ok bool
				id, ok, err = resolveBoardID(client, nil, "삭제할 노트를 선택하세요", "삭제할 노트가 없습니다.")
				if err != nil || !ok {
					return err
				}
			} else {
				var ok bool
				id, _, ok, err = resolveFileID(client, "삭제할 파일을 선택하세요", "삭제할 파일이 없습니다.")
				if err != nil || !ok {
					return err
				}
			}
		} else {
			id, err = parseIDArg(args)
			if err != nil {
				return err
			}
		}

		confirm, err := askConfirm(
			fmt.Sprintf("정말로 ID %d 항목을 삭제하시겠습니까?", id),
			"예 (삭제)",
			"아니오 (취소)",
			"삭제하려면 --yes/-y 플래그를 지정하세요",
		)
		if err != nil {
			return err
		}
		if !confirm {
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
	deleteCmd.Flags().BoolVar(&deleteFile, "file", "f", false, "파일 삭제 모드")
	rootCmd.AddCommand(deleteCmd)
}
