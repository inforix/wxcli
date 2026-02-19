package errfmt

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/99designs/keyring"
	"github.com/alecthomas/kong"
)

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("API error (%d): %s", e.Code, e.Message)
}

func Format(err error) string {
	if err == nil {
		return ""
	}
	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return parseErr.Error()
	}
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return "Secret not found in keyring. Run: wxcli auth set --appid <id> --appsecret <secret>"
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Error()
	}
	return redactSecrets(err.Error())
}

var secretParams = regexp.MustCompile(`(?i)(secret|access_token)=([^&\s]+)`)

func redactSecrets(s string) string {
	return secretParams.ReplaceAllString(s, "$1=REDACTED")
}
