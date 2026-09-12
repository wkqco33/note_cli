package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/wkqco33/wcli"
)

var (
	loginEmail         string
	loginPassword      string
	loginPasswordStdin bool
	loginPasswordFile  string
)

var loginCmd = &wcli.Command{
	Use:   "login",
	Short: "Note API 로그인",
	Long: `Note API에 로그인합니다.

플래그를 지정하면 대화형 폼 없이 로그인합니다.
에이전트/CI에서는 --email과 --password (또는 --password-stdin)을 사용하세요.

예시:
  ncli login
  ncli login --email user@example.com --password-file ~/.secrets/pw
  cat pw.txt | ncli login --email user@example.com --password-stdin`,
	Run: func(ctx *wcli.Context) error {
		password, err := resolvePasswordFlag(loginPassword, loginPasswordStdin, loginPasswordFile, os.Stdin)
		if err != nil {
			return err
		}
		warnInsecurePassword(loginPassword)
		email := loginEmail

		switch resolveCredentialSource(email, password, promptAllowed()) {
		case credentialsFromPrompt:
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

			if err := form.Run(); err != nil {
				return handlePromptError(err, "로그인이 취소되었습니다.")
			}
		case credentialsUnavailable:
			return interactionRequired(credentialsRequiredHint())
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
	loginCmd.Flags().StringVar(&loginEmail, "email", "", "", "로그인 이메일 (지정 시 폼 생략)")
	loginCmd.Flags().StringVar(&loginPassword, "password", "", "", "로그인 비밀번호 (ps/셸 히스토리 노출 위험, --password-stdin/--password-file 권장)")
	loginCmd.Flags().BoolVar(&loginPasswordStdin, "password-stdin", "", false, "stdin에서 비밀번호를 읽음")
	loginCmd.Flags().StringVar(&loginPasswordFile, "password-file", "", "", "비밀번호를 읽을 파일 경로")
	rootCmd.AddCommand(loginCmd)
}
