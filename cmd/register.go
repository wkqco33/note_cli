package cmd

import (
	"fmt"
	"note_cli/api"

	"github.com/charmbracelet/huh"
	"github.com/wkqco33/wcli"
)

var registerCmd = &wcli.Command{
	Use:   "register",
	Short: "새 사용자 계정 등록",
	Run: func(ctx *wcli.Context) error {
		var name, email, password string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("이름").
					Value(&name).
					Validate(requiredInput("이름을 입력해야 합니다")),
				huh.NewInput().
					Title("이메일").
					Value(&email).
					Validate(requiredInput("이메일을 입력해야 합니다")),
				huh.NewInput().
					Title("비밀번호").
					EchoMode(huh.EchoModePassword).
					Value(&password).
					Validate(requiredInput("비밀번호를 입력해야 합니다")),
			),
		)

		err := form.Run()
		if err != nil {
			fmt.Println("회원가입이 취소되었습니다.")
			return nil
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		user := api.UserCreate{
			Name:     name,
			Email:    email,
			Password: password,
		}

		if err := client.Register(user); err != nil {
			return fmt.Errorf("회원가입에 실패했습니다: %w", err)
		}

		fmt.Printf("회원가입이 완료되었습니다. 이제 '%s login'으로 로그인하세요.\n", binaryName())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
