package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Note API 로그인",
	Run: func(cmd *cobra.Command, args []string) {
		var email, password string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("이메일").
					Value(&email).
					Validate(func(str string) error {
						if str == "" {
							return fmt.Errorf("이메일을 입력해야 합니다")
						}
						return nil
					}),
				huh.NewInput().
					Title("비밀번호").
					EchoMode(huh.EchoModePassword).
					Value(&password).
					Validate(func(str string) error {
						if str == "" {
							return fmt.Errorf("비밀번호를 입력해야 합니다")
						}
						return nil
					}),
			),
		)

		err := form.Run()
		if err != nil {
			fmt.Println("로그인이 취소되었습니다.")
			return
		}

		client, err := newClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		err = client.Login(email, password)
		if err != nil {
			fmt.Printf("로그인에 실패했습니다: %v\n", err)
			return
		}

		fmt.Println("로그인했습니다.")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
