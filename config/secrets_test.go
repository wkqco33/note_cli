package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRuntimeSecretsFromEnv(t *testing.T) {
	t.Setenv("NOTE_CLI_API_KEY", "env-api")
	t.Setenv("NOTE_CLI_SECRET_KEY", "env-secret")

	secrets, err := LoadRuntimeSecrets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secrets.APIKey != "env-api" || secrets.SecretKey != "env-secret" {
		t.Fatalf("unexpected secrets: %#v", secrets)
	}
}

func TestLoadRuntimeSecretsFromDotEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("NOTE_CLI_API_KEY=file-api\nNOTE_CLI_SECRET_KEY=file-secret\n"), 0600); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	os.Unsetenv("NOTE_CLI_API_KEY")
	os.Unsetenv("NOTE_CLI_SECRET_KEY")
	os.Unsetenv("API_KEY")
	os.Unsetenv("SECRET_KEY")
	t.Setenv("NOTE_CLI_ENV_FILE", envPath)

	secrets, err := LoadRuntimeSecrets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secrets.APIKey != "file-api" || secrets.SecretKey != "file-secret" {
		t.Fatalf("unexpected secrets: %#v", secrets)
	}
}

func TestLoadRuntimeSecretsAcceptsLegacyNames(t *testing.T) {
	os.Unsetenv("NOTE_CLI_API_KEY")
	os.Unsetenv("NOTE_CLI_SECRET_KEY")
	t.Setenv("API_KEY", "legacy-api")
	t.Setenv("SECRET_KEY", "legacy-secret")

	secrets, err := LoadRuntimeSecrets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secrets.APIKey != "legacy-api" || secrets.SecretKey != "legacy-secret" {
		t.Fatalf("unexpected secrets: %#v", secrets)
	}
}

func TestLoadRuntimeSecretsFailsWhenMissing(t *testing.T) {
	os.Unsetenv("NOTE_CLI_API_KEY")
	os.Unsetenv("NOTE_CLI_SECRET_KEY")
	os.Unsetenv("API_KEY")
	os.Unsetenv("SECRET_KEY")
	t.Setenv("NOTE_CLI_ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))

	if _, err := LoadRuntimeSecrets(); err == nil {
		t.Fatal("expected missing secret error")
	}
}
