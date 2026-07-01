package corpus

import (
	"context"
	"os"
	"path/filepath"
	"sort"
)

// File is a listed file, path relative to Root.
type File struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// List returns the indexable files under sub (root-relative), paginated.
func (c *Corpus) List(ctx context.Context, sub, glob string, offset, limit int) (Page[File], error) {
	dir, err := c.resolve(sub)
	if err != nil {
		return Page[File]{}, err
	}

	var files []File
	err = c.walk(ctx, dir, func(path string) error {
		if glob != "" {
			if ok, _ := filepath.Match(glob, filepath.Base(path)); !ok {
				return nil
			}
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil // skip
		}
		files = append(files, File{Path: c.rel(path), Size: info.Size()})
		return nil
	})
	if err != nil {
		return Page[File]{}, err
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return page(files, offset, clampLimit(limit)), nil
}
