// Package transport implements MCP transports: stdio and HTTP streamable.
package transport

import (
	"context"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RunStdio(
	ctx context.Context,
	logger *slog.Logger,
	srv *mcp.Server,
) {
	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil {
		logger.Error("stdio server stopped", "err", err)
		os.Exit(1)
	}
}
