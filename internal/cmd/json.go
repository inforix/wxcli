package cmd

import (
	"context"
	"encoding/json"
	"os"

	"wxcli/internal/outfmt"
)

func writeJSON(ctx context.Context, v any) error {
	if !outfmt.IsJSON(ctx) {
		return nil
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
