//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/IssierHM/CraftShelf/internal/handler"
	"github.com/IssierHM/CraftShelf/internal/middleware"
	"github.com/IssierHM/CraftShelf/internal/pkg/config"
	jwtpkg "github.com/IssierHM/CraftShelf/internal/pkg/jwt"
	"github.com/IssierHM/CraftShelf/internal/repository"
	"github.com/IssierHM/CraftShelf/internal/service"
)

func InitializeApp(ctx context.Context, cfg *config.Config) (*App, func(), error) {
	wire.Build(
		// Infrastructure
		newLogger,
		newPool,

		// Repositories
		repository.NewUserRepo,
		wire.Bind(new(service.UserRepository), new(*repository.UserRepo)),

		// Signers
		newSigner,
		wire.Bind(new(service.TokenSigner), new(*jwtpkg.Signer)),
		wire.Bind(new(middleware.TokenParser), new(*jwtpkg.Signer)),

		// Services
		newAuthService,

		// Handlers
		handler.NewAuthHandler,
		handler.NewUserHandler,
		handler.NewHealthHandler,
		wire.Bind(new(handler.Pinger), new(*pgxpool.Pool)),

		// Router
		wire.Struct(new(handler.RouterDeps), "*"),
		handler.NewRouter,

		// App
		newHTTPServer,
		wire.Struct(new(App), "*"),
	)
	return nil, nil, nil
}
