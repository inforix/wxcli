package ui

import (
	"os"

	"golang.org/x/term"
)

func readPassword() ([]byte, error) {
	return term.ReadPassword(int(os.Stdin.Fd()))
}
