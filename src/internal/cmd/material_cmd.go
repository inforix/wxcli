package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"wxcli/src/internal/auth"
	"wxcli/src/internal/material"
	"wxcli/src/internal/outfmt"
	"wxcli/src/internal/secrets"
	"wxcli/src/internal/ui"
)

type MaterialGetCmd struct {
	MediaID string `arg:"" name:"media_id" help:"Material media_id" required:""`
	Output  string `name:"output" help:"Write binary material to file (use '-' for stdout)"`
}

type MaterialUploadCmd struct {
	Type        string `name:"type" help:"Material type: image|voice|video|thumb" default:"image"`
	File        string `name:"file" help:"Path to file" required:""`
	Title       string `name:"title" help:"Video title (video only)"`
	Description string `name:"description" help:"Video description (video only)"`
}

type MaterialAddNewsCmd struct {
	Title              string `name:"title" help:"Article title" required:""`
	Content            string `name:"content" help:"Article content (HTML); use '-' to read from stdin" required:""`
	ThumbMediaID       string `name:"thumb-media-id" help:"Thumb media ID" required:""`
	Author             string `name:"author" help:"Article author"`
	Digest             string `name:"digest" help:"Article digest"`
	ContentSourceURL   string `name:"content-source-url" help:"Original article URL"`
	ShowCoverPic       int    `name:"show-cover-pic" help:"Show cover pic (0|1)" default:"0"`
	NeedOpenComment    int    `name:"need-open-comment" help:"Open comment (0|1)" default:"0"`
	OnlyFansCanComment int    `name:"only-fans-can-comment" help:"Only fans can comment (0|1)" default:"0"`
}

type MaterialUpdateNewsCmd struct {
	MediaID            string `arg:"" name:"media_id" help:"Material media_id" required:""`
	Index              int    `name:"index" help:"Article index (0-based)" default:"0"`
	Title              string `name:"title" help:"Article title" required:""`
	Content            string `name:"content" help:"Article content (HTML); use '-' to read from stdin" required:""`
	ThumbMediaID       string `name:"thumb-media-id" help:"Thumb media ID" required:""`
	Author             string `name:"author" help:"Article author"`
	Digest             string `name:"digest" help:"Article digest"`
	ContentSourceURL   string `name:"content-source-url" help:"Original article URL"`
	ShowCoverPic       int    `name:"show-cover-pic" help:"Show cover pic (0|1)" default:"0"`
	NeedOpenComment    int    `name:"need-open-comment" help:"Open comment (0|1)" default:"0"`
	OnlyFansCanComment int    `name:"only-fans-can-comment" help:"Only fans can comment (0|1)" default:"0"`
}

type MaterialListCmd struct {
	Type   string `name:"type" help:"Material type: news|image|video|voice" default:"news"`
	Offset int    `name:"offset" help:"Offset" default:"0"`
	Count  int    `name:"count" help:"Count (1-20)" default:"10"`
}

type MaterialDeleteCmd struct {
	MediaID string `arg:"" name:"media_id" help:"Material media_id" required:""`
}

type MaterialCountCmd struct{}

