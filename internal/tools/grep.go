package tools

import (
	"context"

	"github.com/dmorgam/BandicootFS/internal/corpus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type grepInput struct {
	Pattern    string `json:"pattern" jsonschema:"regular expression to search for"`
	BaseFolder string `json:"base_folder,omitempty" jsonschema:"sub-folder relative to the server root; defaults to root"`
	IgnoreCase bool   `json:"ignore_case,omitempty" jsonschema:"case-insensitive match"`
	Limit      int    `json:"limit,omitempty" jsonschema:"max matches to return (default 50, max 200)"`
	Offset     int    `json:"offset,omitempty" jsonschema:"matches to skip, for pagination"`
}

func (h *handler) registerGrep(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "grep",
		Description: "Search files under the base folder for a regular expression. Paginated.",
	}, h.grep)
}

func (h *handler) grep(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	in grepInput,
) (*mcp.CallToolResult, corpus.Page[corpus.Match], error) {
	pg, err := h.corpus.Grep(ctx, in.BaseFolder, in.Pattern, in.IgnoreCase, in.Offset, in.Limit)
	if err != nil {
		return nil, corpus.Page[corpus.Match]{}, err
	}
	return nil, pg, nil
}
