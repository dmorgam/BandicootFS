package tools

import (
	"context"

	"github.com/dmorgam/BandicootFS/internal/corpus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listFilesInput struct {
	BaseFolder string `json:"base_folder,omitempty" jsonschema:"sub-folder relative to the server root; defaults to root"`
	Glob       string `json:"glob,omitempty" jsonschema:"optional filename glob filter, e.g. *.go"`
	Limit      int    `json:"limit,omitempty" jsonschema:"max files to return (default 50, max 200)"`
	Offset     int    `json:"offset,omitempty" jsonschema:"files to skip, for pagination"`
}

func (h *handler) registerListFiles(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "list_files",
		Description: "List indexable files under the base folder. Paginated.",
	}, h.listFiles)
}

func (h *handler) listFiles(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	in listFilesInput,
) (*mcp.CallToolResult, corpus.Page[corpus.File], error) {
	pg, err := h.corpus.List(ctx, in.BaseFolder, in.Glob, in.Offset, in.Limit)
	if err != nil {
		return nil, corpus.Page[corpus.File]{}, err
	}
	return nil, pg, nil
}
