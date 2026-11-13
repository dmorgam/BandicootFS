package corpus

import (
	"bufio"
	"context"
	"regexp"
	"slices"
)

const maxContext = 10 // cap on context lines per side of a match

// Match is one line matching a grep pattern, plus optional context lines.
type Match struct {
	Source string   `json:"source"`
	Line   int      `json:"line"`
	Text   string   `json:"text"`
	Before []string `json:"before,omitempty"` // lines right above Line, in order
	After  []string `json:"after,omitempty"`  // lines right below Line, in order
}

// GrepOptions tunes Grep. The zero value is a case-sensitive search without
// context lines, first page.
type GrepOptions struct {
	IgnoreCase bool
	Context    int // context lines before and after each match (max 10)
	Offset     int // matches to skip
	Limit      int // max matches returned (default 50, max 200)
}

// Grep searches files under sub (root-relative) for pattern, paginated. Only
// the requested page is kept in memory; the other matches are just counted.
func (c *Corpus) Grep(ctx context.Context, sub, pattern string, opts GrepOptions) (Page[Match], error) {
	dir, err := c.Resolve(sub)
	if err != nil {
		return Page[Match]{}, err
	}
	if pattern == "" {
		return Page[Match]{Items: []Match{}}, nil
	}
	if opts.IgnoreCase {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return Page[Match]{}, err
	}
	n := min(max(opts.Context, 0), maxContext)
	offset, limit := max(opts.Offset, 0), clampLimit(opts.Limit)

	matches := []Match{} // marshal as [] not null
	total := 0
	err = c.walk(ctx, dir, func(path string) error {
		f, err := c.FS().Open(path)
		if err != nil {
			return nil // skip unreadable files and symlinks leaving root
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024) // tolerate long lines

		// before holds the last n lines of this file; open indexes the kept
		// matches that are still collecting After lines.
		var before []string
		var open []int
		line := 0
		for sc.Scan() {
			line++
			t := truncate(sc.Text(), maxLineLen)
			for _, i := range open {
				matches[i].After = append(matches[i].After, t)
			}
			open = slices.DeleteFunc(open, func(i int) bool { return len(matches[i].After) == n })

			if re.MatchString(sc.Text()) {
				if total >= offset && total < offset+limit {
					matches = append(matches, Match{Source: path, Line: line, Text: t, Before: slices.Clone(before)})
					if n > 0 {
						open = append(open, len(matches)-1)
					}
				}
				total++
			}

			if n > 0 {
				before = append(before, t)
				if len(before) > n {
					before = before[1:]
				}
			}
		}
		return nil
	})
	if err != nil {
		return Page[Match]{}, err
	}

	next := -1
	if offset+limit < total {
		next = offset + limit
	}
	return Page[Match]{Items: matches, Total: total, NextOffset: next}, nil
}
