package tools

import (
	"context"

	"github.com/dmorgam/BandicootFS/internal/retriever"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getFileChunksInput struct {
	BaseFolder string `json:"base_folder,omitempty" jsonschema:"sub-folder relative to the server root; defaults to root"`
	Prompt     string `json:"prompt" jsonschema:"natural-language query to retrieve related chunks for"`
}

type getFileChunksOutput struct {
	Chunks []retriever.Chunk `json:"chunks"`
}

func (h *handler) registerGetFileChunks(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "get_file_chunks",
		Description: "Retrieve the chunks most related to a prompt within a base folder.",
	}, h.getFileChunks)
}

func (h *handler) getFileChunks(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	in getFileChunksInput,
) (*mcp.CallToolResult, getFileChunksOutput, error) {
	dir, err := h.corpus.Resolve(in.BaseFolder) // "" -> root; rejects paths outside root
	if err != nil {
		return nil, getFileChunksOutput{}, err
	}

	chunks, err := h.ret.Chunks(ctx, h.corpus.FS(), dir, in.Prompt)
	if err != nil {
		return nil, getFileChunksOutput{}, err
	}

	return nil, getFileChunksOutput{Chunks: chunks}, nil
}
