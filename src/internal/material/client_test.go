package material

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetMaterialNews(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"news_item": []map[string]any{{"title": "t", "author": "a"}},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.Get(context.Background(), "token", GetMaterialRequest{MediaID: "m"})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.News == nil || len(resp.News.NewsItem) != 1 {
		t.Fatalf("expected news result, got %+v", resp)
	}
}

func TestGetMaterialVideo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"title":       "t",
			"description": "d",
			"down_url":    "http://example.com/video.mp4",
		})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.Get(context.Background(), "token", GetMaterialRequest{MediaID: "m"})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.Video == nil || resp.Video.DownURL == "" {
		t.Fatalf("expected video result, got %+v", resp)
	}
}

func TestGetMaterialBinary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Disposition", "attachment; filename=photo.jpg")
		_, _ = w.Write([]byte{0x01, 0x02, 0x03})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.Get(context.Background(), "token", GetMaterialRequest{MediaID: "m"})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(resp.Data) != 3 || resp.ContentType != "image/jpeg" || resp.Filename != "photo.jpg" {
		t.Fatalf("unexpected binary result: %+v", resp)
	}
}

func TestGetMaterialError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 40014, "errmsg": "invalid"})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	if _, err := client.Get(context.Background(), "token", GetMaterialRequest{MediaID: "m"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCountMaterial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"voice_count": 1,
			"video_count": 2,
			"image_count": 3,
			"news_count":  4,
		})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.Count(context.Background(), "token")
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if resp.VoiceCount != 1 || resp.VideoCount != 2 || resp.ImageCount != 3 || resp.NewsCount != 4 {
		t.Fatalf("unexpected counts: %+v", resp)
	}
}

func TestListMaterial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"total_count": 2,
			"item_count":  1,
			"item": []map[string]any{
				{"media_id": "m1", "name": "n1", "update_time": 1, "url": "http://example.com"},
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.List(context.Background(), "token", BatchGetRequest{Type: "image", Offset: 0, Count: 1})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.TotalCount != 2 || resp.ItemCount != 1 || len(resp.Item) != 1 || resp.Item[0].MediaID != "m1" {
		t.Fatalf("unexpected list response: %+v", resp)
	}
}

func TestDeleteMaterial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 0, "errmsg": "ok"})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	if err := client.Delete(context.Background(), "token", "m1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestUploadMaterialImage(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "img.jpg")
	if err := os.WriteFile(filePath, []byte("img"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("access_token") != "token" {
			t.Fatalf("missing access_token")
		}
		if r.URL.Query().Get("type") != "image" {
			t.Fatalf("unexpected type: %q", r.URL.Query().Get("type"))
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Fatalf("expected multipart content-type")
		}
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		file, header, err := r.FormFile("media")
		if err != nil {
			t.Fatalf("form file: %v", err)
		}
		defer file.Close()
		if header.Filename != "img.jpg" {
			t.Fatalf("unexpected filename: %q", header.Filename)
		}
		data, _ := io.ReadAll(file)
		if string(data) != "img" {
			t.Fatalf("unexpected file data")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"media_id": "m1", "url": "http://example.com/img"})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.Upload(context.Background(), "token", UploadMaterialRequest{Type: "image", FilePath: filePath})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if resp.MediaID != "m1" || resp.URL == "" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestUploadMaterialVideoDescription(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "video.mp4")
	if err := os.WriteFile(filePath, []byte("video"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		desc := r.FormValue("description")
		var payload map[string]string
		if err := json.Unmarshal([]byte(desc), &payload); err != nil {
			t.Fatalf("decode description: %v", err)
		}
		if payload["title"] != "t" || payload["introduction"] != "d" {
			t.Fatalf("unexpected description: %v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"media_id": "mv"})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	_, err := client.Upload(context.Background(), "token", UploadMaterialRequest{
		Type:             "video",
		FilePath:         filePath,
		VideoTitle:       "t",
		VideoDescription: "d",
	})
	if err != nil {
		t.Fatalf("upload video: %v", err)
	}
}

func TestAddNews(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		articles, _ := body["articles"].([]any)
		if len(articles) != 1 {
			t.Fatalf("expected one article")
		}
		first := articles[0].(map[string]any)
		if first["title"] != "t" {
			t.Fatalf("unexpected title")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"media_id": "mid"})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.AddNews(context.Background(), "token", AddNewsRequest{Articles: []NewsArticle{{Title: "t", Content: "c", ThumbMediaID: "thumb"}}})
	if err != nil {
		t.Fatalf("add news: %v", err)
	}
	if resp.MediaID != "mid" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestUploadImage(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "img.png")
	if err := os.WriteFile(filePath, []byte("img"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/media/uploadimg" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, header, err := r.FormFile("media")
		if err != nil {
			t.Fatalf("form file: %v", err)
		}
		if header.Filename != "img.png" {
			t.Fatalf("unexpected filename: %q", header.Filename)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"url": "http://example.com/img.png"})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.UploadImage(context.Background(), "token", filePath)
	if err != nil {
		t.Fatalf("upload img: %v", err)
	}
	if resp.URL == "" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}
