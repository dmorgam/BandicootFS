package corpus

import "unicode/utf8"

const (
	defaultLimit = 50
	maxLimit     = 200
	maxLineLen   = 300
)

// Page is a paginated slice of results.
type Page[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	NextOffset int `json:"next_offset"` // -1 when no more results
}

// clampLimit applies the default and cap to a caller-supplied limit.
func clampLimit(n int) int {
	switch {
	case n <= 0:
		return defaultLimit
	case n > maxLimit:
		return maxLimit
	default:
		return n
	}
}

// page returns the window [offset, offset+limit) of items.
func page[T any](items []T, offset, limit int) Page[T] {
	total := len(items)
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	next := -1
	if end < total {
		next = end
	}
	window := items[offset:end]
	if window == nil {
		window = []T{} // marshal as [] not null
	}
	return Page[T]{Items: window, Total: total, NextOffset: next}
}

// truncate caps s to n bytes, marking elision. It never splits a UTF-8 rune.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n] + "…"
}
