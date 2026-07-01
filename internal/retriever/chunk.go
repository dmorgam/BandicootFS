package retriever

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// collect walks baseFolder and splits every file with an allowed extension
// into chunks. Unreadable files are skipped; a bad baseFolder is an error.
func (l *Local) collect(ctx context.Context, baseFolder string) ([]Chunk, error) {
	var chunks []Chunk

	err := filepath.WalkDir(baseFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() || !l.allowed(path) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}
		chunks = append(chunks, chunkFile(path, string(data), l.ChunkLines, l.Overlap)...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return chunks, nil
}

// allowed reports whether path has one of l.Exts (case-insensitive).
func (l *Local) allowed(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range l.Exts {
		if ext == e {
			return true
		}
	}
	return false
}

// chunkFile splits content into overlapping windows of lines.
func chunkFile(source, content string, size, overlap int) []Chunk {
	lines := strings.Split(content, "\n")
	if size <= 0 {
		size = len(lines)
	}
	step := size - overlap
	if step < 1 {
		step = 1
	}

	var chunks []Chunk
	for start := 0; start < len(lines); start += step {
		end := start + size
		if end > len(lines) {
			end = len(lines)
		}
		text := strings.Join(lines[start:end], "\n")
		if strings.TrimSpace(text) != "" {
			chunks = append(chunks, Chunk{
				Text:      text,
				Source:    source,
				StartLine: start + 1, // 1-indexed
			})
		}
		if end == len(lines) {
			break // last window reached the end
		}
	}
	return chunks
}
