package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"korp/faturamento/internal/nota"
)

func NewRouter(notaService *nota.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), Recovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept"},
		MaxAge:          12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	h := NewNotaHandler(notaService)
	notas := r.Group("/notas")
	{
		notas.POST("", h.Create)
		notas.GET("", h.List)
		notas.GET("/:numero", h.Get)
		notas.POST("/:numero/itens", h.AddItem)
		notas.DELETE("/:numero/itens/:itemId", h.RemoveItem)
		notas.POST("/:numero/imprimir", h.Imprimir)
	}

	return r
}
