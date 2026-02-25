package mdinline

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestUniversalAndElementOverride(t *testing.T) {
	markdown := "## header 2\n\nparagraph"
	css := `* { font-family: sans-serif; line-height: 2em; }
h2 { line-height: 1.5em; font-weight: bold; }`

	out, err := RenderMarkdownWithInlineCSS(markdown, css, "screen")
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	h2Styles := styleMap(t, out, "h2")
	assertStyle(t, h2Styles, "font-family", "sans-serif")
	assertStyle(t, h2Styles, "line-height", "1.5em")
	assertStyle(t, h2Styles, "font-weight", "bold")

	pStyles := styleMap(t, out, "p")
	assertStyle(t, pStyles, "font-family", "sans-serif")
	assertStyle(t, pStyles, "line-height", "2em")
}

func TestSpecificityOrder(t *testing.T) {
	markdown := `<div id="main"><h2 class="title">hello</h2></div>`
	css := `* { color: red; }
h2 { color: blue; }
.title { color: green; }
#main h2 { color: purple; }`

	out, err := RenderMarkdownWithInlineCSS(markdown, css, "screen")
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	h2Styles := styleMap(t, out, "h2")
	assertStyle(t, h2Styles, "color", "purple")
}

func TestLaterRuleWins(t *testing.T) {
	markdown := "paragraph"
	css := "p { color: red; }\np { color: green; }"

	out, err := RenderMarkdownWithInlineCSS(markdown, css, "screen")
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	pStyles := styleMap(t, out, "p")
	assertStyle(t, pStyles, "color", "green")
}

func TestMediaQueries(t *testing.T) {
	markdown := "paragraph"
	css := "p { color: red; }\n\n@media print {\n  p { color: blue; }\n}"

	screenOut, err := RenderMarkdownWithInlineCSS(markdown, css, "screen")
	if err != nil {
		t.Fatalf("render screen: %v", err)
	}
	screenStyles := styleMap(t, screenOut, "p")
	assertStyle(t, screenStyles, "color", "red")

	printOut, err := RenderMarkdownWithInlineCSS(markdown, css, "print")
	if err != nil {
		t.Fatalf("render print: %v", err)
	}
	printStyles := styleMap(t, printOut, "p")
	assertStyle(t, printStyles, "color", "blue")
}

func TestDescendantAndChildSelectors(t *testing.T) {
	markdown := `<div>
  <p>direct</p>
  <div><span><p>nested</p></span></div>
</div>`
	css := "div p { color: blue; }\n" +
		"div > p { font-weight: bold; }"

	out, err := RenderMarkdownWithInlineCSS(markdown, css, "screen")
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	direct := styleMap(t, out, "div > p")
	assertStyle(t, direct, "color", "blue")
	assertStyle(t, direct, "font-weight", "bold")

	nested := styleMap(t, out, "div div span > p")
	assertStyle(t, nested, "color", "blue")
	if _, ok := nested["font-weight"]; ok {
		t.Fatalf("expected nested p to not have font-weight, got %v", nested["font-weight"])
	}
}

func styleMap(t *testing.T, html string, selector string) map[string]string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body>" + html + "</body></html>"))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}
	selection := doc.Find(selector).First()
	style, _ := selection.Attr("style")
	return parseStyle(style)
}

func parseStyle(style string) map[string]string {
	out := map[string]string{}
	parts := strings.Split(style, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		prop := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if prop == "" || val == "" {
			continue
		}
		out[prop] = val
	}
	return out
}

func assertStyle(t *testing.T, styles map[string]string, prop string, expected string) {
	actual, ok := styles[prop]
	if !ok {
		t.Fatalf("expected %s to be set", prop)
	}
	if actual != expected {
		t.Fatalf("expected %s=%s, got %s", prop, expected, actual)
	}
}
