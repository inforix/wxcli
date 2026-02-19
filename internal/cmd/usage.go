package cmd

import "fmt"

func usage(message string) error {
	if message == "" {
		message = "invalid usage"
	}
	return &ExitError{Code: 2, Err: fmt.Errorf("%s", message)}
}
