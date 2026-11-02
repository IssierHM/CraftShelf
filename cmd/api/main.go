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

	"github.com/IssierHM/CraftShelf/internal/handler"
	"github.com/IssierHM/CraftShelf/internal/pkg/config"
	"github.com/IssierHM/CraftShelf/internal/pkg/db"
	jwtpkg "github.com/IssierHM/CraftShelf/internal/pkg/jwt"
	"github.com/IssierHM/CraftShelf/internal/repository"
	"github.com/IssierHM/CraftShelf/internal/service"
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

	// Logger
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(log)

	// Context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Database
	pool, err := db.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		slog.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Dependencies
	userRepo := repository.NewUserRepo(pool)
	signer, err := jwtpkg.NewSigner(cfg.JWTSecret, cfg.JWTTTL)
	if err != nil {
		log.Error("init jwt signer", "error", err)
		os.Exit(1)
	}
	authSvc := service.NewAuthService(userRepo, signer, cfg.BcryptCost)
	authHandler := handler.NewAuthHandler(authSvc, log)
	userHandler := handler.NewUserHandler(authSvc, log)
	healthHandler := handler.NewHealthHandler(pool)

	//Router
	router := handler.NewRouter(handler.RouterDeps{
		Auth:   authHandler,
		User:   userHandler,
		Health: healthHandler,
		Tokens: signer,
		Log:    log,
	})

	// HTTP Service
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start server
	go func() {
		log.Info("starting HTTP server", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server", "error", err)
			stop()
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	log.Info("shutting down HTTP server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown", "error", err)
	}
	log.Info("bye")
}
