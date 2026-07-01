package tools

import (
	"github.com/dmorgam/BandicootFS/internal/retriever"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type handler struct {
	ret     retriever.Retriever
	baseDir string
}

func Register(srv *mcp.Server, r retriever.Retriever, baseDir string) {
	h := &handler{ret: r, baseDir: baseDir}

	h.registerGetFileChunks(srv)
}
