package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/IssierHM/CraftShelf/internal/handler/dto"
)

func respondData(c *gin.Context, status int, data any) {
	c.JSON(status, dto.Response{Data: data})
}

func respondError(c *gin.Context, log *slog.Logger, status int, code, msg string, cause error) {
	if status >= http.StatusInternalServerError && cause != nil {
		log.Error(msg, "code", code, "cause", cause, "path", c.FullPath())
	}
	c.AbortWithStatusJSON(status, dto.Response{
		Error: &dto.Error{Code: code, Message: msg},
	})
}
