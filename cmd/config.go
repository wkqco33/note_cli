package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"note_cli/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "현재 설정 조회 및 수정",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		path, err := config.Path()
		if err != nil {
			return fmt.Errorf("설정 파일 경로를 확인하지 못했습니다: %w", err)
		}

		loginStatus := "로그아웃 상태"
		if cfg.AccessToken != "" {
			loginStatus = "로그인됨"
		}

		autoLogin := "꺼짐"
		if cfg.AutoLogin {
			autoLogin = "켜짐"
			if cfg.Username == "" {
				autoLogin += fmt.Sprintf(" (계정 미저장, '%s login' 필요)", binaryName())
			}
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintf(w, "mode\t%s\n", cfg.Mode)
		_, _ = fmt.Fprintf(w, "host\t%s\n", cfg.Host)
		_, _ = fmt.Fprintf(w, "port\t%d\n", cfg.Port)
		_, _ = fmt.Fprintf(w, "auto_login\t%s\n", autoLogin)
		if cfg.Username != "" {
			_, _ = fmt.Fprintf(w, "계정\t%s\n", cfg.Username)
		}
		_, _ = fmt.Fprintf(w, "로그인\t%s\n", loginStatus)
		_, _ = fmt.Fprintf(w, "설정 파일\t%s\n", path)
		_ = w.Flush()
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "설정 값 수정 (mode, host, port, auto_login)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		key, value := args[0], args[1]
		if err := applyConfigValue(cfg, key, value); err != nil {
			return err
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("설정을 저장하지 못했습니다: %w", err)
		}

		fmt.Printf("%s = %s (으)로 설정했습니다.\n", key, value)
		if key == "auto_login" {
			if cfg.AutoLogin {
				fmt.Printf("다음 '%s login' 성공 시 계정 정보가 저장되어 인증 만료 시 자동으로 재로그인합니다.\n", binaryName())
				fmt.Println("비밀번호는 암호화되어 설정 파일에 저장됩니다.")
			} else {
				fmt.Println("저장된 자동 로그인 계정 정보를 삭제했습니다.")
			}
		}
		return nil
	},
}

func applyConfigValue(cfg *config.Config, key, value string) error {
	switch key {
	case "mode":
		if value != "local" && value != "remote" {
			return fmt.Errorf("mode 값은 local 또는 remote여야 합니다")
		}
		cfg.Mode = value
	case "host":
		if value == "" {
			return fmt.Errorf("host 값은 비워둘 수 없습니다")
		}
		cfg.Host = value
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("port는 1~65535 사이의 정수여야 합니다")
		}
		cfg.Port = port
	case "auto_login":
		enabled, err := parseOnOff(value)
		if err != nil {
			return fmt.Errorf("auto_login 값은 true/false 또는 on/off여야 합니다")
		}
		cfg.AutoLogin = enabled
		if !enabled {
			// 자동 로그인을 끄면 저장된 계정 정보도 함께 삭제
			cfg.Username = ""
			cfg.Password = ""
		}
	case "access_token", "refresh_token":
		return fmt.Errorf("토큰은 직접 수정할 수 없습니다. '%s login'을 사용하세요", binaryName())
	case "username", "password":
		return fmt.Errorf("계정 정보는 직접 수정할 수 없습니다. auto_login을 켠 뒤 '%s login'을 사용하세요", binaryName())
	default:
		return fmt.Errorf("알 수 없는 설정 키입니다: %s (사용 가능: mode, host, port, auto_login)", key)
	}

	return nil
}

func parseOnOff(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "on":
		return true, nil
	case "off":
		return false, nil
	}

	return strconv.ParseBool(value)
}

func init() {
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
