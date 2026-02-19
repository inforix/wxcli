package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"wxcli/internal/config"
	"wxcli/internal/secrets"
)

func TestE2EDraftListAndDelete(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("WXCLI_KEYRING_BACKEND", "file")
	t.Setenv("WXCLI_KEYRING_PASSWORD", "test-pass")

	if err := os.MkdirAll(filepath.Join(root, "Library", "Application Support"), 0o700); err != nil {
		t.Fatalf("prepare config dir: %v", err)
	}
	if err := config.WriteConfig(config.File{AppID: "app"}); err != nil {
		t.Fatalf("write config: %v", err)
	}
	store, err := secrets.OpenDefault()
	if err != nil {
		t.Fatalf("open secrets: %v", err)
	}
	if err := store.SetAccessToken("app", secrets.AccessToken{Token: "tok", ExpiresAt: time.Now().Add(30 * time.Minute)}); err != nil {
		t.Fatalf("set token: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("access_token") != "tok" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 40014, "errmsg": "invalid access_token"})
			return
		}
		switch r.URL.Path {
		case "/cgi-bin/draft/batchget":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"total_count": 1,
				"item_count":  1,
				"item":        []map[string]any{{"media_id": "m1", "update_time": 1}},
			})
		case "/cgi-bin/draft/delete":
			_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 0, "errmsg": "ok"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	stdout, stderr, err := runCLI(t, []string{"draft", "list", "--json", "--base-url", server.URL})
	if err != nil {
		t.Fatalf("draft list: %v (stderr=%s)", err, stderr)
	}
	var listResp map[string]any
	if err := json.Unmarshal([]byte(stdout), &listResp); err != nil {
		t.Fatalf("parse list json: %v", err)
	}

	stdout, stderr, err = runCLI(t, []string{"draft", "delete", "m1", "--json", "--base-url", server.URL})
	if err != nil {
		t.Fatalf("draft delete: %v (stderr=%s)", err, stderr)
	}
	if !bytes.Contains([]byte(stdout), []byte("\"deleted\"")) {
		t.Fatalf("expected deleted json, got %s", stdout)
	}
}

func runCLI(t *testing.T, args []string) (string, string, error) {
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return "", "", err
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		return "", "", err
	}
	oldOut := os.Stdout
	oldErr := os.Stderr
	os.Stdout = stdoutW
	os.Stderr = stderrW

	var outBuf, errBuf bytes.Buffer
	readDone := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(&outBuf, stdoutR)
		readDone <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(&errBuf, stderrR)
		readDone <- struct{}{}
	}()

	err = Execute(args)
	_ = stdoutW.Close()
	_ = stderrW.Close()
	<-readDone
	<-readDone

	os.Stdout = oldOut
	os.Stderr = oldErr

	return outBuf.String(), errBuf.String(), err
}
