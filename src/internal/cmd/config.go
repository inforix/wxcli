package cmd

import (
	"context"
	"os"

	"wxcli/src/internal/config"
	"wxcli/src/internal/outfmt"
)

type ConfigCmd struct {
	Path ConfigPathCmd `cmd:"" name:"path" help:"Print config file path"`
}

type ConfigPathCmd struct{}

func (c *ConfigPathCmd) Run(ctx context.Context, _ *RootFlags) error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}
	if outfmt.IsJSON(ctx) {
		return writeJSON(ctx, map[string]any{"path": path})
	}
	_, err = os.Stdout.WriteString(path + "\n")
	return err
}
