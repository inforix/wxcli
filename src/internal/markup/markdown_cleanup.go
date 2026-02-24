package markup

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var bylinePattern = regexp.MustCompile(`^\*?\s*作者[:：].+\*?\s*$`)
var markdownH1Pattern = regexp.MustCompile(`^#\s+.+$`)

// StripMarkdownFrontMatter removes a leading YAML front matter block.
func StripMarkdownFrontMatter(input string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}
	trimmed := strings.TrimPrefix(input, "\ufeff")
	lines := strings.Split(trimmed, "\n")
	if len(lines) == 0 {
		return input
	}
	if strings.TrimSpace(lines[0]) != "---" {
		return input
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return input
	}
	rest := strings.Join(lines[end+1:], "\n")
	rest = strings.TrimPrefix(rest, "\n")
	return rest
}

// StripLeadingMarkdownByline removes a leading author byline if present.
func StripLeadingMarkdownByline(input string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}
	lines := strings.Split(input, "\n")
	idx := 0
	for idx < len(lines) && strings.TrimSpace(lines[idx]) == "" {
		idx++
	}
	if idx >= len(lines) {
		return input
	}
	if !bylinePattern.MatchString(strings.TrimSpace(lines[idx])) {
		return input
	}
	lines = append(lines[:idx], lines[idx+1:]...)
	for idx < len(lines) && strings.TrimSpace(lines[idx]) == "" {
		lines = append(lines[:idx], lines[idx+1:]...)
	}
	return strings.Join(lines, "\n")
}

// StripLeadingMarkdownH1 removes a leading ATX H1 heading if present.
func StripLeadingMarkdownH1(input string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}
	lines := strings.Split(input, "\n")
	idx := 0
	for idx < len(lines) && strings.TrimSpace(lines[idx]) == "" {
		idx++
	}
	if idx >= len(lines) {
		return input
	}
	if !markdownH1Pattern.MatchString(strings.TrimSpace(lines[idx])) {
		return input
	}
	lines = append(lines[:idx], lines[idx+1:]...)
	for idx < len(lines) && strings.TrimSpace(lines[idx]) == "" {
		lines = append(lines[:idx], lines[idx+1:]...)
	}
	return strings.Join(lines, "\n")
}

// StripTitleHeadingHTML removes the first H1 that matches the provided title.
func StripTitleHeadingHTML(html, title string) (string, error) {
	if strings.TrimSpace(html) == "" || strings.TrimSpace(title) == "" {
		return html, nil
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body>" + html + "</body></html>"))
	if err != nil {
		return "", err
	}
	target := strings.TrimSpace(title)
	doc.Find("h1").EachWithBreak(func(_ int, h1 *goquery.Selection) bool {
		text := strings.TrimSpace(h1.Text())
		if strings.EqualFold(text, target) {
			h1.Remove()
			return false
		}
		return true
	})
	body := doc.Find("body")
	if body.Length() == 0 {
		return strings.TrimSpace(html), nil
	}
	out, err := body.Html()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
