package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/99designs/keyring"

	"wxcli/internal/config"
)

const (
	keyringPasswordEnv = "WXCLI_KEYRING_PASSWORD"
	keyringBackendEnv  = "WXCLI_KEYRING_BACKEND"

	backendAuto     = "auto"
	backendKeychain = "keychain"
	backendFile     = "file"
)

type BackendInfo struct {
	Value  string
	Source string
}

func KeyringBackendEnv() string {
	return keyringBackendEnv
}

func KeyringPasswordEnv() string {
	return keyringPasswordEnv
}

func ResolveKeyringBackendInfo() (BackendInfo, error) {
	if v := normalizeBackend(os.Getenv(keyringBackendEnv)); v != "" {
		return BackendInfo{Value: v, Source: "env"}, nil
	}
	cfg, err := config.ReadConfig()
	if err != nil {
		return BackendInfo{}, fmt.Errorf("read config: %w", err)
	}
	if v := normalizeBackend(cfg.KeyringBackend); v != "" {
		return BackendInfo{Value: v, Source: "config"}, nil
	}
	return BackendInfo{Value: backendAuto, Source: "default"}, nil
}

func OpenDefault() (*Store, error) {
	info, err := ResolveKeyringBackendInfo()
	if err != nil {
		return nil, err
	}
	backends, err := allowedBackends(info)
	if err != nil {
		return nil, err
	}
	keyringDir, err := config.EnsureKeyringDir()
	if err != nil {
		return nil, fmt.Errorf("ensure keyring dir: %w", err)
	}
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends:          backends,
		FileDir:                  keyringDir,
		FilePasswordFunc:         filePasswordFunc(),
		ServiceName:              "wxcli",
		KeychainName:             "wxcli",
		KeychainTrustApplication: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}
	return &Store{ring: kr}, nil
}

func allowedBackends(info BackendInfo) ([]keyring.BackendType, error) {
	switch info.Value {
	case "", backendAuto:
		return []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.FileBackend,
		}, nil
	case backendKeychain:
		return []keyring.BackendType{keyring.KeychainBackend}, nil
	case backendFile:
		return []keyring.BackendType{keyring.FileBackend}, nil
	default:
		return nil, fmt.Errorf("invalid keyring backend: %q", info.Value)
	}
}

func normalizeBackend(v string) string {
	if v == "" {
		return ""
	}
	value := strings.ToLower(strings.TrimSpace(v))
	if value == backendAuto || value == backendKeychain || value == backendFile {
		return value
	}
	return ""
}

func filePasswordFunc() keyring.PromptFunc {
	if password, ok := os.LookupEnv(keyringPasswordEnv); ok {
		return keyring.FixedStringPrompt(password)
	}
	return keyring.TerminalPrompt
}

type Store struct {
	ring keyring.Keyring
}

func NewStoreWithKeyring(r keyring.Keyring) *Store {
	return &Store{ring: r}
}

type AccessToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func appSecretKey(appID string) string {
	return fmt.Sprintf("wxcli.appsecret.%s", appID)
}

func accessTokenKey(appID string) string {
	return fmt.Sprintf("wxcli.accesstoken.%s", appID)
}

func (s *Store) SetAppSecret(appID, secret string) error {
	return s.ring.Set(keyring.Item{Key: appSecretKey(appID), Data: []byte(secret)})
}

func (s *Store) GetAppSecret(appID string) (string, error) {
	item, err := s.ring.Get(appSecretKey(appID))
	if err != nil {
		return "", err
	}
	return string(item.Data), nil
}

func (s *Store) SetAccessToken(appID string, token AccessToken) error {
	b, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return s.ring.Set(keyring.Item{Key: accessTokenKey(appID), Data: b})
}

func (s *Store) GetAccessToken(appID string) (AccessToken, error) {
	item, err := s.ring.Get(accessTokenKey(appID))
	if err != nil {
		return AccessToken{}, err
	}
	var tok AccessToken
	if err := json.Unmarshal(item.Data, &tok); err != nil {
		return AccessToken{}, err
	}
	return tok, nil
}

func (s *Store) Clear(appID string) error {
	if err := s.ring.Remove(appSecretKey(appID)); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return err
	}
	if err := s.ring.Remove(accessTokenKey(appID)); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
		return err
	}
	return nil
}
