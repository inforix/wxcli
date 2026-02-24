package cmd

import (
	"fmt"
	"strings"

	"wxcli/src/internal/markup"
)

func normalizeDraftFormat(format, content string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(format))
	if value == "" {
		value = "auto"
	}
	switch value {
	case "auto":
		return detectDraftFormat(content), nil
	case "html", "markdown":
		return value, nil
	default:
		return "", usage("format must be html|markdown|auto")
	}
}

func detectDraftFormat(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return "html"
	}
	if strings.HasPrefix(trimmed, "<") && strings.Contains(trimmed, ">") {
		return "html"
	}
	// Basic heuristics for Markdown markers.
	if strings.Contains(trimmed, "\n#") ||
		strings.Contains(trimmed, "\n- ") ||
		strings.Contains(trimmed, "\n* ") ||
		strings.Contains(trimmed, "\n1. ") ||
		strings.Contains(trimmed, "```") ||
		strings.Contains(trimmed, "](") {
		return "markdown"
	}
	return "html"
}

func renderMarkdown(input string) (string, error) {
	html, err := renderMarkdownBase(input)
	if err != nil {
		return "", err
	}
	return finalizeMarkdownHTML(html)
}

func renderMarkdownBase(input string) (string, error) {
	html, err := markup.MarkdownToHTML(input)
	if err != nil {
		return "", err
	}
	html, err = markup.StripNotionMetadataHTML(html)
	if err != nil {
		return "", err
	}
	html, err = markup.WrapHeadingContent(html)
	if err != nil {
		return "", err
	}
	html, err = markup.NormalizeCodeLeaf(html)
	if err != nil {
		return "", err
	}
	html, err = markup.NormalizeCodeBlocks(html)
	if err != nil {
		return "", err
	}
	html, err = markup.CleanupProseMirrorArtifacts(html)
	if err != nil {
		return "", err
	}
	html, err = markup.ConstrainCodeWidth(html)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(html), nil
}

func finalizeMarkdownHTML(html string) (string, error) {
	out := html
	var err error
	out, err = markup.TightenListParagraphs(out)
	if err != nil {
		return "", err
	}
	out, err = markup.RemoveListBreaks(out)
	if err != nil {
		return "", err
	}
	out, err = markup.StripNewlines(out)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func inlineCSS(html, css string) (string, error) {
	if strings.TrimSpace(css) == "" {
		return html, nil
	}
	out, err := markup.InlineCSS(html, css)
	if err != nil {
		return "", fmt.Errorf("inline css: %w", err)
	}
	return strings.TrimSpace(out), nil
}
