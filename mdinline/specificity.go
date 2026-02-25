package mdinline

import (
	"fmt"
	"unicode"
)

// ComputeSpecificity calculates selector specificity as (ID, class, element).
// It supports the selector subset required by this project and avoids regex parsing.
func ComputeSpecificity(selector string) ([3]int, error) {
	var idCount, classCount, elementCount int
	expectType := true
	inAttr := false
	inIdent := false

	runes := []rune(selector)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch r {
		case '#':
			idCount++
			expectType = false
			i = consumeIdent(runes, i+1)
			continue
		case '.':
			classCount++
			expectType = false
			i = consumeIdent(runes, i+1)
			continue
		case '[':
			classCount++
			inAttr = true
			expectType = false
		case ']':
			inAttr = false
		case '*':
			expectType = false
		case ':':
			// treat pseudo-classes as class specificity
			classCount++
			expectType = false
			i = consumeIdent(runes, i+1)
			continue
		case '>', '+', '~', ',':
			expectType = true
		case ' ', '\n', '\t', '\r', '\f':
			expectType = true
		}

		if inAttr {
			continue
		}

		if isIdentStart(r) {
			if expectType {
				elementCount++
				expectType = false
			}
			inIdent = true
			i = consumeIdent(runes, i+1)
			continue
		}
		if inIdent {
			inIdent = false
		}
	}

	if idCount < 0 || classCount < 0 || elementCount < 0 {
		return [3]int{}, fmt.Errorf("invalid specificity")
	}

	return [3]int{idCount, classCount, elementCount}, nil
}

func isIdentStart(r rune) bool {
	return r == '_' || r == '-' || unicode.IsLetter(r)
}

func isIdentChar(r rune) bool {
	return r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func consumeIdent(runes []rune, start int) int {
	idx := start
	for idx < len(runes) && isIdentChar(runes[idx]) {
		idx++
	}
	return idx - 1
}
