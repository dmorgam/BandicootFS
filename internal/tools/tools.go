package tools

import (
	"github.com/dmorgam/BandicootFS/internal/corpus"
	"github.com/dmorgam/BandicootFS/internal/retriever"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type handler struct {
	ret    retriever.Retriever
	corpus *corpus.Corpus
}

// Register adds the tools to srv. Every tool reads through c, so none can reach
// outside c.Root.
func Register(srv *mcp.Server, r retriever.Retriever, c *corpus.Corpus) {
	h := &handler{ret: r, corpus: c}

	h.registerGetFileChunks(srv)
	h.registerGrep(srv)
	h.registerListFiles(srv)
}
