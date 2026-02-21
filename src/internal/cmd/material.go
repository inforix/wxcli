package cmd

import "context"

type MaterialCmd struct {
	Get        MaterialGetCmd        `cmd:"" name:"get" help:"Get a permanent material"`
	List       MaterialListCmd       `cmd:"" name:"list" help:"List permanent materials"`
	Delete     MaterialDeleteCmd     `cmd:"" name:"delete" help:"Delete a permanent material"`
	Count      MaterialCountCmd      `cmd:"" name:"count" help:"Get permanent material counts"`
	Upload     MaterialUploadCmd     `cmd:"" name:"upload" help:"Upload a permanent material"`
	AddNews    MaterialAddNewsCmd    `cmd:"" name:"add-news" help:"Upload a permanent news material"`
	UpdateNews MaterialUpdateNewsCmd `cmd:"" name:"update-news" help:"Update a permanent news material"`
}

func (c *MaterialCmd) Run(_ context.Context, _ *RootFlags) error {
	return nil
}
