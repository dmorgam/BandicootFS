// Package corpus provides filesystem browsing/search over the server root:
// listing and grepping indexable files. All access goes through an os.Root,
// so nothing outside Root is reachable, not even through symlinks.
//
// It must NOT import the mcp package.
package corpus

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Corpus lists and searches indexable files under Root.
type Corpus struct {
	Root string   // absolute, cleaned sandbox boundary
	Exts []string // file extensions to include (with dot, lowercase)

	root *os.Root // every read goes through this handle
}

// Open returns a Corpus confined to root (resolved to an absolute path).
// It fails if root cannot be opened as a directory. Close releases it.
func Open(root string) (*Corpus, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	r, err := os.OpenRoot(abs)
	if err != nil {
		return nil, err
	}
	return &Corpus{
		Root: abs,
		Exts: []string{".md", ".txt", ".go", ".py", ".js", ".ts", ".json", ".yaml", ".yml"},
		root: r,
	}, nil
}

// Close releases the root handle.
func (c *Corpus) Close() error { return c.root.Close() }

// FS returns the sandboxed file system. Paths are slash-separated and relative
// to Root; opening anything outside Root fails, including through symlinks.
func (c *Corpus) FS() fs.FS { return c.root.FS() }

// Resolve maps a caller-supplied folder to a path in FS. It accepts paths
// relative to Root ("" means Root) and absolute paths under Root; anything
// pointing outside Root is rejected.
func (c *Corpus) Resolve(p string) (string, error) {
	rel := p
	if filepath.IsAbs(p) {
		r, err := filepath.Rel(c.Root, p)
		if err != nil {
			return "", fmt.Errorf("path %q escapes root", p)
		}
		rel = r
	}
	rel = path.Clean(filepath.ToSlash(rel))
	if !fs.ValidPath(rel) { // ".." or "../x" survive cleaning: outside root
		return "", fmt.Errorf("path %q escapes root", p)
	}
	return rel, nil
}

// allowed reports whether name has one of c.Exts (case-insensitive).
func (c *Corpus) allowed(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, e := range c.Exts {
		if ext == e {
			return true
		}
	}
	return false
}

// walk visits every allowed file under dir (a path in FS), calling fn with its
// root-relative path.
func (c *Corpus) walk(ctx context.Context, dir string, fn func(p string) error) error {
	return fs.WalkDir(c.FS(), dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() || !c.allowed(p) {
			return nil
		}
		return fn(p)
	})
}
