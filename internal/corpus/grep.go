package corpus

import (
	"bufio"
	"context"
	"os"
	"regexp"
)

// Match is one line matching a grep pattern.
type Match struct {
	Source string `json:"source"`
	Line   int    `json:"line"`
	Text   string `json:"text"`
}

// Grep searches files under sub (root-relative) for pattern, paginated.
func (c *Corpus) Grep(ctx context.Context, sub, pattern string, ignoreCase bool, offset, limit int) (Page[Match], error) {
	dir, err := c.resolve(sub)
	if err != nil {
		return Page[Match]{}, err
	}
	if pattern == "" {
		return Page[Match]{Items: []Match{}}, nil
	}
	if ignoreCase {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return Page[Match]{}, err
	}

	var matches []Match
	err = c.walk(ctx, dir, func(path string) error {
		f, err := os.Open(path)
		if err != nil {
			return nil // skip
		}
		defer f.Close()

		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024) // tolerate long lines
		line := 0
		for sc.Scan() {
			line++
			if t := sc.Text(); re.MatchString(t) {
				matches = append(matches, Match{Source: c.rel(path), Line: line, Text: truncate(t, maxLineLen)})
			}
		}
		return nil
	})
	if err != nil {
		return Page[Match]{}, err
	}

	return page(matches, offset, clampLimit(limit)), nil
}
