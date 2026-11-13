package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmorgam/BandicootFS/internal/corpus"
	"github.com/dmorgam/BandicootFS/internal/retriever"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connect serves the tools, sandboxed to root, over an in-memory transport.
func connect(t *testing.T, root string) *mcp.ClientSession {
	t.Helper()
	c, err := corpus.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })

	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	Register(srv, retriever.New(), c)

	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ss.Close() })

	cs, err := mcp.NewClient(&mcp.Implementation{Name: "c", Version: "0"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

func TestGetFileChunksTool(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("cosine similarity vector search"), 0o644); err != nil {
		t.Fatal(err)
	}
	cs := connect(t, dir)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_file_chunks",
		Arguments: map[string]any{"prompt": "vector search"}, // base_folder omitted -> root
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("tool error: %+v", res.Content)
	}

	var out struct {
		Chunks []retriever.Chunk `json:"chunks"`
	}
	if err := json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Chunks) == 0 {
		t.Fatal("no chunks")
	}
}

// get_file_chunks must not read outside root: neither through base_folder nor
// through a symlink inside root.
func TestGetFileChunksSandbox(t *testing.T) {
	base := t.TempDir()
	root, out := filepath.Join(base, "root"), filepath.Join(base, "out")
	for _, d := range []string{root, out} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(out, "secret.md"), []byte("oauth token s3cret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../out/secret.md", filepath.Join(root, "leak.md")); err != nil {
		t.Fatal(err)
	}
	cs := connect(t, root)

	call := func(args map[string]any) *mcp.CallToolResult {
		t.Helper()
		res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "get_file_chunks", Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	for _, folder := range []string{out, "../out", ".."} {
		if res := call(map[string]any{"prompt": "oauth token", "base_folder": folder}); !res.IsError {
			t.Errorf("base_folder %q: want error, got %+v", folder, res.Content)
		}
	}

	res := call(map[string]any{"prompt": "oauth token"})
	if txt := res.Content[0].(*mcp.TextContent).Text; res.IsError || strings.Contains(txt, "s3cret") {
		t.Errorf("symlink escaped root: %s", txt)
	}
}
