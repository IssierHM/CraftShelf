package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/IssierHM/CraftShelf/internal/handler/dto"
	"github.com/IssierHM/CraftShelf/internal/middleware"
	"github.com/IssierHM/CraftShelf/internal/service"
)

type UserHandler struct {
	authService *service.AuthService
	log         *slog.Logger
}

func NewUserHandler(authService *service.AuthService, log *slog.Logger) *UserHandler {
	return &UserHandler{
		authService: authService,
		log:         log,
	}
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		respondError(c, h.log, http.StatusUnauthorized, "unauthorized", "缺少有效的访问令牌", nil)
		return
	}

	u, err := h.authService.GetByID(c.Request.Context(), userID)
	if err != nil {
		status, code, msg := mapDomainError(err)
		respondError(c, h.log, status, code, msg, err)
		return
	}

	respondData(c, http.StatusOK, dto.ToUserResponse(u))
}
