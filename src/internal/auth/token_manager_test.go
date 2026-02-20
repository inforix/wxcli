package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/99designs/keyring"

	"wxcli/src/internal/secrets"
)

type fakeDoer struct {
	fn func(*http.Request) (*http.Response, error)
}

func (d fakeDoer) Do(req *http.Request) (*http.Response, error) {
	return d.fn(req)
}

func newManagerForTest(t *testing.T) (*TokenManager, *secrets.Store) {
	t.Helper()
	store := secrets.NewStoreWithKeyring(keyring.NewArrayKeyring(nil))
	manager := NewTokenManager(nil, store)
	return manager, store
}

func TestRefreshByExpiry(t *testing.T) {
	manager, store := newManagerForTest(t)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	manager.Now = func() time.Time { return now }
	store.SetAppSecret("app", "secret")
	store.SetAccessToken("app", secrets.AccessToken{Token: "old", ExpiresAt: now.Add(1 * time.Minute)})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "new",
			"expires_in":   7200,
		})
	}))
	defer srv.Close()
	manager.HTTP = srv.Client()
	manager.TokenURL = srv.URL

	got, err := manager.GetValidAccessToken(context.Background(), "app")
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if got != "new" {
		t.Fatalf("expected new token, got %q", got)
	}
}

func TestRefreshOnFailure(t *testing.T) {
	manager, store := newManagerForTest(t)
	store.SetAppSecret("app", "secret")

	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		call++
		if call == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"errcode": 40014,
				"errmsg":  "invalid access_token",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "ok",
			"expires_in":   7200,
		})
	}))
	defer srv.Close()
	manager.HTTP = srv.Client()
	manager.TokenURL = srv.URL

	_, err := manager.Refresh(context.Background(), "app")
	if err != nil {
		t.Fatalf("expected refresh success, got %v", err)
	}
	if _, err := manager.Refresh(context.Background(), "app"); err != nil {
		t.Fatalf("expected refresh success, got %v", err)
	}
	if call < 2 {
		t.Fatalf("expected retry behavior, got %d calls", call)
	}
}
