package cmd

import (
	"path/filepath"
	"testing"
	"time"

	"wxcli/src/internal/config"
)

func TestAuthSetPreservesConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, ".config"))
	t.Setenv("APPDATA", root)
	t.Setenv("LOCALAPPDATA", root)

	expiresAt := time.Now().Add(5 * time.Minute).UTC()
	cfg := config.File{
		AppID:          "old",
		Name:           "existing",
		KeyringBackend: "file",
		AccessToken:    &config.AccessToken{Token: "tok", ExpiresAt: expiresAt},
	}
	if err := config.WriteConfig(cfg); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, _, err := runCLI(t, []string{"auth", "set", "--appid", "new", "--appsecret", "secret"}); err != nil {
		t.Fatalf("auth set: %v", err)
	}

	updated, err := config.ReadConfig()
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	if updated.AppID != "new" {
		t.Fatalf("expected appid new, got %q", updated.AppID)
	}
	if updated.Name != "existing" {
		t.Fatalf("expected name preserved, got %q", updated.Name)
	}
	if updated.KeyringBackend != "file" {
		t.Fatalf("expected keyring_backend preserved, got %q", updated.KeyringBackend)
	}
	if updated.AccessToken == nil || updated.AccessToken.Token != "tok" {
		t.Fatalf("expected access_token preserved, got %+v", updated.AccessToken)
	}
}

func TestAuthKeyringRejectsInvalidBackend(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, ".config"))
	t.Setenv("APPDATA", root)
	t.Setenv("LOCALAPPDATA", root)

	_, _, err := runCLI(t, []string{"auth", "keyring", "nope"})
	if err == nil {
		t.Fatalf("expected error for invalid backend")
	}
	if ExitCode(err) != 2 {
		t.Fatalf("expected exit code 2, got %d", ExitCode(err))
	}
}
