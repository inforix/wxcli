package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"wxcli/src/internal/config"
	"wxcli/src/internal/errfmt"
	"wxcli/src/internal/outfmt"
	"wxcli/src/internal/secrets"
	"wxcli/src/internal/ui"
)

type RootFlags struct {
	Color   string `help:"Color output: auto|always|never" default:"auto"`
	JSON    bool   `help:"Output JSON to stdout" short:"j"`
	Plain   bool   `help:"Output plain text to stdout" short:"p"`
	NoInput bool   `name:"no-input" help:"Never prompt; fail instead"`
	BaseURL string `name:"base-url" help:"Override API base URL"`
	Verbose bool   `help:"Enable verbose logging" short:"v"`
}

type rootFlagsCtxKey struct{}

type CLI struct {
	RootFlags `embed:""`

	Version kong.VersionFlag `help:"Print version and exit"`

	Auth     AuthCmd     `cmd:"" help:"Auth and credentials"`
	Config   ConfigCmd   `cmd:"" help:"Configuration"`
	Draft    DraftCmd    `cmd:"" help:"Draftbox management"`
	Material MaterialCmd `cmd:"" help:"Material management"`
}

type exitPanic struct{ code int }

func Execute(args []string) (err error) {
	parser, cli, err := newParser()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				if ep.code == 0 {
					err = nil
					return
				}
				err = &ExitError{Code: ep.code, Err: errors.New("exited")}
				return
			}
			panic(r)
		}
	}()

	kctx, err := parser.Parse(args)
	if err != nil {
		parsedErr := wrapParseError(err)
		_, _ = fmt.Fprintln(os.Stderr, errfmt.Format(parsedErr))
		return parsedErr
	}

	ctx := context.Background()
	mode, err := outfmt.FromFlags(cli.JSON, cli.Plain)
	if err != nil {
		return usage(err.Error())
	}
	ctx = outfmt.WithMode(ctx, mode)

	uiColor := cli.Color
	if outfmt.IsJSON(ctx) || outfmt.IsPlain(ctx) {
		uiColor = "never"
	}
	uiInstance, err := ui.New(ui.Options{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Color:  uiColor,
	})
	if err != nil {
		return err
	}
	ctx = ui.WithUI(ctx, uiInstance)
	ctx = context.WithValue(ctx, rootFlagsCtxKey{}, &cli.RootFlags)

	kctx.BindTo(ctx, (*context.Context)(nil))
	kctx.Bind(&cli.RootFlags)

	if err = kctx.Run(); err != nil {
		if ExitCode(err) == 0 {
			return nil
		}
		msg := errfmt.Format(err)
		if msg != "" {
			_, _ = fmt.Fprintln(os.Stderr, msg)
		}
		return err
	}
	return nil
}

func RootFlagsFromContext(ctx context.Context) *RootFlags {
	if v := ctx.Value(rootFlagsCtxKey{}); v != nil {
		if flags, ok := v.(*RootFlags); ok {
			return flags
		}
	}
	return nil
}

func newParser() (*kong.Kong, *CLI, error) {
	cli := &CLI{}
	parser, err := kong.New(
		cli,
		kong.Name("wxcli"),
		kong.Description(helpDescription()),
		kong.Vars(kong.Vars{
			"version": VersionString(),
		}),
		kong.Writers(os.Stdout, os.Stderr),
		kong.Exit(func(code int) { panic(exitPanic{code: code}) }),
	)
	if err != nil {
		return nil, nil, err
	}
	return parser, cli, nil
}

func helpDescription() string {
	configPath, err := config.ConfigPath()
	configLine := "unknown"
	if err != nil {
		configLine = fmt.Sprintf("error: %v", err)
	} else if configPath != "" {
		configLine = configPath
	}
	backendInfo, err := secrets.ResolveKeyringBackendInfo()
	backendLine := "unknown"
	if err != nil {
		backendLine = fmt.Sprintf("error: %v", err)
	} else if backendInfo.Value != "" {
		backendLine = fmt.Sprintf("%s (source: %s)", backendInfo.Value, backendInfo.Source)
	}
	return fmt.Sprintf("Weixin CLI for draftbox and material management\n\nConfig:\n  file: %s\n  keyring backend: %s", configLine, backendLine)
}

func wrapParseError(err error) error {
	if err == nil {
		return nil
	}
	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return &ExitError{Code: 2, Err: parseErr}
	}
	return err
}
