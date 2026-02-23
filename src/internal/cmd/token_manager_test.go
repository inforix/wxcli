package cmd

import (
	"context"
	"testing"

	"github.com/99designs/keyring"

	"wxcli/src/internal/secrets"
)

func TestNewTokenManagerRespectsBaseURL(t *testing.T) {
	ctx := context.WithValue(context.Background(), rootFlagsCtxKey{}, &RootFlags{
		BaseURL: "http://localhost:8080/api",
	})
	store := secrets.NewStoreWithKeyring(keyring.NewArrayKeyring(nil))
	manager, err := newTokenManager(ctx, store)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}
	want := "http://localhost:8080/api/cgi-bin/token"
	if manager.TokenURL != want {
		t.Fatalf("expected token URL %q, got %q", want, manager.TokenURL)
	}
}
