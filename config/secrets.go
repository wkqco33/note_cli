package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type RuntimeSecrets struct {
	APIKey    string
	SecretKey string
}

func LoadRuntimeSecrets() (RuntimeSecrets, error) {
	secrets := readRuntimeSecrets()
	if secrets.APIKey != "" && secrets.SecretKey != "" {
		return secrets, nil
	}

	if err := loadDotEnv(); err != nil {
		return RuntimeSecrets{}, err
	}

	secrets = readRuntimeSecrets()
	if secrets.APIKey != "" && secrets.SecretKey != "" {
		return secrets, nil
	}

	return RuntimeSecrets{}, fmt.Errorf("API 키가 설정되지 않았습니다. NOTE_CLI_API_KEY / NOTE_CLI_SECRET_KEY 환경변수(.env 포함)를 설정하거나 로그인 인증을 사용하세요")
}

func loadDotEnv() error {
	envFile := os.Getenv("NOTE_CLI_ENV_FILE")
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf(".env 파일을 불러오지 못했습니다: %w", err)
		}
		return nil
	}

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(".env 파일을 불러오지 못했습니다: %w", err)
	}

	return nil
}

func readRuntimeSecrets() RuntimeSecrets {
	return RuntimeSecrets{
		APIKey:    firstNonEmpty(os.Getenv("NOTE_CLI_API_KEY"), os.Getenv("API_KEY")),
		SecretKey: firstNonEmpty(os.Getenv("NOTE_CLI_SECRET_KEY"), os.Getenv("SECRET_KEY")),
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}
