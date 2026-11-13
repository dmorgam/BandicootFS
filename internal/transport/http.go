package transport

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPOptions configures RunHTTP.
type HTTPOptions struct {
	Host  string // listen address; a loopback one keeps the server local
	Port  int
	Path  string // mount path of the MCP handler
	Token string // if set, the MCP handler requires "Authorization: Bearer <Token>"
}

func RunHTTP(
	ctx context.Context,
	logger *slog.Logger,
	srv *mcp.Server,
	opts HTTPOptions,
) {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/healthz",
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		},
	)

	var streamHandler http.Handler = mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server { return srv },
		nil,
	)
	if opts.Token != "" {
		streamHandler = requireBearer(opts.Token, streamHandler)
	} else if !isLoopback(opts.Host) {
		logger.Warn(
			"no auth token on a non-loopback address: anyone reaching it can read files under root",
			"host", opts.Host,
		)
	}
	mux.Handle(opts.Path, streamHandler)

	httpServer := &http.Server{
		Addr:              net.JoinHostPort(opts.Host, strconv.Itoa(opts.Port)),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info(
			"mcp server listening",
			"addr", httpServer.Addr,
			"path", opts.Path,
			"auth", opts.Token != "",
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

// requireBearer rejects requests without "Authorization: Bearer <token>".
// Both sides are hashed so the comparison is constant-time and length-blind.
func requireBearer(token string, next http.Handler) http.Handler {
	want := sha256.Sum256([]byte(token))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scheme, got, _ := strings.Cut(r.Header.Get("Authorization"), " ")
		sum := sha256.Sum256([]byte(got))
		if !strings.EqualFold(scheme, "Bearer") || subtle.ConstantTimeCompare(sum[:], want[:]) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// isLoopback reports whether host only accepts local connections.
func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
