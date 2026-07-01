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
	"path/filepath"
	"syscall"

	"github.com/dmorgam/BandicootFS/internal/retriever"
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

		root = flag.String(
			"root",
			".",
			"root directory the server is allowed to access",
			)

		topK = flag.Int(
			"topk",
			5,
			"default number of chunks returned per query",
			)

		chunkLines = flag.Int(
			"chunk-lines",
			80,
			"lines per chunk",
			)

		overlap = flag.Int(
			"overlap",
			20,
			"overlapping lines between chunks",
			)

		minScore = flag.Float64(
			"min-score",
			0,
			"drop chunks scoring at or below this",
			)

		k1 = flag.Float64(
			"bm25-k1",
			1.2,
			"BM25 term-frequency saturation",
			)

		b = flag.Float64(
			"bm25-b",
			0.75,
			"BM25 length normalization (0..1)",
			)
	)

	flag.Parse()

	// Resolve root to an absolute path (what we serve and log).
	absRoot, err := filepath.Abs(*root)
	if err != nil {
		absRoot = *root
	}

	// stdio speaks JSON-RPC over stdout, so its logs must go to stderr.
	logOut := os.Stdout
	if *transport == "stdio" {
		logOut = os.Stderr
	}
	logger := slog.New(slog.NewTextHandler(logOut, nil))

	// Context for clean stop
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		)

	defer stop()

	// MCP server (transport-agnostic)
	srv := server.New(logger, retriever.New(
		retriever.WithTopK(*topK),
		retriever.WithChunkLines(*chunkLines),
		retriever.WithOverlap(*overlap),
		retriever.WithK1(*k1),
		retriever.WithB(*b),
		retriever.WithMinScore(float32(*minScore)),
	), absRoot)

	switch *transport {

	case "stdio":

		printBanner(logger)

		logger.Info(
			"Selected transport",
			"name", *transport,
			"root", absRoot)

		transportpkg.RunStdio(ctx, logger, srv)

	case "http":

		printBanner(logger)

		logger.Info(
			"Selected transport",
			"name", *transport,
			"port", *port,
			"path", *mountPath,
			"root", absRoot)

		transportpkg.RunHTTP(ctx, logger, srv, *port, *mountPath)

	default:
		logger.Error(
			"Unknown transport",
			"name",
			*transport)

	}
}
