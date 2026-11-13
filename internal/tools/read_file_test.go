package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dmorgam/BandicootFS/internal/corpus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestReadFileTool(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "root")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.md"), []byte("alpha\nbeta\ngamma\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "secret.md"), []byte("s3cret"), 0o644); err != nil {
		t.Fatal(err)
	}
	cs := connect(t, root)

	call := func(args map[string]any) *mcp.CallToolResult {
		t.Helper()
		res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "read_file", Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	res := call(map[string]any{"path": "a.md", "start_line": 2})
	if res.IsError {
		t.Fatalf("tool error: %+v", res.Content)
	}
	var s corpus.Slice
	if err := json.Unmarshal([]byte(res.Content[0].(*mcp.TextContent).Text), &s); err != nil {
		t.Fatal(err)
	}
	if s.Text != "2\tbeta\n3\tgamma" || s.TotalLines != 3 || s.NextLine != -1 {
		t.Errorf("got %+v", s)
	}

	if res := call(map[string]any{"path": "../secret.md"}); !res.IsError {
		t.Errorf("read outside root: want error, got %+v", res.Content)
	}
}
