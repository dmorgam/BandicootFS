package corpus

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// sandbox opens a Corpus on root/, next to an out/ dir it must never read:
//
//	root/a.md
//	root/rel.yml -> ../out/secret.yml
//	root/abs.yml -> /abs/path/out/secret.yml
//	root/link    -> ../out
func sandbox(t *testing.T) *Corpus {
	t.Helper()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	base := t.TempDir()
	root, out := filepath.Join(base, "root"), filepath.Join(base, "out")
	must(os.Mkdir(root, 0o755))
	must(os.Mkdir(out, 0o755))
	must(os.WriteFile(filepath.Join(root, "a.md"), []byte("hello"), 0o644))
	must(os.WriteFile(filepath.Join(out, "secret.yml"), []byte("token: s3cret"), 0o644))
	must(os.Symlink("../out/secret.yml", filepath.Join(root, "rel.yml")))
	must(os.Symlink(filepath.Join(out, "secret.yml"), filepath.Join(root, "abs.yml")))
	must(os.Symlink("../out", filepath.Join(root, "link")))

	c, err := Open(root)
	must(err)
	t.Cleanup(func() { c.Close() })
	return c
}

func TestResolve(t *testing.T) {
	c := sandbox(t)
	for in, want := range map[string]string{
		"":               ".",
		".":              ".",
		"docs":           "docs",
		"./docs/":        "docs",
		"docs/../docs":   "docs",
		c.Root:           ".",
		c.Root + "/docs": "docs",
	} {
		if got, err := c.Resolve(in); err != nil || got != want {
			t.Errorf("Resolve(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"..", "../out", "docs/../..", "/etc", c.Root + "/../out"} {
		if got, err := c.Resolve(in); err == nil {
			t.Errorf("Resolve(%q) = %q; want error", in, got)
		}
	}
}

func TestSymlinksCannotEscape(t *testing.T) {
	c := sandbox(t)
	ctx := context.Background()

	if pg, err := c.Grep(ctx, "", "s3cret", false, 0, 0); err != nil || pg.Total != 0 {
		t.Errorf("grep through file symlinks = %+v, %v; want no matches", pg, err)
	}
	if _, err := c.Grep(ctx, "link", "s3cret", false, 0, 0); err == nil {
		t.Error("grep through dir symlink: want error")
	}
	pg, err := c.List(ctx, "", "", 0, 0)
	if err != nil || len(pg.Items) != 1 || pg.Items[0].Path != "a.md" {
		t.Errorf("list = %+v, %v; want only a.md", pg, err)
	}
}
