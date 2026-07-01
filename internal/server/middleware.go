package server

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// logToolCalls logs every tool invocation and its raw arguments (never results).
func logToolCalls(logger *slog.Logger) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			// Only tools/call carries CallToolParamsRaw; other methods skip.
			if p, ok := req.GetParams().(*mcp.CallToolParamsRaw); ok {
				logger.Info("tool call", "name", p.Name, "args", string(p.Arguments))
			}
			return next(ctx, method, req)
		}
	}
}