func (c *MaterialGetCmd) Run(ctx context.Context, _ *RootFlags) error {
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
	client := material.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	resp, err := client.Get(ctx, accessToken, material.GetMaterialRequest{MediaID: c.MediaID})
	if err != nil {
		return err
	}

	if len(resp.JSON) > 0 {
		if outfmt.IsJSON(ctx) {
			return writeJSONBytes(ctx, resp.JSON)
		}
		u := ui.FromContext(ctx)
		if resp.News != nil && len(resp.News.NewsItem) > 0 {
			for _, item := range resp.News.NewsItem {
				if outfmt.IsPlain(ctx) {
					u.Out().Printf("title=%s", item.Title)
					u.Out().Printf("author=%s", item.Author)
					u.Out().Printf("url=%s", item.URL)
					continue
				}
				u.Out().Printf("title\t%s", item.Title)
				u.Out().Printf("author\t%s", item.Author)
				u.Out().Printf("url\t%s", item.URL)
			}
			return nil
		}
		if resp.Video != nil {
			if outfmt.IsPlain(ctx) {
				u.Out().Printf("title=%s", resp.Video.Title)
				u.Out().Printf("description=%s", resp.Video.Description)
				u.Out().Printf("down_url=%s", resp.Video.DownURL)
				return nil
			}
			u.Out().Printf("title\t%s", resp.Video.Title)
			u.Out().Printf("description\t%s", resp.Video.Description)
			u.Out().Printf("down_url\t%s", resp.Video.DownURL)
			return nil
		}
		trimmed := strings.TrimSpace(string(resp.JSON))
		if outfmt.IsPlain(ctx) {
			u.Out().Printf("json=%s", trimmed)
			return nil
		}
		u.Out().Printf("json\t%s", trimmed)
		return nil
	}

	if len(resp.Data) == 0 {
		return errors.New("empty material response")
	}
	if c.Output == "" {
		return errors.New("binary material returned; use --output to save")
	}
	if c.Output == "-" {
		if outfmt.IsJSON(ctx) || outfmt.IsPlain(ctx) {
			return usage("binary output is not compatible with --json/--plain")
		}
		_, err := os.Stdout.Write(resp.Data)
		return err
	}
	outputPath, err := resolveMaterialOutput(c.Output, resp, c.MediaID)
	if err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, resp.Data, 0o644); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{
			"saved":        true,
			"path":         outputPath,
			"content_type": resp.ContentType,
			"size":         len(resp.Data),
			"filename":     resp.Filename,
		})
	}
	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("saved=true")
		u.Out().Printf("path=%s", outputPath)
		if resp.ContentType != "" {
			u.Out().Printf("content_type=%s", resp.ContentType)
		}
		u.Out().Printf("size=%d", len(resp.Data))
		if resp.Filename != "" {
			u.Out().Printf("filename=%s", resp.Filename)
		}
		return nil
	}
	u.Out().Printf("saved\ttrue")
	u.Out().Printf("path\t%s", outputPath)
	if resp.ContentType != "" {
		u.Out().Printf("content_type\t%s", resp.ContentType)
	}
	u.Out().Printf("size\t%d", len(resp.Data))
	if resp.Filename != "" {
		u.Out().Printf("filename\t%s", resp.Filename)
	}
	return nil
}

func (c *MaterialUploadCmd) Run(ctx context.Context, _ *RootFlags) error {
	if !isValidUploadMaterialType(c.Type) {
		return usage("type must be image|voice|video|thumb")
	}
	if c.Type == "video" {
		if strings.TrimSpace(c.Title) == "" {
			return usage("title is required for video uploads")
		}
		if strings.TrimSpace(c.Description) == "" {
			return usage("description is required for video uploads")
		}
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
	client := material.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	resp, err := client.Upload(ctx, accessToken, material.UploadMaterialRequest{
		Type:             c.Type,
		FilePath:         c.File,
		VideoTitle:       c.Title,
		VideoDescription: c.Description,
	})
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp)
	}
	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("media_id=%s", resp.MediaID)
		if resp.URL != "" {
			u.Out().Printf("url=%s", resp.URL)
		}
		return nil
	}
	u.Out().Printf("media_id\t%s", resp.MediaID)
	if resp.URL != "" {
		u.Out().Printf("url\t%s", resp.URL)
	}
	return nil
}

