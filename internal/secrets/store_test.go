package secrets

import (
	"testing"
	"time"

	"github.com/99designs/keyring"
)

func TestStoreAppSecret(t *testing.T) {
	store := &Store{ring: keyring.NewArrayKeyring(nil)}
	if err := store.SetAppSecret("app", "secret"); err != nil {
		t.Fatalf("set appsecret: %v", err)
	}
	got, err := store.GetAppSecret("app")
	if err != nil {
		t.Fatalf("get appsecret: %v", err)
	}
	if got != "secret" {
		t.Fatalf("expected secret, got %q", got)
	}
}

func TestStoreAccessToken(t *testing.T) {
	store := &Store{ring: keyring.NewArrayKeyring(nil)}
	input := AccessToken{Token: "tok", ExpiresAt: time.Now().Add(time.Hour).UTC()}
	if err := store.SetAccessToken("app", input); err != nil {
		t.Fatalf("set token: %v", err)
	}
	got, err := store.GetAccessToken("app")
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if got.Token != input.Token {
		t.Fatalf("expected token %q, got %q", input.Token, got.Token)
	}
	if got.ExpiresAt.IsZero() {
		t.Fatalf("expected expires_at set")
	}
}

func TestClear(t *testing.T) {
	store := &Store{ring: keyring.NewArrayKeyring(nil)}
	if err := store.SetAppSecret("app", "secret"); err != nil {
		t.Fatalf("set appsecret: %v", err)
	}
	if err := store.SetAccessToken("app", AccessToken{Token: "tok"}); err != nil {
		t.Fatalf("set token: %v", err)
	}
	if err := store.Clear("app"); err != nil {
		t.Fatalf("clear: %v", err)
	}
}
