package tools

import (
	"github.com/dmorgam/BandicootFS/internal/corpus"
	"github.com/dmorgam/BandicootFS/internal/retriever"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type handler struct {
	ret     retriever.Retriever
	corpus  *corpus.Corpus
	baseDir string
}

func Register(srv *mcp.Server, r retriever.Retriever, baseDir string) {
	h := &handler{ret: r, corpus: corpus.New(baseDir), baseDir: baseDir}

	h.registerGetFileChunks(srv)
	h.registerGrep(srv)
	h.registerListFiles(srv)
}
