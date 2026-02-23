package cmd

import (
	"context"

	"wxcli/src/internal/auth"
	"wxcli/src/internal/secrets"
)

func newTokenManager(ctx context.Context, store *secrets.Store) (*auth.TokenManager, error) {
	manager := auth.NewTokenManager(nil, store)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		tokenURL, err := auth.TokenURLForBase(flags.BaseURL)
		if err != nil {
			return nil, err
		}
		manager.TokenURL = tokenURL
	}
	return manager, nil
}
