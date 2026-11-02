package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/IssierHM/CraftShelf/internal/handler/dto"
)

const ctxUserIDKey = "auth.user_id"

type TokenParser interface {
	Parse(token string) (int64, error)
}

func Auth(tokens TokenParser, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			abortUnauthorized(c, "缺少访问令牌")
			return
		}

		raw := strings.TrimPrefix(header, "Bearer ")
		if raw == "" {
			abortUnauthorized(c, "缺少访问令牌")
			return
		}

		userID, err := tokens.Parse(raw)
		if err != nil {
			log.Info("invalid token", "path", c.FullPath(), "err", err)
			abortUnauthorized(c, "访问令牌无效或已过期")
			return
		}

		c.Set(ctxUserIDKey, userID)
		c.Next()
	}
}

// UserIDFrom 供 handler 读取当前用户
func UserIDFrom(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ctxUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

func abortUnauthorized(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
		Error: &dto.Error{Code: "unauthorized", Message: msg},
	})
}
