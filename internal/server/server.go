// Package server defines the mcp server, tools etc...
package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	ServerName    = "BandicootFS"
	ServerVersion = "0.1.0"
)

// New builds a configured *mcp.Server.
func New() *mcp.Server {
	srv := mcp.NewServer(
		&mcp.Implementation{
			Name:    ServerName,
			Version: ServerVersion,
		},
		&mcp.ServerOptions{},
	)

	// TODO: Register tools

	return srv
}
