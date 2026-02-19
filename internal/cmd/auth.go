package cmd

import (
	"context"
	"errors"
	"time"

	"github.com/99designs/keyring"

	"wxcli/internal/config"
	"wxcli/internal/outfmt"
	"wxcli/internal/secrets"
	"wxcli/internal/ui"
)

type AuthCmd struct {
	Set     AuthSetCmd     `cmd:"" name:"set" help:"Set AppID/AppSecret"`
	Status  AuthStatusCmd  `cmd:"" name:"status" help:"Show auth status"`
	Clear   AuthClearCmd   `cmd:"" name:"clear" help:"Clear stored secrets"`
	Keyring AuthKeyringCmd `cmd:"" name:"keyring" help:"Show/set keyring backend"`
}

func (c *AuthCmd) Run(_ context.Context, _ *RootFlags) error {
	return nil
}

type AuthSetCmd struct {
	AppID     string `name:"appid" help:"Weixin AppID" required:""`
	AppSecret string `name:"appsecret" help:"Weixin AppSecret" required:""`
	Name      string `name:"name" help:"Optional label"`
}

func (c *AuthSetCmd) Run(ctx context.Context, flags *RootFlags) error {
	if flags != nil && flags.NoInput {
		if c.AppID == "" || c.AppSecret == "" {
			return usage("--no-input requires --appid and --appsecret")
		}
	}
	if c.AppID == "" || c.AppSecret == "" {
		return usage("missing appid/appsecret")
	}
	cfg := config.File{AppID: c.AppID, Name: c.Name}
	if err := config.WriteConfig(cfg); err != nil {
		return err
	}
	store, err := secrets.OpenDefault()
	if err != nil {
		return err
	}
	if err := store.SetAppSecret(c.AppID, c.AppSecret); err != nil {
		return err
	}
	u := ui.FromContext(ctx)
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{"stored": true})
	}
	u.Out().Printf("stored\ttrue")
	return nil
}

type AuthStatusCmd struct{}

func (c *AuthStatusCmd) Run(ctx context.Context, _ *RootFlags) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}
	store, err := secrets.OpenDefault()
	if err != nil {
		return err
	}
	info, err := secrets.ResolveKeyringBackendInfo()
	if err != nil {
		return err
	}
	appID := cfg.AppID
	secretOK := false
	if appID != "" {
		if _, err := store.GetAppSecret(appID); err == nil {
			secretOK = true
		}
	}
	tokExp := ""
	if appID != "" {
		if tok, err := store.GetAccessToken(appID); err == nil {
			if !tok.ExpiresAt.IsZero() {
				tokExp = tok.ExpiresAt.UTC().Format(time.RFC3339)
			}
		} else if !errors.Is(err, keyring.ErrKeyNotFound) {
			return err
		}
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"appid":                   appID,
			"has_appsecret":           secretOK,
			"access_token_expires_at": tokExp,
			"keyring_backend":         info.Value,
			"keyring_backend_source":  info.Source,
		})
	}
	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("appid=%s", appID)
		u.Out().Printf("has_appsecret=%t", secretOK)
		u.Out().Printf("access_token_expires_at=%s", tokExp)
		u.Out().Printf("keyring_backend=%s", info.Value)
		u.Out().Printf("keyring_backend_source=%s", info.Source)
		return nil
	}
	u.Out().Printf("appid\t%s", appID)
	u.Out().Printf("has_appsecret\t%t", secretOK)
	u.Out().Printf("access_token_expires_at\t%s", tokExp)
	u.Out().Printf("keyring_backend\t%s", info.Value)
	u.Out().Printf("keyring_backend_source\t%s", info.Source)
	return nil
}

type AuthClearCmd struct{}

func (c *AuthClearCmd) Run(ctx context.Context, _ *RootFlags) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}
	store, err := secrets.OpenDefault()
	if err != nil {
		return err
	}
	if cfg.AppID != "" {
		if err := store.Clear(cfg.AppID); err != nil {
			return err
		}
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{"cleared": true})
	}
	u := ui.FromContext(ctx)
	u.Out().Printf("cleared\ttrue")
	return nil
}

type AuthKeyringCmd struct {
	Backend string `arg:"" optional:"" name:"backend" help:"Backend: auto|keychain|file"`
}

func (c *AuthKeyringCmd) Run(ctx context.Context, _ *RootFlags) error {
	backend := c.Backend
	if backend != "" {
		cfg, err := config.ReadConfig()
		if err != nil {
			return err
		}
		cfg.KeyringBackend = backend
		if err := config.WriteConfig(cfg); err != nil {
			return err
		}
	}
	info, err := secrets.ResolveKeyringBackendInfo()
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"keyring_backend":        info.Value,
			"keyring_backend_source": info.Source,
		})
	}
	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("keyring_backend=%s", info.Value)
		u.Out().Printf("keyring_backend_source=%s", info.Source)
		return nil
	}
	u.Out().Printf("keyring_backend\t%s", info.Value)
	u.Out().Printf("keyring_backend_source\t%s", info.Source)
	return nil
}
