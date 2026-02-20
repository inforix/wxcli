package cmd

import (
	"strings"
	"testing"
)

func TestHelpShowsConfigAndKeyring(t *testing.T) {
	out, _, err := runCLI(t, []string{"--help"})
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if !strings.Contains(out, "config") || !strings.Contains(out, "keyring") {
		t.Fatalf("expected help to include config and keyring info")
	}
}
