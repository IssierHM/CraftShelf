package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/IssierHM/CraftShelf/internal/domain"
	"github.com/IssierHM/CraftShelf/internal/handler/dto"
	"github.com/IssierHM/CraftShelf/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	log         *slog.Logger
}

func NewAuthHandler(authService *service.AuthService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         log,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, h.log, http.StatusBadRequest, "invalid_request", "请求参数无效", err)
		return
	}

	u, err := h.authService.Register(c.Request.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		status, code, msg := mapDomainError(err)
		respondError(c, h.log, status, code, msg, err)
		return
	}

	respondData(c, http.StatusCreated, dto.ToUserResponse(u))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, h.log, http.StatusBadRequest, "invalid_request", "请求参数无效", err)
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			respondError(c, h.log, http.StatusUnauthorized, "invalid_credentials", "邮箱或密码错误", err)
			return
		}

		status, code, msg := mapDomainError(err)
		respondError(c, h.log, status, code, msg, err)
		return
	}

	respondData(c, http.StatusOK, dto.LoginResponse{Token: token})
}
