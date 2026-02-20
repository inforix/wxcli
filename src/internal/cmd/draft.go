package cmd

import "context"

type DraftCmd struct {
	Add    DraftAddCmd    `cmd:"" name:"add" help:"Add a draft"`
	Get    DraftGetCmd    `cmd:"" name:"get" help:"Get a draft"`
	List   DraftListCmd   `cmd:"" name:"list" help:"List drafts"`
	Delete DraftDeleteCmd `cmd:"" name:"delete" help:"Delete a draft"`
}

func (c *DraftCmd) Run(_ context.Context, _ *RootFlags) error {
	return nil
}
