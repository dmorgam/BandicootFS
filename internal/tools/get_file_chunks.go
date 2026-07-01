package tools

import (
	"context"

	"github.com/dmorgam/BandicootFS/internal/retriever"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getFileChunksInput struct {
	BaseFolder string `json:"base_folder,omitempty" jsonschema:"folder to search within; defaults to the server base directory"`
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
	baseFolder := in.BaseFolder
	if baseFolder == "" {
		baseFolder = h.baseDir // default to the base dir given to the MCP server
	}

	chunks, err := h.ret.Chunks(ctx, baseFolder, in.Prompt)
	if err != nil {
		return nil, getFileChunksOutput{}, err
	}

	return nil, getFileChunksOutput{Chunks: chunks}, nil
}
