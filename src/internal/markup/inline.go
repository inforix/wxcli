package markup

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aymerick/douceur/inliner"
)

func InlineCSS(html, css string) (string, error) {
	doc := "<html><head><style>" + css + "</style></head><body><div id=\"nice\">" + html + "</div></body></html>"
	inlined, err := inliner.Inline(doc)
	if err != nil {
		return "", err
	}
	parsed, err := goquery.NewDocumentFromReader(strings.NewReader(inlined))
	if err != nil {
		return "", err
	}
	parsed.Find("style").Remove()
	parsed.Find("*").Each(func(_ int, s *goquery.Selection) {
		if shouldPreserveClass(s) {
			return
		}
		s.RemoveAttr("class")
	})
	wrapper := parsed.Find("div#nice")
	if wrapper.Length() > 0 {
		propagateWrapperStyles(wrapper)
		inner, err := wrapper.Html()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(inner), nil
	}
	body := parsed.Find("body")
	if body.Length() == 0 {
		return strings.TrimSpace(inlined), nil
	}
	inner, err := body.Html()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(inner), nil
}

func shouldPreserveClass(sel *goquery.Selection) bool {
	if sel == nil {
		return false
	}
	if class, ok := sel.Attr("class"); ok {
		if strings.Contains(class, "code-snippet") {
			return true
		}
	}
	if sel.ParentsFiltered(".code-snippet__js, .code-snippet").Length() > 0 {
		return true
	}
	return false
}


func propagateWrapperStyles(wrapper *goquery.Selection) {
	if wrapper == nil {
		return
	}
	style, ok := wrapper.Attr("style")
	if !ok || strings.TrimSpace(style) == "" {
		return
	}
	parsed := parseInlineStyle(style)
	inherited := filterInheritableStyles(parsed)
	if len(inherited) == 0 {
		return
	}
	wrapper.Find("*").Each(func(_ int, sel *goquery.Selection) {
		mergeInlineStyle(sel, inherited)
	})
}
