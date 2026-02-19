package outfmt

import (
	"context"
	"errors"
)

type Mode int

const (
	ModeText Mode = iota
	ModeJSON
	ModePlain
)

type ctxKey struct{}

func FromFlags(json, plain bool) (Mode, error) {
	if json && plain {
		return ModeText, errors.New("cannot combine --json and --plain")
	}
	if json {
		return ModeJSON, nil
	}
	if plain {
		return ModePlain, nil
	}
	return ModeText, nil
}

func WithMode(ctx context.Context, mode Mode) context.Context {
	return context.WithValue(ctx, ctxKey{}, mode)
}

func ModeFromContext(ctx context.Context) Mode {
	if v := ctx.Value(ctxKey{}); v != nil {
		if m, ok := v.(Mode); ok {
			return m
		}
	}
	return ModeText
}

func IsJSON(ctx context.Context) bool {
	return ModeFromContext(ctx) == ModeJSON
}

func IsPlain(ctx context.Context) bool {
	return ModeFromContext(ctx) == ModePlain
}
