package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Note API 로그인",
	RunE: func(cmd *cobra.Command, args []string) error {
		var email, password string

		form := huh.NewForm(
			huh.NewGroup(
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
			fmt.Println("로그인이 취소되었습니다.")
			return nil
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		// 자동 로그인이 켜져 있으면 계정 정보도 함께 저장
		if client.Config.AutoLogin {
			client.Config.Username = email
			client.Config.Password = password
		}

		if err := client.Login(email, password); err != nil {
			return fmt.Errorf("로그인에 실패했습니다: %w", err)
		}

		fmt.Println("로그인했습니다.")
		if client.Config.AutoLogin {
			fmt.Println("자동 로그인용 계정 정보를 저장했습니다.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
