// Command server runs the fizz-buzz HTTP API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/HelmyBc/fizzbuzz-server/internal/server"
	"github.com/HelmyBc/fizzbuzz-server/internal/stats"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Validate the port before attempting to bind, so operators get a clear
	// error message instead of a cryptic "bind: invalid argument".
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		slog.Error("invalid PORT value, must be an integer between 1 and 65535", "port", port)
		os.Exit(1)
	}

	store := stats.New()
	router := server.NewRouter(store)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
		// ReadHeaderTimeout defends against Slowloris attacks where a client
		// holds a connection open by sending headers very slowly.
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run the server in a goroutine so the main goroutine can wait for a
	// shutdown signal without blocking.
	go func() {
		slog.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for SIGINT/SIGTERM (e.g. `docker stop`, Kubernetes pod eviction)
	// and shut down cleanly instead of dropping in-flight connections.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
