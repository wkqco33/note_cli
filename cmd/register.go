package cmd

import (
	"fmt"
	"note_cli/api"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "새 사용자 계정 등록",
	Run: func(cmd *cobra.Command, args []string) {
		var name, email, password string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("이름").
					Value(&name).
					Validate(func(str string) error {
						if str == "" {
							return fmt.Errorf("이름을 입력해야 합니다")
						}
						return nil
					}),
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
			fmt.Println("회원가입이 취소되었습니다.")
			return
		}

		client, err := newClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		user := api.UserCreate{
			Name:     name,
			Email:    email,
			Password: password,
		}

		err = client.Register(user)
		if err != nil {
			fmt.Printf("회원가입에 실패했습니다: %v\n", err)
			return
		}

		fmt.Println("회원가입이 완료되었습니다. 이제 'note_cli login'으로 로그인하세요.")
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
