package config

import (
	"os"
	"path/filepath"
	"testing"

	"note_cli/config/buildinfo"
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

func TestLoadRuntimeSecretsFallsBackToEmbeddedValues(t *testing.T) {
	oldAPIKey := buildinfo.APIKey
	oldSecretKey := buildinfo.SecretKey
	buildinfo.APIKey = "embedded-api"
	buildinfo.SecretKey = "embedded-secret"
	t.Cleanup(func() {
		buildinfo.APIKey = oldAPIKey
		buildinfo.SecretKey = oldSecretKey
	})

	_ = os.Unsetenv("NOTE_CLI_API_KEY")
	_ = os.Unsetenv("NOTE_CLI_SECRET_KEY")
	_ = os.Unsetenv("API_KEY")
	_ = os.Unsetenv("SECRET_KEY")
	t.Setenv("NOTE_CLI_ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))

	secrets, err := LoadRuntimeSecrets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secrets.APIKey != "embedded-api" || secrets.SecretKey != "embedded-secret" {
		t.Fatalf("unexpected secrets: %#v", secrets)
	}
}

func TestLoadRuntimeSecretsPrefersRuntimeEnvOverEmbedded(t *testing.T) {
	oldAPIKey := buildinfo.APIKey
	oldSecretKey := buildinfo.SecretKey
	buildinfo.APIKey = "embedded-api"
	buildinfo.SecretKey = "embedded-secret"
	t.Cleanup(func() {
		buildinfo.APIKey = oldAPIKey
		buildinfo.SecretKey = oldSecretKey
	})

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

	_ = os.Unsetenv("NOTE_CLI_API_KEY")
	_ = os.Unsetenv("NOTE_CLI_SECRET_KEY")
	_ = os.Unsetenv("API_KEY")
	_ = os.Unsetenv("SECRET_KEY")
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
	_ = os.Unsetenv("NOTE_CLI_API_KEY")
	_ = os.Unsetenv("NOTE_CLI_SECRET_KEY")
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
	_ = os.Unsetenv("NOTE_CLI_API_KEY")
	_ = os.Unsetenv("NOTE_CLI_SECRET_KEY")
	_ = os.Unsetenv("API_KEY")
	_ = os.Unsetenv("SECRET_KEY")
	t.Setenv("NOTE_CLI_ENV_FILE", filepath.Join(t.TempDir(), "missing.env"))

	if _, err := LoadRuntimeSecrets(); err == nil {
		t.Fatal("expected missing secret error")
	}
}
