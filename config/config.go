package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	AccessToken  string `yaml:"access_token,omitempty"`
	RefreshToken string `yaml:"refresh_token,omitempty"`
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

// Load reads the config file from ~/.config/note_cli/config.yaml
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

	// 기존 설정에 Host/Port가 없는 경우 기본값 채우기 (선택적)
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 8880
	}

	return &cfg, nil
}

// Save writes the config to ~/.config/note_cli/config.yaml
func Save(cfg *Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
