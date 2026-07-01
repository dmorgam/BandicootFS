// Package retriever is the domain layer: it retrieves the chunks most related
// to a prompt within a base folder (the RAG core).
//
// It must NOT import the mcp package. Tool shells (package tools) translate
// MCP input/output to and from the types here, so this package stays testable
// and reusable on its own.
//
// Ranking is *lexical* — no embedding model, no external service. Chunks under
// the base folder are tokenized (stopwords removed) and scored against the
// prompt with BM25, rebuilt on every call. To go semantic later, swap the
// tokenizer/BM25 for a dense embedder + cosine.
package retriever

import (
	"context"
	"sort"
)

// Chunk is a piece of a file plus its retrieval score.
type Chunk struct {
	Text      string  `json:"text"`
	Source    string  `json:"source"`
	StartLine int     `json:"start_line"`
	Score     float32 `json:"score"`
}

// Retriever returns the chunks most related to a prompt within a base folder.
type Retriever interface {
	Chunks(ctx context.Context, baseFolder, prompt string) ([]Chunk, error)
}

// Local is a Retriever over the host filesystem.
type Local struct {
	Exts       []string // file extensions to include (with dot, lowercase)
	ChunkLines int      // lines per chunk
	Overlap    int      // overlapping lines between consecutive chunks
	TopK       int      // max chunks returned
	K1         float64  // BM25 term-frequency saturation
	B          float64  // BM25 length normalization (0..1)
	MinScore   float32  // drop chunks scoring at or below this
}

// Option configures a Local.
type Option func(*Local)

// WithTopK sets the default max number of chunks returned (ignored if n <= 0).
func WithTopK(n int) Option {
	return func(l *Local) {
		if n > 0 {
			l.TopK = n
		}
	}
}

// WithChunkLines sets the lines per chunk (ignored if n <= 0).
func WithChunkLines(n int) Option {
	return func(l *Local) {
		if n > 0 {
			l.ChunkLines = n
		}
	}
}

// WithOverlap sets the overlapping lines between chunks (ignored if n < 0).
func WithOverlap(n int) Option {
	return func(l *Local) {
		if n >= 0 {
			l.Overlap = n
		}
	}
}

// WithK1 sets the BM25 term-frequency saturation (ignored if f < 0).
func WithK1(f float64) Option {
	return func(l *Local) {
		if f >= 0 {
			l.K1 = f
		}
	}
}

// WithB sets the BM25 length normalization (ignored unless 0 <= f <= 1).
func WithB(f float64) Option {
	return func(l *Local) {
		if f >= 0 && f <= 1 {
			l.B = f
		}
	}
}

// WithMinScore sets the score at or below which chunks are dropped (ignored if f < 0).
func WithMinScore(f float32) Option {
	return func(l *Local) {
		if f >= 0 {
			l.MinScore = f
		}
	}
}

// New returns a Local with sensible defaults, overridden by opts.
func New(opts ...Option) *Local {
	l := &Local{
		Exts:       []string{".md", ".txt", ".go", ".py", ".js", ".ts", ".json", ".yaml", ".yml"},
		ChunkLines: 80,
		Overlap:    20,
		TopK:       5,
		K1:         1.2,
		B:          0.75,
		MinScore:   0,
	}
	for _, o := range opts {
		o(l)
	}
	return l
}

// Chunks walks baseFolder, splits every matching file into chunks, then ranks
// them against prompt with BM25.
func (l *Local) Chunks(ctx context.Context, baseFolder, prompt string) ([]Chunk, error) {
	chunks, err := l.collect(ctx, baseFolder)
	if err != nil {
		return nil, err
	}
	if len(chunks) == 0 {
		return nil, nil
	}

	docTokens := make([][]string, len(chunks))
	for i, c := range chunks {
		docTokens[i] = tokenize(c.Text)
	}
	query := uniq(tokenize(prompt))
	if len(query) == 0 {
		return nil, nil // prompt has no usable terms
	}

	ix := newBM25(docTokens, l.K1, l.B)

	var scored []Chunk
	for i := range chunks {
		s := float32(ix.score(i, query))
		if s > l.MinScore {
			chunks[i].Score = s
			scored = append(scored, chunks[i])
		}
	}

	sort.Slice(scored, func(i, j int) bool { return scored[i].Score > scored[j].Score })
	if len(scored) > l.TopK {
		scored = scored[:l.TopK]
	}
	return scored, nil
}
