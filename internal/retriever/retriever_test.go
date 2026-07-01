package retriever

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestChunksRanking(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("payments.md", "refund customer invoice billing statements")
	write("cooking.md", "bake bread flour yeast oven recipe")

	got, err := New().Chunks(context.Background(), dir, "refund customer invoice")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || filepath.Base(got[0].Source) != "payments.md" {
		t.Fatalf("top = %+v, want payments.md", got)
	}
}

func TestChunksEmptyPrompt(t *testing.T) {
	got, err := New().Chunks(context.Background(), t.TempDir(), "")
	if err != nil || got != nil {
		t.Fatalf("got %+v, %v; want nil, nil", got, err)
	}
}

func TestChunksBadFolder(t *testing.T) {
	if _, err := New().Chunks(context.Background(), "/no/such/dir", "x"); err == nil {
		t.Fatal("want error for missing folder")
	}
}
