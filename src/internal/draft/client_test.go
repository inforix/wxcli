package draft

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAddDraft(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
			t.Fatalf("unexpected content-type: %q", ct)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		got := string(body)
		if strings.Contains(got, "\\u4f60") || strings.Contains(got, "\\u597d") {
			t.Fatalf("expected utf-8 body, got escaped unicode: %s", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"media_id": "mid"})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.Add(context.Background(), "token", AddDraftRequest{
		Articles: []DraftArticle{{
			Title:   "t",
			Content: "<p>你好</p>",
		}},
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if resp.MediaID != "mid" {
		t.Fatalf("expected media_id mid, got %q", resp.MediaID)
	}
}

func TestListPaging(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"total_count": 2,
			"item_count":  1,
			"item": []map[string]any{
				{"media_id": "m1", "update_time": 1},
			},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.Client())
	client.BaseURL = srv.URL

	resp, err := client.List(context.Background(), "token", BatchGetRequest{Offset: 0, Count: 1})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.TotalCount != 2 || resp.ItemCount != 1 {
		t.Fatalf("unexpected counts: %+v", resp)
	}
	if len(resp.Item) != 1 || resp.Item[0].MediaID != "m1" {
		t.Fatalf("unexpected item: %+v", resp.Item)
	}
}
