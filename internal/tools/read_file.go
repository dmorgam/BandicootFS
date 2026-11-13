package tools

import (
	"context"

	"github.com/dmorgam/BandicootFS/internal/corpus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type readFileInput struct {
	Path      string `json:"path" jsonschema:"file path relative to the server root, as returned by list_files, grep or get_file_chunks"`
	StartLine int    `json:"start_line,omitempty" jsonschema:"first line to return, 1-based (default 1)"`
	EndLine   int    `json:"end_line,omitempty" jsonschema:"last line to return, inclusive (default and max: start_line+499)"`
}

func (h *handler) registerReadFile(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "read_file",
		Description: "Read a line range of an indexable file under the root. Each text row is \"<line>\\t<content>\". Paginated: if next_line is not -1, pass it as start_line to continue.",
	}, h.readFile)
}

func (h *handler) readFile(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	in readFileInput,
) (*mcp.CallToolResult, corpus.Slice, error) {
	s, err := h.corpus.Read(ctx, in.Path, in.StartLine, in.EndLine)
	if err != nil {
		return nil, corpus.Slice{}, err
	}
	return nil, s, nil
}
