package http

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"korp/assistente/internal/apperror"
	"korp/assistente/internal/conferencia"
)

func NewRouter(service *conferencia.Service, iaDisponivel bool) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), Recovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept"},
		MaxAge:          12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "iaDisponivel": iaDisponivel})
	})

	r.POST("/conferencia", func(c *gin.Context) {
		var req conferencia.Request
		if err := c.ShouldBindJSON(&req); err != nil {
			respondError(c, apperror.Validation("payload inválido: "+err.Error()))
			return
		}

		resultado, err := service.Conferir(c.Request.Context(), req)
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(http.StatusOK, resultado)
	})

	return r
}
