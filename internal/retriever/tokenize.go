package retriever

import (
	"strings"
	"unicode"
)

// tokenize lowercases text, splits on non-alphanumeric runes, and drops
// stopwords and single-character tokens.
func tokenize(text string) []string {
	raw := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	out := raw[:0] // filter in place
	for _, t := range raw {
		if len(t) < 2 || stopwords[t] {
			continue
		}
		out = append(out, t)
	}
	return out
}

// uniq returns the set of distinct tokens.
func uniq(tokens []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		set[t] = struct{}{}
	}
	return set
}
