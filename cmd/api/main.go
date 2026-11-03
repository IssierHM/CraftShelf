package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IssierHM/CraftShelf/internal/pkg/config"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		slog.Warn("load .env", "err", err)
	}

	// Configure
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	// Context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize app
	app, cleanup, err := InitializeApp(ctx, cfg)
	if err != nil {
		slog.Error("initialize app", "err", err)
		os.Exit(1)
	}
	defer cleanup()

	go func() {
		app.Log.Info("http server listening", "addr", app.Server.Addr)
		if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Log.Error("http server", "err", err)
			stop()
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	app.Log.Info("shutting down HTTP server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		app.Log.Error("HTTP server shutdown", "error", err)
	}
	app.Log.Info("bye")
}
