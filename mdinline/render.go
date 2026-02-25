package mdinline

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

// RenderMarkdownWithInlineCSS converts Markdown to HTML and inlines CSS rules
// according to a simplified CSS cascade engine.
func RenderMarkdownWithInlineCSS(markdown string, cssText string, media string) (string, error) {
	html, err := renderMarkdown(markdown)
	if err != nil {
		return "", err
	}

	rules, err := ParseCSS(cssText)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(wrapHTMLFragment(html)))
	if err != nil {
		return "", fmt.Errorf("parse html: %w", err)
	}

	if err := ApplyInlineCSS(doc, rules, media); err != nil {
		return "", err
	}

	body := doc.Find("body")
	if body.Length() == 0 {
		return strings.TrimSpace(html), nil
	}
	out, err := body.Html()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(unescapeStyleEntities(out)), nil
}

func renderMarkdown(input string) (string, error) {
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(gmhtml.WithUnsafe()),
	)
	if err := md.Convert([]byte(input), &buf); err != nil {
		return "", fmt.Errorf("convert markdown: %w", err)
	}
	return buf.String(), nil
}

func wrapHTMLFragment(fragment string) string {
	return "<html><body>" + fragment + "</body></html>"
}

// unescapeStyleEntities avoids HTML entity escaping inside style attributes,
// which can cause downstream sanitizers to drop the styles.
func unescapeStyleEntities(html string) string {
	const key = `style="`
	var b strings.Builder
	i := 0
	for {
		idx := strings.Index(html[i:], key)
		if idx < 0 {
			b.WriteString(html[i:])
			break
		}
		idx += i
		b.WriteString(html[i : idx+len(key)])
		start := idx + len(key)
		end := start
		for end < len(html) && html[end] != '"' {
			end++
		}
		if end > start {
			value := html[start:end]
			value = strings.ReplaceAll(value, "&#39;", "'")
			value = strings.ReplaceAll(value, "&apos;", "'")
			b.WriteString(value)
		}
		if end < len(html) {
			b.WriteByte('"')
			i = end + 1
		} else {
			i = end
		}
	}
	return b.String()
}
