package markup

import (
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var inheritableStyles = map[string]struct{}{
	"color":          {},
	"direction":      {},
	"font":           {},
	"font-family":    {},
	"font-size":      {},
	"font-style":     {},
	"font-stretch":   {},
	"font-variant":   {},
	"font-weight":    {},
	"letter-spacing": {},
	"line-height":    {},
	"text-align":     {},
	"text-indent":    {},
	"text-transform": {},
	"visibility":     {},
	"white-space":    {},
	"word-spacing":   {},
}

func transferInheritableStyles(from, to *goquery.Selection) {
	if from == nil || to == nil {
		return
	}
	style, ok := from.Attr("style")
	if !ok || strings.TrimSpace(style) == "" {
		return
	}
	parsed := parseInlineStyle(style)
	inheritable := filterInheritableStyles(parsed)
	if len(inheritable) == 0 {
		return
	}
	mergeInlineStyle(to, inheritable)
}

func filterInheritableStyles(parsed map[string]string) map[string]string {
	if len(parsed) == 0 {
		return nil
	}
	out := make(map[string]string)
	for prop, val := range parsed {
		if _, ok := inheritableStyles[prop]; !ok {
			continue
		}
		if strings.TrimSpace(val) == "" {
			continue
		}
		out[prop] = val
	}
	return out
}

func mergeInlineStyle(sel *goquery.Selection, add map[string]string) {
	if sel == nil || len(add) == 0 {
		return
	}
	style, _ := sel.Attr("style")
	existing := parseInlineStyle(style)
	missing := make(map[string]string)
	for prop, val := range add {
		if _, ok := existing[prop]; ok {
			continue
		}
		missing[prop] = val
	}
	if len(missing) == 0 {
		return
	}
	newStyle := strings.TrimSpace(style)
	if newStyle != "" && !strings.HasSuffix(newStyle, ";") {
		newStyle += ";"
	}
	keys := make([]string, 0, len(missing))
	for prop := range missing {
		keys = append(keys, prop)
	}
	sort.Strings(keys)
	for _, prop := range keys {
		val := strings.TrimSpace(missing[prop])
		if val == "" {
			continue
		}
		if newStyle != "" && !strings.HasSuffix(newStyle, " ") {
			newStyle += " "
		}
		newStyle += prop + ": " + val + ";"
	}
	sel.SetAttr("style", strings.TrimSpace(newStyle))
}

func parseInlineStyle(style string) map[string]string {
	out := make(map[string]string)
	for _, chunk := range strings.Split(style, ";") {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		parts := strings.SplitN(chunk, ":", 2)
		if len(parts) != 2 {
			continue
		}
		prop := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		if prop == "" || val == "" {
			continue
		}
		out[prop] = val
	}
	return out
}
