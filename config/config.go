package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	AccessToken  string `yaml:"access_token,omitempty"`
	RefreshToken string `yaml:"refresh_token,omitempty"`
	AutoLogin    bool   `yaml:"auto_login,omitempty"`
	Username     string `yaml:"username,omitempty"`
	Password     string `yaml:"password,omitempty"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// ~/.config/note_cli 디렉토리 사용 (없으면 생성)
	configDir := filepath.Join(home, ".config", "note_cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.yaml"), nil
}

// Path 설정 파일 경로 반환 (~/.config/note_cli/config.yaml)
func Path() (string, error) {
	return getConfigPath()
}

// Load 설정 파일을 ~/.config/note_cli/config.yaml에서 읽는다.
func Load() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 파일이 없으면 기본 설정값으로 생성
			defaultCfg := &Config{
				Host: "127.0.0.1",
				Port: 8880,
			}
			if saveErr := Save(defaultCfg); saveErr != nil {
				return nil, saveErr
			}
			return defaultCfg, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	decodeSecretFields(&cfg)

	// 기존 설정에 Host/Port가 없는 경우 기본값 채우기 (선택적)
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 8880
	}

	return &cfg, nil
}

// Save 설정을 ~/.config/note_cli/config.yaml에 저장한다.
// 토큰과 비밀번호는 평문 대신 암호화된 형태로 기록한다.
func Save(cfg *Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	toSave := *cfg
	if err := encodeSecretFields(&toSave); err != nil {
		return err
	}

	data, err := yaml.Marshal(&toSave)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// secretFields 암호화 대상 필드 목록 (액세스/리프레시 토큰, 자동 로그인 비밀번호)
func secretFields(cfg *Config) map[string]*string {
	return map[string]*string{
		"access_token":  &cfg.AccessToken,
		"refresh_token": &cfg.RefreshToken,
		"password":      &cfg.Password,
	}
}

func encodeSecretFields(cfg *Config) error {
	for name, value := range secretFields(cfg) {
		encrypted, err := encodeSecret(*value)
		if err != nil {
			return fmt.Errorf("%s 값을 암호화하지 못했습니다: %w", name, err)
		}
		*value = encrypted
	}

	return nil
}

// decodeSecretFields 저장된 비밀값 복호화. 다른 사용자/컴퓨터에서 복사된 설정 등으로
// 복호화가 불가능한 필드는 초기화하고 계속 진행한다 (재로그인으로 복구 가능).
func decodeSecretFields(cfg *Config) {
	cleared := false
	for _, value := range secretFields(cfg) {
		plain, err := decodeSecret(*value)
		if err != nil {
			*value = ""
			cleared = true
			continue
		}
		*value = plain
	}

	if cleared {
		fmt.Fprintln(os.Stderr, "경고: 저장된 인증 정보를 복호화하지 못해 초기화했습니다. 'login' 명령으로 다시 로그인하세요.")
	}
}
