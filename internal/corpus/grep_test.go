package corpus

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGrepContext(t *testing.T) {
	c := sandbox(t)
	ctx := context.Background()
	body := "l1\nhit one\nl3\nl4\nhit two\nl6\n"
	if err := os.WriteFile(filepath.Join(c.Root, "g.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	pg, err := c.Grep(ctx, "", "^hit", GrepOptions{Context: 2})
	want := []Match{
		{Source: "g.md", Line: 2, Text: "hit one", Before: []string{"l1"}, After: []string{"l3", "l4"}},
		{Source: "g.md", Line: 5, Text: "hit two", Before: []string{"l3", "l4"}, After: []string{"l6"}},
	}
	if err != nil || !reflect.DeepEqual(pg.Items, want) {
		t.Errorf("context 2 = %+v, %v; want %+v", pg.Items, err, want)
	}

	// Only the requested page is kept, but every match is counted.
	pg, err = c.Grep(ctx, "", "^hit", GrepOptions{Offset: 1, Limit: 1})
	if err != nil || pg.Total != 2 || pg.NextOffset != -1 || len(pg.Items) != 1 ||
		pg.Items[0].Line != 5 || pg.Items[0].Before != nil {
		t.Errorf("page 2 = %+v, %v; want only line 5, no context", pg, err)
	}
	pg, err = c.Grep(ctx, "", "^hit", GrepOptions{Limit: 1})
	if err != nil || pg.Total != 2 || pg.NextOffset != 1 || len(pg.Items) != 1 {
		t.Errorf("page 1 = %+v, %v; want 1 item, next_offset 1", pg, err)
	}
}

func TestTruncateKeepsRunes(t *testing.T) {
	if got := truncate("añb", 2); got != "a…" { // "ñ" is 2 bytes: cut before it
		t.Errorf("truncate = %q, want %q", got, "a…")
	}
}
