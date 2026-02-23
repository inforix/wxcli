package markup

import (
	stdhtml "html"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func TightenListParagraphs(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return html, nil
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body>" + html + "</body></html>"))
	if err != nil {
		return "", err
	}
	doc.Find("li > p").Each(func(_ int, s *goquery.Selection) {
		inner, err := s.Html()
		if err != nil {
			return
		}
		if strings.TrimSpace(inner) == "" {
			s.Remove()
			return
		}
		_ = s.ReplaceWithHtml(inner)
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

func RemoveListBreaks(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return html, nil
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body>" + html + "</body></html>"))
	if err != nil {
		return "", err
	}
	doc.Find("ul br, ol br, li br").Remove()
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

func ConstrainCodeWidth(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return html, nil
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body>" + html + "</body></html>"))
	if err != nil {
		return "", err
	}
	doc.Find("pre, div[data-code-block]").Each(func(_ int, s *goquery.Selection) {
		appendInlineStyle(s, "max-width: 100%; overflow-x: auto; white-space: pre;")
		s.Find("code").Each(func(_ int, codeSel *goquery.Selection) {
			appendInlineStyle(codeSel, "max-width: 100%; white-space: pre;")
		})
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

func appendInlineStyle(sel *goquery.Selection, style string) {
	if sel == nil || strings.TrimSpace(style) == "" {
		return
	}
	existing, _ := sel.Attr("style")
	if strings.TrimSpace(existing) == "" {
		sel.SetAttr("style", style)
		return
	}
	if !strings.HasSuffix(strings.TrimSpace(existing), ";") {
		existing += ";"
	}
	sel.SetAttr("style", existing+" "+style)
}

func ConvertPreToDiv(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return html, nil
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body>" + html + "</body></html>"))
	if err != nil {
		return "", err
	}
	doc.Find("pre").Each(func(_ int, pre *goquery.Selection) {
		inner, err := pre.Html()
		if err != nil {
			return
		}
		existingStyle, _ := pre.Attr("style")
		style := strings.TrimSpace(existingStyle)
		if style != "" && !strings.HasSuffix(style, ";") {
			style += ";"
		}
		style += " font-family: monospace;"
		div := `<div data-code-block="1" style="` + style + `">` + inner + `</div>`
		_ = pre.ReplaceWithHtml(div)
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

func NormalizeCodeBlocks(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return html, nil
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body>" + html + "</body></html>"))
	if err != nil {
		return "", err
	}
	doc.Find("pre").Each(func(_ int, pre *goquery.Selection) {
		code := pre.Find("code")
		var content string
		if code.Length() > 0 {
			content = code.Text()
		} else {
			content = pre.Text()
		}
		lines := strings.Split(content, "\n")
		var b strings.Builder
		b.WriteString(`<section class="code-snippet__js">`)
		b.WriteString(`<pre class="code-snippet__js code-snippet code-snippet_nowrap" data-lang="json">`)
		for _, line := range lines {
			escaped := stdhtml.EscapeString(line)
			if strings.TrimSpace(escaped) == "" {
				b.WriteString(`<code>&nbsp;</code>`)
			} else {
				b.WriteString(`<code>`)
				b.WriteString(escaped)
				b.WriteString(`</code>`)
			}
		}
		b.WriteString(`</pre></section>`)
		_ = pre.ReplaceWithHtml(b.String())
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

func CleanupProseMirrorArtifacts(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return html, nil
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body>" + html + "</body></html>"))
	if err != nil {
		return "", err
	}
	// Remove ProseMirror br tags from everywhere
	doc.Find("br.ProseMirror-trailingBreak").Remove()
	// Remove [leaf] attributes from all elements
	doc.Find("[leaf]").Each(func(_ int, s *goquery.Selection) {
		s.RemoveAttr("leaf")
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

func NormalizeCodeLeaf(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return html, nil
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body>" + html + "</body></html>"))
	if err != nil {
		return "", err
	}
	for {
		nodes := doc.Find("[leaf]")
		if nodes.Length() == 0 {
			break
		}
		nodes.Each(func(_ int, n *goquery.Selection) {
			inner, err := n.Html()
			if err != nil {
				return
			}
			if strings.TrimSpace(inner) == "" {
				_ = n.ReplaceWithHtml("<div></div>")
				return
			}
			_ = n.ReplaceWithHtml("<div>" + inner + "</div>")
		})
	}
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

func StripNotionMetadataHTML(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return html, nil
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body>" + html + "</body></html>"))
	if err != nil {
		return "", err
	}
	doc.Find("h2").Each(func(_ int, h2 *goquery.Selection) {
		text := h2.Text()
		if strings.Contains(text, "properties_json:") && strings.Contains(text, "last_edited_time") {
			if n := h2.Get(0); n != nil && n.PrevSibling != nil && n.PrevSibling.Type == n.Type && n.PrevSibling.Data == "hr" {
				if n.PrevSibling.Parent != nil {
					n.PrevSibling.Parent.RemoveChild(n.PrevSibling)
				}
			}
			h2.Remove()
		}
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
