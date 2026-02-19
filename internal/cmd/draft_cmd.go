package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/99designs/keyring"

	"wxcli/internal/auth"
	"wxcli/internal/draft"
	"wxcli/internal/outfmt"
	"wxcli/internal/secrets"
	"wxcli/internal/ui"
)

type DraftAddCmd struct {
	Title        string `name:"title" help:"Article title" required:""`
	Content      string `name:"content" help:"Article content (HTML)" required:""`
	ThumbMediaID string `name:"thumb-media-id" help:"Thumb media ID" required:""`
}

func (c *DraftAddCmd) Run(ctx context.Context, _ *RootFlags) error {
	appID, err := auth.RequireAppID()
	if err != nil {
		return err
	}
	store, err := secrets.OpenDefault()
	if err != nil {
		return err
	}
	manager := auth.NewTokenManager(nil, store)
	accessToken, err := manager.GetValidAccessToken(ctx, appID)
	if err != nil {
		return err
	}
	client := draft.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	resp, err := client.Add(ctx, accessToken, draft.AddDraftRequest{
		Articles: []draft.DraftArticle{{
			Title:        c.Title,
			Content:      c.Content,
			ThumbMediaID: c.ThumbMediaID,
		}},
	})
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{"media_id": resp.MediaID})
	}
	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("media_id=%s", resp.MediaID)
		return nil
	}
	u.Out().Printf("media_id\t%s", resp.MediaID)
	return nil
}

type DraftGetCmd struct {
	MediaID string `arg:"" name:"media_id" help:"Draft media_id"`
}

func (c *DraftGetCmd) Run(ctx context.Context, _ *RootFlags) error {
	appID, err := auth.RequireAppID()
	if err != nil {
		return err
	}
	store, err := secrets.OpenDefault()
	if err != nil {
		return err
	}
	manager := auth.NewTokenManager(nil, store)
	accessToken, err := manager.GetValidAccessToken(ctx, appID)
	if err != nil {
		return err
	}
	client := draft.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	resp, err := client.Get(ctx, accessToken, draft.GetDraftRequest{MediaID: c.MediaID})
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp)
	}
	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		for _, item := range resp.NewsItem {
			u.Out().Printf("title=%s", item.Title)
			u.Out().Printf("author=%s", item.Author)
		}
		return nil
	}
	for _, item := range resp.NewsItem {
		u.Out().Printf("title\t%s", item.Title)
		u.Out().Printf("author\t%s", item.Author)
	}
	return nil
}

type DraftListCmd struct {
	Offset    int `name:"offset" help:"Offset" default:"0"`
	Count     int `name:"count" help:"Count (1-20)" default:"10"`
	NoContent int `name:"no-content" help:"Skip content" default:"1"`
}

func (c *DraftListCmd) Run(ctx context.Context, _ *RootFlags) error {
	if c.Count < 1 || c.Count > 20 {
		return usage("count must be 1-20")
	}
	appID, err := auth.RequireAppID()
	if err != nil {
		return err
	}
	store, err := secrets.OpenDefault()
	if err != nil {
		return err
	}
	manager := auth.NewTokenManager(nil, store)
	accessToken, err := manager.GetValidAccessToken(ctx, appID)
	if err != nil {
		return err
	}
	client := draft.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	resp, err := client.List(ctx, accessToken, draft.BatchGetRequest{Offset: c.Offset, Count: c.Count, NoContent: c.NoContent})
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp)
	}
	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("total_count=%d", resp.TotalCount)
		u.Out().Printf("item_count=%d", resp.ItemCount)
		for _, item := range resp.Item {
			u.Out().Printf("media_id=%s", item.MediaID)
		}
		return nil
	}
	u.Out().Printf("total_count\t%d", resp.TotalCount)
	u.Out().Printf("item_count\t%d", resp.ItemCount)
	for _, item := range resp.Item {
		u.Out().Printf("media_id\t%s", item.MediaID)
	}
	return nil
}

type DraftDeleteCmd struct {
	MediaID string `arg:"" name:"media_id" help:"Draft media_id"`
}

func (c *DraftDeleteCmd) Run(ctx context.Context, _ *RootFlags) error {
	appID, err := auth.RequireAppID()
	if err != nil {
		return err
	}
	store, err := secrets.OpenDefault()
	if err != nil {
		return err
	}
	manager := auth.NewTokenManager(nil, store)
	accessToken, err := manager.GetValidAccessToken(ctx, appID)
	if err != nil {
		return err
	}
	client := draft.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	if err := client.Delete(ctx, accessToken, draft.DeleteDraftRequest{MediaID: c.MediaID}); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{"deleted": true})
	}
	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("deleted=true")
		return nil
	}
	u.Out().Printf("deleted\ttrue")
	return nil
}

func requireAppSecret(store *secrets.Store, appID string) error {
	if appID == "" {
		return errors.New("missing appid")
	}
	if _, err := store.GetAppSecret(appID); err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return fmt.Errorf("appsecret not configured for %s", appID)
		}
		return err
	}
	return nil
}
