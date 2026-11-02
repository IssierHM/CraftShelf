package handler

import (
	"errors"
	"net/http"

	"github.com/IssierHM/CraftShelf/internal/domain"
)

func mapDomainError(err error) (status int, code, msg string) {
	switch {
	case errors.Is(err, domain.ErrEmailExists):
		return http.StatusConflict, "email_exists", "该邮箱已被注册"
	case errors.Is(err, domain.ErrUsernameExists):
		return http.StatusConflict, "username_exists", "该用户名已被占用"
	case errors.Is(err, domain.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid_credentials", "邮箱或密码错误"
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, "user_not_found", "用户不存在"
	default:
		return http.StatusInternalServerError, "internal", "服务内部错误"
	}
}
