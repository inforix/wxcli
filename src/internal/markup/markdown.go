package markup

import (
	"bytes"

	"github.com/yuin/goldmark"
)

func MarkdownToHTML(input string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(input), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
