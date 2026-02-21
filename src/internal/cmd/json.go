package cmd

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"wxcli/src/internal/outfmt"
)

func writeJSON(ctx context.Context, v any) error {
	if !outfmt.IsJSON(ctx) {
		return nil
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func writeJSONBytes(ctx context.Context, data []byte) error {
	if !outfmt.IsJSON(ctx) {
		return nil
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil
	}
	if _, err := os.Stdout.Write([]byte(trimmed)); err != nil {
		return err
	}
	if !strings.HasSuffix(trimmed, "\n") {
		_, err := os.Stdout.Write([]byte("\n"))
		return err
	}
	return nil
}
