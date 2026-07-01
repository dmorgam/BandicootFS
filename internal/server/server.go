// Package server defines the mcp server, tools etc...
package server

import (
	"log/slog"

	"github.com/dmorgam/BandicootFS/internal/retriever"
	"github.com/dmorgam/BandicootFS/internal/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	ServerName    = "BandicootFS"
	ServerVersion = "0.1.0"
)

// New builds a configured *mcp.Server backed by r. baseDir is the default base
// folder tools use when a caller omits one. logger records every tool call.
func New(logger *slog.Logger, r retriever.Retriever, baseDir string) *mcp.Server {
	srv := mcp.NewServer(
		&mcp.Implementation{
			Name:    ServerName,
			Version: ServerVersion,
		},
		&mcp.ServerOptions{},
	)

	srv.AddReceivingMiddleware(logToolCalls(logger))
	tools.Register(srv, r, baseDir)

	return srv
}