func (c *MaterialAddNewsCmd) Run(ctx context.Context, _ *RootFlags) error {
	content := c.Content
	if content == "-" {
		stdinContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		content = string(stdinContent)
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
	client := material.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	resp, err := client.AddNews(ctx, accessToken, material.AddNewsRequest{
		Articles: []material.NewsArticle{{
			Title:              c.Title,
			ThumbMediaID:       c.ThumbMediaID,
			Author:             c.Author,
			Digest:             c.Digest,
			ShowCoverPic:       c.ShowCoverPic,
			Content:            content,
			ContentSourceURL:   c.ContentSourceURL,
			NeedOpenComment:    c.NeedOpenComment,
			OnlyFansCanComment: c.OnlyFansCanComment,
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

func (c *MaterialUpdateNewsCmd) Run(ctx context.Context, _ *RootFlags) error {
	if c.Index < 0 {
		return usage("index must be >= 0")
	}
	content := c.Content
	if content == "-" {
		stdinContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		content = string(stdinContent)
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
	client := material.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	if err := client.UpdateNews(ctx, accessToken, material.UpdateNewsRequest{
		MediaID: c.MediaID,
		Index:   c.Index,
		Article: material.NewsArticle{
			Title:              c.Title,
			ThumbMediaID:       c.ThumbMediaID,
			Author:             c.Author,
			Digest:             c.Digest,
			ShowCoverPic:       c.ShowCoverPic,
			Content:            content,
			ContentSourceURL:   c.ContentSourceURL,
			NeedOpenComment:    c.NeedOpenComment,
			OnlyFansCanComment: c.OnlyFansCanComment,
		},
	}); err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{"updated": true})
	}
	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("updated=true")
		return nil
	}
	u.Out().Printf("updated\ttrue")
	return nil
}

func (c *MaterialListCmd) Run(ctx context.Context, _ *RootFlags) error {
	if c.Count < 1 || c.Count > 20 {
		return usage("count must be 1-20")
	}
	if !isValidMaterialType(c.Type) {
		return usage("type must be news|image|video|voice")
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
	client := material.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	resp, err := client.List(ctx, accessToken, material.BatchGetRequest{
		Type:   c.Type,
		Offset: c.Offset,
		Count:  c.Count,
	})
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
			if item.Name != "" {
				u.Out().Printf("name=%s", item.Name)
			}
			if item.URL != "" {
				u.Out().Printf("url=%s", item.URL)
			}
			if item.Content != nil && len(item.Content.NewsItem) > 0 {
				u.Out().Printf("news_count=%d", len(item.Content.NewsItem))
				for i, news := range item.Content.NewsItem {
					u.Out().Printf("news_title_%d=%s", i, news.Title)
				}
			}
		}
		return nil
	}
	u.Out().Printf("total_count\t%d", resp.TotalCount)
	u.Out().Printf("item_count\t%d", resp.ItemCount)
	for _, item := range resp.Item {
		u.Out().Printf("media_id\t%s", item.MediaID)
		if item.Name != "" {
			u.Out().Printf("name\t%s", item.Name)
		}
		if item.URL != "" {
			u.Out().Printf("url\t%s", item.URL)
		}
		if item.Content != nil && len(item.Content.NewsItem) > 0 {
			u.Out().Printf("news_count\t%d", len(item.Content.NewsItem))
			for i, news := range item.Content.NewsItem {
				u.Out().Printf("news_title[%d]\t%s", i, news.Title)
			}
		}
	}
	return nil
}

func (c *MaterialDeleteCmd) Run(ctx context.Context, _ *RootFlags) error {
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
	client := material.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	if err := client.Delete(ctx, accessToken, c.MediaID); err != nil {
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

func (c *MaterialCountCmd) Run(ctx context.Context, _ *RootFlags) error {
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
	client := material.NewClient(nil)
	if flags := RootFlagsFromContext(ctx); flags != nil && flags.BaseURL != "" {
		client.BaseURL = flags.BaseURL
	}
	resp, err := client.Count(ctx, accessToken)
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, resp)
	}
	u := ui.FromContext(ctx)
	if outfmt.IsPlain(ctx) {
		u.Out().Printf("voice_count=%d", resp.VoiceCount)
		u.Out().Printf("video_count=%d", resp.VideoCount)
		u.Out().Printf("image_count=%d", resp.ImageCount)
		u.Out().Printf("news_count=%d", resp.NewsCount)
		return nil
	}
	u.Out().Printf("voice_count\t%d", resp.VoiceCount)
	u.Out().Printf("video_count\t%d", resp.VideoCount)
	u.Out().Printf("image_count\t%d", resp.ImageCount)
	u.Out().Printf("news_count\t%d", resp.NewsCount)
	return nil
}

func resolveMaterialOutput(output string, resp material.GetMaterialResult, mediaID string) (string, error) {
	info, err := os.Stat(output)
	if err == nil && info.IsDir() {
		name := resp.Filename
		if name == "" {
			name = mediaID
		}
		return filepath.Join(output, name), nil
	}
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	return output, nil
}

func isValidMaterialType(value string) bool {
	switch value {
	case "news", "image", "video", "voice":
		return true
	default:
		return false
	}
}

func isValidUploadMaterialType(value string) bool {
	switch value {
	case "image", "voice", "video", "thumb":
		return true
	default:
		return false
	}
}
