package mdinline

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

// CSSRule represents a parsed CSS rule with selector(s), declarations, and metadata.
type CSSRule struct {
	Selectors    []string
	Declarations map[string]string
	Specificity  [3]int
	Order        int
	Media        string
}

type mediaContext struct {
	media string
}

// ParseCSS parses a stylesheet into CSS rules using a real CSS parser.
func ParseCSS(input string) ([]CSSRule, error) {
	p := css.NewParser(parse.NewInputString(input), false)

	var rules []CSSRule
	var selectors []string
	var declarations map[string]string
	var inRuleset bool
	order := 0
	var mediaStack []mediaContext

	for {
		gt, _, data := p.Next()
		if gt == css.ErrorGrammar {
			if err := p.Err(); err != io.EOF {
				return nil, fmt.Errorf("parse css: %w", err)
			}
			break
		}
		switch gt {
		case css.BeginAtRuleGrammar:
			atName := strings.ToLower(string(data))
			if atName == "@media" {
				media := parseMediaTypes(p.Values())
				mediaStack = append(mediaStack, mediaContext{media: media})
			} else {
				mediaStack = append(mediaStack, mediaContext{})
			}
		case css.EndAtRuleGrammar:
			if len(mediaStack) > 0 {
				mediaStack = mediaStack[:len(mediaStack)-1]
			}
		case css.QualifiedRuleGrammar:
			selector := strings.TrimSpace(tokensToString(data, p.Values()))
			if selector != "" {
				selectors = append(selectors, selector)
			}
		case css.BeginRulesetGrammar:
			selector := strings.TrimSpace(tokensToString(data, p.Values()))
			if selector != "" {
				selectors = append(selectors, selector)
			}
			inRuleset = true
			declarations = make(map[string]string)
		case css.DeclarationGrammar, css.CustomPropertyGrammar:
			if !inRuleset {
				continue
			}
			property := strings.ToLower(strings.TrimSpace(string(data)))
			if property == "" {
				continue
			}
			value := normalizeCSSValue(tokensToString(nil, p.Values()))
			if value == "" {
				continue
			}
			declarations[property] = value
		case css.EndRulesetGrammar:
			if !inRuleset {
				selectors = nil
				continue
			}
			media := currentMedia(mediaStack)
			for _, selector := range selectors {
				specificity, err := ComputeSpecificity(selector)
				if err != nil {
					return nil, fmt.Errorf("specificity for %q: %w", selector, err)
				}
				rules = append(rules, CSSRule{
					Selectors:    []string{selector},
					Declarations: copyDeclarations(declarations),
					Specificity:  specificity,
					Order:        order,
					Media:        media,
				})
			}
			order++
			selectors = nil
			inRuleset = false
		}
	}

	return rules, nil
}

// ParseInlineStyle parses an inline style attribute into declarations.
func ParseInlineStyle(style string) (map[string]string, error) {
	p := css.NewParser(parse.NewInputString(style), true)
	declarations := make(map[string]string)

	for {
		gt, _, data := p.Next()
		if gt == css.ErrorGrammar {
			if err := p.Err(); err != io.EOF {
				return nil, fmt.Errorf("parse inline style: %w", err)
			}
			break
		}
		if gt != css.DeclarationGrammar && gt != css.CustomPropertyGrammar {
			continue
		}
		property := strings.ToLower(strings.TrimSpace(string(data)))
		if property == "" {
			continue
		}
		value := normalizeCSSValue(tokensToString(nil, p.Values()))
		if value == "" {
			continue
		}
		declarations[property] = value
	}

	return declarations, nil
}

func tokensToString(head []byte, values []css.Token) string {
	var buf bytes.Buffer
	buf.Write(head)
	for _, val := range values {
		buf.Write(val.Data)
	}
	return buf.String()
}

func normalizeCSSValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	// Avoid HTML entity escaping of double quotes inside style attributes.
	// WeChat tends to drop styles when values contain HTML-escaped quotes.
	value = strings.ReplaceAll(value, "\"", "'")
	return value
}

func parseMediaTypes(tokens []css.Token) string {
	seen := map[string]struct{}{}
	for _, token := range tokens {
		if token.TokenType != css.IdentToken {
			continue
		}
		ident := strings.ToLower(string(token.Data))
		switch ident {
		case "all", "screen", "print":
			seen[ident] = struct{}{}
		}
	}
	if _, ok := seen["all"]; ok {
		return "all"
	}
	order := []string{"screen", "print"}
	var parts []string
	for _, name := range order {
		if _, ok := seen[name]; ok {
			parts = append(parts, name)
		}
	}
	return strings.Join(parts, ",")
}

func currentMedia(stack []mediaContext) string {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].media != "" {
			return stack[i].media
		}
	}
	return ""
}

func copyDeclarations(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for key, val := range in {
		out[key] = val
	}
	return out
}
