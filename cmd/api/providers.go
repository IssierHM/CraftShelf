package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/IssierHM/CraftShelf/internal/pkg/config"
	"github.com/IssierHM/CraftShelf/internal/pkg/db"
	jwtpkg "github.com/IssierHM/CraftShelf/internal/pkg/jwt"
	"github.com/IssierHM/CraftShelf/internal/service"
)

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func newPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, func(), error) {
	pool, err := db.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, nil, err
	}
	return pool, func() { pool.Close() }, nil
}

func newSigner(cfg *config.Config) (*jwtpkg.Signer, error) {
	return jwtpkg.NewSigner(cfg.JWTSecret, cfg.JWTTTL)
}

func newAuthService(users service.UserRepository, tokens service.TokenSigner, cfg *config.Config) *service.AuthService {
	return service.NewAuthService(users, tokens, cfg.BcryptCost)
}

func newHTTPServer(cfg *config.Config, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
