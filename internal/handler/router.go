package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/IssierHM/CraftShelf/internal/middleware"
)

// RouterDeps 汇总 NewRouter 需要的全部依赖
type RouterDeps struct {
	Auth   *AuthHandler
	User   *UserHandler
	Health *HealthHandler
	Tokens middleware.TokenParser
	Log    *slog.Logger
}

func NewRouter(d RouterDeps) *gin.Engine {
	r := gin.New()

	r.Use(middleware.Recovery(d.Log), gin.Logger())

	r.GET("/healthz", d.Health.Live)
	r.GET("/readyz", d.Health.Ready)

	v1 := r.Group("/api/v1")
	{
		// Public endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/register", d.Auth.Register)
			auth.POST("/login", d.Auth.Login)
		}

		// Authenticated endpoints
		users := v1.Group("/users")
		users.Use(middleware.Auth(d.Tokens, d.Log))
		{
			users.GET("/me", d.User.Me)
		}
	}

	return r
}
