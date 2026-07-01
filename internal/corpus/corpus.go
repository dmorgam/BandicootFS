// Package corpus provides filesystem browsing/search over the server root:
// listing and grepping indexable files. All access is sandboxed to Root.
//
// It must NOT import the mcp package.
package corpus

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Corpus lists and searches indexable files under Root.
type Corpus struct {
	Root string   // absolute, cleaned sandbox boundary
	Exts []string // file extensions to include (with dot, lowercase)
}

// New returns a Corpus rooted at root (resolved to an absolute path).
func New(root string) *Corpus {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	return &Corpus{
		Root: filepath.Clean(abs),
		Exts: []string{".md", ".txt", ".go", ".py", ".js", ".ts", ".json", ".yaml", ".yml"},
	}
}

// resolve maps a root-relative sub path to an absolute path, rejecting anything
// that escapes Root.
func (c *Corpus) resolve(sub string) (string, error) {
	abs := filepath.Clean(filepath.Join(c.Root, sub))
	if abs != c.Root && !strings.HasPrefix(abs, c.Root+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q escapes root", sub)
	}
	return abs, nil
}

// rel returns path relative to Root (for display).
func (c *Corpus) rel(path string) string {
	if r, err := filepath.Rel(c.Root, path); err == nil {
		return r
	}
	return path
}

// allowed reports whether path has one of c.Exts (case-insensitive).
func (c *Corpus) allowed(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range c.Exts {
		if ext == e {
			return true
		}
	}
	return false
}

// walk visits every allowed file under dir, calling fn with its path.
func (c *Corpus) walk(ctx context.Context, dir string, fn func(path string) error) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() || !c.allowed(path) {
			return nil
		}
		return fn(path)
	})
}
