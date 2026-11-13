package corpus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	c := sandbox(t)
	ctx := context.Background()
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(c.Root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var b strings.Builder
	for i := 1; i <= 1200; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	write("big.md", b.String())
	write("wide.md", strings.Repeat(strings.Repeat("x", 1000)+"\n", 100)) // ~100 KB
	write("bin.txt", "abc\x00def")
	write(".env", "TOKEN=x")

	meta := func(s Slice) string {
		return fmt.Sprintf("start=%d end=%d next=%d total=%d", s.StartLine, s.EndLine, s.NextLine, s.TotalLines)
	}
	for _, tc := range []struct {
		start, end int
		want       string
	}{
		{0, 0, "start=1 end=500 next=501 total=1200"},          // default page
		{1190, 5000, "start=1190 end=1200 next=-1 total=1200"}, // last page
		{10, 12, "start=10 end=12 next=13 total=1200"},         // explicit range
	} {
		s, err := c.Read(ctx, "big.md", tc.start, tc.end)
		if err != nil || meta(s) != tc.want {
			t.Errorf("Read(big.md, %d, %d) = %s, %v; want %s", tc.start, tc.end, meta(s), err, tc.want)
		}
	}
	if s, _ := c.Read(ctx, "big.md", 10, 12); s.Text != "10\tline 10\n11\tline 11\n12\tline 12" {
		t.Errorf("text = %q", s.Text)
	}

	s, err := c.Read(ctx, "wide.md", 1, 100)
	if err != nil || s.EndLine >= 100 || s.NextLine != s.EndLine+1 || len(s.Text) > maxReadBytes {
		t.Errorf("byte cap: %s, len=%d, %v", meta(s), len(s.Text), err)
	}

	for _, p := range []string{".", "..", "../out/secret.yml", "rel.yml", "abs.yml", "link", "bin.txt", ".env", "nope.md"} {
		if _, err := c.Read(ctx, p, 0, 0); err == nil {
			t.Errorf("Read(%q): want error", p)
		}
	}
	if _, err := c.Read(ctx, "big.md", 1201, 0); err == nil {
		t.Error("start past the end: want error")
	}
}
