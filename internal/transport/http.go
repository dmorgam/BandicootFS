package transport

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RunHTTP(
	ctx context.Context,
	logger *slog.Logger,
	srv *mcp.Server,
	port int,
	path string,
) {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/healthz",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		},
	)

	streamHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return srv },
		nil,
	)
	mux.Handle(path, streamHandler)

	httpServer := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info(
			"mcp server listening",
			"port", port,
			"path", path,
		)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server failed", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(), 10*time.Second,
	)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	}
}
