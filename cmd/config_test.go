package cmd

import (
	"testing"

	"note_cli/config"
)

func TestApplyConfigValue(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
		check   func(t *testing.T, cfg *config.Config)
	}{
		{
			name:  "set host",
			key:   "host",
			value: "example.com",
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.Host != "example.com" {
					t.Fatalf("unexpected host: %q", cfg.Host)
				}
			},
		},
		{
			name:  "set port",
			key:   "port",
			value: "9000",
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.Port != 9000 {
					t.Fatalf("unexpected port: %d", cfg.Port)
				}
			},
		},
		{
			name:  "enable auto_login",
			key:   "auto_login",
			value: "true",
			check: func(t *testing.T, cfg *config.Config) {
				if !cfg.AutoLogin {
					t.Fatal("expected auto_login enabled")
				}
			},
		},
		{
			name:  "enable auto_login with on",
			key:   "auto_login",
			value: "on",
			check: func(t *testing.T, cfg *config.Config) {
				if !cfg.AutoLogin {
					t.Fatal("expected auto_login enabled")
				}
			},
		},
		{
			name:  "disable auto_login clears credentials",
			key:   "auto_login",
			value: "off",
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.AutoLogin {
					t.Fatal("expected auto_login disabled")
				}
				if cfg.Username != "" || cfg.Password != "" {
					t.Fatalf("expected credentials cleared, got %q / %q", cfg.Username, cfg.Password)
				}
			},
		},
		{name: "invalid auto_login value", key: "auto_login", value: "maybe", wantErr: true},
		{name: "username key rejected", key: "username", value: "a@b.c", wantErr: true},
		{name: "empty host", key: "host", value: "", wantErr: true},
		{name: "non-numeric port", key: "port", value: "abc", wantErr: true},
		{name: "port out of range", key: "port", value: "70000", wantErr: true},
		{name: "token key rejected", key: "access_token", value: "x", wantErr: true},
		{name: "unknown key", key: "theme", value: "dark", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Host:      "127.0.0.1",
				Port:      8880,
				AutoLogin: true,
				Username:  "stored@example.com",
				Password:  "stored-password",
			}
			err := applyConfigValue(cfg, tt.key, tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, cfg)
		})
	}
}
