package corpus

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
)

const (
	maxReadLines   = 500      // lines per Read page
	maxReadBytes   = 64 << 10 // text bytes per Read page
	maxReadLineLen = 2000     // longer lines are truncated
)

// Slice is a numbered line range of a file.
type Slice struct {
	Path       string `json:"path"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"` // last line included; 0 if none
	TotalLines int    `json:"total_lines"`
	NextLine   int    `json:"next_line"` // start_line of the next page; -1 at end of file
	Text       string `json:"text"`      // one "<line>\t<content>" row per line
}

// Read returns lines [start, end] (1-based, inclusive) of the indexable file p
// (root-relative), capped at maxReadLines lines and maxReadBytes bytes. An end
// before start means "as many lines as the caps allow".
func (c *Corpus) Read(ctx context.Context, p string, start, end int) (Slice, error) {
	name, err := c.Resolve(p)
	if err != nil {
		return Slice{}, err
	}
	f, err := c.FS().Open(name)
	if err != nil {
		return Slice{}, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return Slice{}, err
	}
	if info.IsDir() {
		return Slice{}, fmt.Errorf("%q is a directory; use list_files", p)
	}
	if !c.allowed(name) {
		return Slice{}, fmt.Errorf("%q is not an indexable file type", p)
	}

	start = max(start, 1)
	if end < start || end-start >= maxReadLines {
		end = start + maxReadLines - 1
	}

	br := bufio.NewReader(f)
	if head, _ := br.Peek(8 << 10); bytes.IndexByte(head, 0) >= 0 {
		return Slice{}, fmt.Errorf("%q looks like a binary file", p)
	}
	sc := bufio.NewScanner(br)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024) // tolerate long lines

	s := Slice{Path: name, StartLine: start, NextLine: -1}
	var b strings.Builder
	line := 0
	for sc.Scan() {
		line++
		if line%4096 == 0 && ctx.Err() != nil {
			return Slice{}, ctx.Err()
		}
		if line < start || s.NextLine != -1 {
			continue // before the range, or page already full: just count
		}
		row := strconv.Itoa(line) + "\t" + truncate(sc.Text(), maxReadLineLen) + "\n"
		if line > end || (b.Len() > 0 && b.Len()+len(row) > maxReadBytes) {
			s.NextLine = line
			continue
		}
		b.WriteString(row)
		s.EndLine = line
	}
	if err := sc.Err(); err != nil {
		return Slice{}, fmt.Errorf("read %q: %w", p, err)
	}
	if line > 0 && start > line {
		return Slice{}, fmt.Errorf("start_line %d is past the end of %q (%d lines)", start, p, line)
	}
	s.TotalLines = line
	s.Text = strings.TrimSuffix(b.String(), "\n")
	return s, nil
}
