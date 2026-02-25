package mdinline

import (
	"math"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
)

type styleValue struct {
	Value       string
	Specificity [3]int
	Order       int
}

type compiledRule struct {
	rule     CSSRule
	selector cascadia.Sel
}

var inlineSpecificity = [3]int{math.MaxInt32, math.MaxInt32, math.MaxInt32}

// ApplyInlineCSS walks the DOM, matches selectors, resolves cascade conflicts,
// and injects computed styles into inline style attributes.
func ApplyInlineCSS(doc *goquery.Document, rules []CSSRule, media string) error {
	compiled, err := compileRules(rules)
	if err != nil {
		return err
	}

	normalizedMedia := strings.ToLower(strings.TrimSpace(media))

	doc.Find("*").Each(func(_ int, sel *goquery.Selection) {
		styles := map[string]styleValue{}

		for _, rule := range compiled {
			if !mediaMatches(rule.rule.Media, normalizedMedia) {
				continue
			}
			node := sel.Get(0)
			if node == nil || !rule.selector.Match(node) {
				continue
			}
			applyDeclarations(styles, rule.rule.Declarations, rule.rule.Specificity, rule.rule.Order)
		}

		if inline, ok := sel.Attr("style"); ok && strings.TrimSpace(inline) != "" {
			if inlineDecls, err := ParseInlineStyle(inline); err == nil {
				applyDeclarations(styles, inlineDecls, inlineSpecificity, math.MaxInt32)
			}
		}

		if len(styles) > 0 {
			sel.SetAttr("style", formatStyle(styles))
		}
	})

	return nil
}

func compileRules(rules []CSSRule) ([]compiledRule, error) {
	compiled := make([]compiledRule, 0, len(rules))
	for _, rule := range rules {
		if len(rule.Selectors) == 0 {
			continue
		}
		selectorText := strings.TrimSpace(rule.Selectors[0])
		if selectorText == "" {
			continue
		}
		selector, err := cascadia.Parse(selectorText)
		if err != nil {
			// Skip unsupported selectors (e.g., pseudo-elements).
			continue
		}
		compiled = append(compiled, compiledRule{rule: rule, selector: selector})
	}
	return compiled, nil
}

func applyDeclarations(styles map[string]styleValue, declarations map[string]string, specificity [3]int, order int) {
	for prop, val := range declarations {
		prop = strings.ToLower(strings.TrimSpace(prop))
		val = strings.TrimSpace(val)
		if prop == "" || val == "" {
			continue
		}
		if existing, ok := styles[prop]; ok {
			if shouldOverride(existing, specificity, order) {
				styles[prop] = styleValue{Value: val, Specificity: specificity, Order: order}
			}
			continue
		}
		styles[prop] = styleValue{Value: val, Specificity: specificity, Order: order}
	}
}

// shouldOverride applies CSS cascade comparison rules.
// Higher specificity wins; if equal, later order wins.
func shouldOverride(existing styleValue, specificity [3]int, order int) bool {
	cmp := compareSpecificity(specificity, existing.Specificity)
	if cmp > 0 {
		return true
	}
	if cmp < 0 {
		return false
	}
	return order >= existing.Order
}

func compareSpecificity(a, b [3]int) int {
	if a[0] != b[0] {
		if a[0] > b[0] {
			return 1
		}
		return -1
	}
	if a[1] != b[1] {
		if a[1] > b[1] {
			return 1
		}
		return -1
	}
	if a[2] != b[2] {
		if a[2] > b[2] {
			return 1
		}
		return -1
	}
	return 0
}

func formatStyle(styles map[string]styleValue) string {
	keys := make([]string, 0, len(styles))
	for key := range styles {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, key := range keys {
		val := strings.TrimSpace(styles[key].Value)
		if val == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString(key)
		b.WriteString(": ")
		b.WriteString(val)
		b.WriteString(";")
	}
	return b.String()
}

func mediaMatches(ruleMedia string, requested string) bool {
	ruleMedia = strings.ToLower(strings.TrimSpace(ruleMedia))
	if ruleMedia == "" || ruleMedia == "all" {
		return true
	}
	if requested == "" {
		return false
	}
	for _, part := range strings.FieldsFunc(ruleMedia, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	}) {
		if part == requested {
			return true
		}
	}
	return false
}
