// BandicootFS MCP
//
// It supports two transports:
//
//	stdio  - one client per process, communicates over stdin/stdout.
//	         Use this for local clients like Claude Desktop or the MCP Inspector.
//	http   - long-running HTTP server using the streamable transport, accepts
//	         many concurrent clients. Mounts on /mcp.

package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmorgam/BandicootFS/internal/server"
	transportpkg "github.com/dmorgam/BandicootFS/internal/transport"
)

func main() {

	// Parameters
	var (
		transport = flag.String(
			"transport",
			"stdio",
			"transport to use: stdio | http",
			)

		port = flag.Int(
			"port",
			8000,
			"port to listen on (http transport only)",
			)

		mountPath = flag.String(
			"path",
			"/mcp",
			"mount path for the streamable HTTP handler",
			)
	)

	flag.Parse()

	// Logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Context for clean stop
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		)

	defer stop()

	// MCP server (transport-agnostic)
	srv := server.New()

	switch *transport {

	case "stdio":
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

		printBanner(logger)

		logger.Info(
			"Selected transport",
			"name",
			*transport)

		transportpkg.RunStdio(ctx, logger, srv)

	case "http":

		printBanner(logger)

		logger.Info(
			"Selected transport",
			"name", *transport,
			"port", *port,
			"path", *mountPath)

		transportpkg.RunHTTP(ctx, logger, srv, *port, *mountPath)

	default:
		logger.Error(
			"Unknown transport",
			"name",
			*transport)

	}
}
