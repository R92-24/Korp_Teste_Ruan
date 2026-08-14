package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"korp/assistente/internal/apperror"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic recuperado", "error", recovered, "path", c.Request.URL.Path)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": apperror.Internal("erro interno inesperado"),
		})
	})
}

func respondError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr})
		return
	}
	slog.Error("erro não mapeado", "error", err)
	internal := apperror.Internal("erro interno inesperado")
	c.JSON(internal.HTTPStatus, gin.H{"error": internal})
}
