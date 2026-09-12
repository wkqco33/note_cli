package cmd

import (
	"fmt"
	"os"

	"note_cli/api"

	"github.com/charmbracelet/huh"
	"github.com/wkqco33/wcli"
)

var (
	registerName          string
	registerEmail         string
	registerPassword      string
	registerPasswordStdin bool
	registerPasswordFile  string
)

// registerFlagsHint 프롬프트 없이 회원가입에 필요한 플래그 안내
func registerFlagsHint() string {
	return "--name, --email, --password (또는 --password-stdin) 플래그를 지정하세요"
}

var registerCmd = &wcli.Command{
	Use:   "register",
	Short: "새 사용자 계정 등록",
	Long: `새 사용자 계정을 등록합니다.

플래그를 지정하면 대화형 폼 없이 가입합니다.
에이전트/CI에서는 --name, --email, --password (또는 --password-stdin)을 사용하세요.

예시:
  ncli register
  ncli register --name 홍길동 --email user@example.com --password-file ~/.secrets/pw
  cat pw.txt | ncli register --name 홍길동 --email user@example.com --password-stdin`,
	Run: func(ctx *wcli.Context) error {
		password, err := resolvePasswordFlag(registerPassword, registerPasswordStdin, registerPasswordFile, os.Stdin)
		if err != nil {
			return err
		}
		warnInsecurePassword(registerPassword)
		name, email := registerName, registerEmail

		if name == "" || email == "" || password == "" {
			if !promptAllowed() {
				return interactionRequired(registerFlagsHint())
			}

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

			if err := form.Run(); err != nil {
				return handlePromptError(err, "회원가입이 취소되었습니다.")
			}
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
	registerCmd.Flags().StringVar(&registerName, "name", "", "", "이름 (지정 시 폼 생략)")
	registerCmd.Flags().StringVar(&registerEmail, "email", "", "", "이메일 (지정 시 폼 생략)")
	registerCmd.Flags().StringVar(&registerPassword, "password", "", "", "비밀번호 (ps/셸 히스토리 노출 위험, --password-stdin/--password-file 권장)")
	registerCmd.Flags().BoolVar(&registerPasswordStdin, "password-stdin", "", false, "stdin에서 비밀번호를 읽음")
	registerCmd.Flags().StringVar(&registerPasswordFile, "password-file", "", "", "비밀번호를 읽을 파일 경로")
	rootCmd.AddCommand(registerCmd)
}
