package config

import (
	"os"
	"strings"
	"testing"
)

func setTestHome(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

func TestSaveEncryptsPassword(t *testing.T) {
	setTestHome(t)

	cfg := &Config{
		Host:      "127.0.0.1",
		Port:      8880,
		AutoLogin: true,
		Username:  "user@example.com",
		Password:  "plain-password",
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// 메모리의 Config는 평문을 유지해야 함
	if cfg.Password != "plain-password" {
		t.Fatalf("in-memory password mutated: %q", cfg.Password)
	}

	path, err := Path()
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if strings.Contains(string(raw), "plain-password") {
		t.Fatalf("config file contains plaintext password:\n%s", raw)
	}
	if !strings.Contains(string(raw), "password: enc:") {
		t.Fatalf("config file missing encrypted password:\n%s", raw)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if loaded.Password != "plain-password" {
		t.Fatalf("round-trip password mismatch: %q", loaded.Password)
	}
}

func TestSaveEncryptsTokens(t *testing.T) {
	setTestHome(t)

	cfg := &Config{
		Host:         "127.0.0.1",
		Port:         8880,
		AccessToken:  "plain-access-token",
		RefreshToken: "plain-refresh-token",
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	if cfg.AccessToken != "plain-access-token" || cfg.RefreshToken != "plain-refresh-token" {
		t.Fatalf("in-memory tokens mutated: %q / %q", cfg.AccessToken, cfg.RefreshToken)
	}

	path, err := Path()
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if strings.Contains(string(raw), "plain-access-token") || strings.Contains(string(raw), "plain-refresh-token") {
		t.Fatalf("config file contains plaintext token:\n%s", raw)
	}
	if !strings.Contains(string(raw), "access_token: enc:") || !strings.Contains(string(raw), "refresh_token: enc:") {
		t.Fatalf("config file missing encrypted tokens:\n%s", raw)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if loaded.AccessToken != "plain-access-token" || loaded.RefreshToken != "plain-refresh-token" {
		t.Fatalf("round-trip token mismatch: %q / %q", loaded.AccessToken, loaded.RefreshToken)
	}
}

func TestLoadKeepsLegacyPlaintextPassword(t *testing.T) {
	setTestHome(t)

	// 암호화 도입 이전에 저장된 평문 비밀번호는 그대로 읽히고
	// 다음 Save 때 암호화된다
	if err := Save(&Config{Host: "127.0.0.1", Port: 8880}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	path, err := Path()
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}
	legacy := "host: 127.0.0.1\nport: 8880\nauto_login: true\nusername: user@example.com\npassword: legacy-plain\n"
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatalf("failed to write legacy config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Password != "legacy-plain" {
		t.Fatalf("unexpected password: %q", cfg.Password)
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("failed to re-save config: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if strings.Contains(string(raw), "legacy-plain") {
		t.Fatalf("legacy password not encrypted on save:\n%s", raw)
	}
}

func TestLoadClearsUndecryptablePassword(t *testing.T) {
	setTestHome(t)

	if err := Save(&Config{Host: "127.0.0.1", Port: 8880}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	path, err := Path()
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}
	corrupt := "host: 127.0.0.1\nport: 8880\npassword: enc:aW52YWxpZC1jaXBoZXJ0ZXh0\n"
	if err := os.WriteFile(path, []byte(corrupt), 0600); err != nil {
		t.Fatalf("failed to write corrupt config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Password != "" {
		t.Fatalf("expected cleared password, got %q", cfg.Password)
	}
}

func TestLoadDefaultsNewConfigToLocalMode(t *testing.T) {
	setTestHome(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Mode != "local" {
		t.Fatalf("default mode = %q, want local", cfg.Mode)
	}
	if cfg.LLM.Provider != "ollama" || cfg.LLM.Model == "" || cfg.LLM.BaseURL == "" {
		t.Fatalf("unexpected LLM defaults: %+v", cfg.LLM)
	}
}

func TestSaveEncryptsLLMAPIKey(t *testing.T) {
	setTestHome(t)
	cfg := &Config{Mode: "local", Host: "127.0.0.1", Port: 8880, LLM: LLMConfig{APIKey: "plain-api-key"}}
	if err := Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
	path, err := Path()
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	if strings.Contains(string(raw), "plain-api-key") || !strings.Contains(string(raw), "api_key: enc:") {
		t.Fatalf("LLM API key was not encrypted: %s", raw)
	}
	loaded, err := Load()
	if err != nil || loaded.LLM.APIKey != "plain-api-key" {
		t.Fatalf("LLM API key round trip = %q, error = %v", loaded.LLM.APIKey, err)
	}
}

func TestLoadLegacyConfigDefaultsToRemoteMode(t *testing.T) {
	setTestHome(t)
	path, err := Path()
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}
	legacy := "host: 127.0.0.1\nport: 8880\n"
	if err := os.WriteFile(path, []byte(legacy), 0600); err != nil {
		t.Fatalf("failed to write legacy config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Mode != "remote" {
		t.Fatalf("legacy mode = %q, want remote", cfg.Mode)
	}
}
