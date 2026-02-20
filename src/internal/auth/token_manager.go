package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/99designs/keyring"

	"wxcli/src/internal/config"
	"wxcli/src/internal/errfmt"
	"wxcli/src/internal/httpclient"
	"wxcli/src/internal/secrets"
)

const (
	defaultTokenURL = "https://api.weixin.qq.com/cgi-bin/token"
	refreshSkew     = 5 * time.Minute
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type TokenManager struct {
	HTTP     HTTPDoer
	Store    *secrets.Store
	Now      func() time.Time
	TokenURL string
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

func NewTokenManager(httpClient HTTPDoer, store *secrets.Store) *TokenManager {
	if httpClient == nil {
		httpClient = &http.Client{Transport: httpclient.NewRetryTransport(nil)}
	}
	return &TokenManager{
		HTTP:     httpClient,
		Store:    store,
		Now:      time.Now,
		TokenURL: defaultTokenURL,
	}
}

func (m *TokenManager) GetValidAccessToken(ctx context.Context, appID string) (string, error) {
	if m.Store == nil {
		return "", errors.New("missing token store")
	}
	if m.Now == nil {
		m.Now = time.Now
	}
	if m.TokenURL == "" {
		m.TokenURL = defaultTokenURL
	}
	if appID == "" {
		return "", errors.New("missing appid")
	}
	if tok, err := m.Store.GetAccessToken(appID); err == nil {
		if tok.Token != "" && tok.ExpiresAt.After(m.Now().Add(refreshSkew)) {
			return tok.Token, nil
		}
	} else if !errors.Is(err, keyring.ErrKeyNotFound) {
		return "", err
	}
	return m.refresh(ctx, appID)
}

func (m *TokenManager) Refresh(ctx context.Context, appID string) (string, error) {
	if appID == "" {
		return "", errors.New("missing appid")
	}
	if token, err := m.refresh(ctx, appID); err != nil {
		var apiErr *errfmt.APIError
		if errors.As(err, &apiErr) {
			if apiErr.Code == 40014 || apiErr.Code == 42001 {
				return m.refresh(ctx, appID)
			}
		}
		return "", err
	} else {
		return token, nil
	}
}

func (m *TokenManager) refresh(ctx context.Context, appID string) (string, error) {
	secret, err := m.Store.GetAppSecret(appID)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", keyring.ErrKeyNotFound
		}
		return "", err
	}

	query := url.Values{}
	query.Set("grant_type", "client_credential")
	query.Set("appid", appID)
	query.Set("secret", secret)
	endpoint := m.TokenURL + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := m.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var payload tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.ErrCode != 0 {
		return "", &errfmt.APIError{Code: payload.ErrCode, Message: payload.ErrMsg}
	}
	if payload.AccessToken == "" || payload.ExpiresIn <= 0 {
		return "", errors.New("token response missing access_token")
	}
	if m.Now == nil {
		m.Now = time.Now
	}
	expiresAt := m.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	if err := m.Store.SetAccessToken(appID, secrets.AccessToken{Token: payload.AccessToken, ExpiresAt: expiresAt}); err != nil {
		return "", err
	}
	return payload.AccessToken, nil
}

func RequireAppID() (string, error) {
	cfg, err := config.ReadConfig()
	if err != nil {
		return "", err
	}
	if cfg.AppID == "" {
		return "", errors.New("appid not configured")
	}
	return cfg.AppID, nil
}
