package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/IssierHM/CraftShelf/internal/handler/dto"
)

// 捕获 panic，记录堆栈，返回 500 响应
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					"panic", r,
					"path", c.FullPath(),
					"stack", string(debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Response{
					Error: &dto.Error{Code: "internal", Message: "服务内部错误"},
				})
			}
		}()
		c.Next()
	}
}
