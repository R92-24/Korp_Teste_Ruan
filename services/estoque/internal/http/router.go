package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"korp/estoque/internal/apperror"
	"korp/estoque/internal/produto"
)

func NewRouter(produtoService *produto.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), Recovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept"},
		MaxAge:          12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	h := NewProdutoHandler(produtoService)
	produtos := r.Group("/produtos")
	{
		produtos.POST("", h.Create)
		produtos.GET("", h.List)
		produtos.GET("/:codigo", h.Get)
		produtos.PUT("/:codigo", h.Update)
		produtos.DELETE("/:codigo", h.Delete)
		produtos.POST("/:codigo/baixa", h.Baixa)
		produtos.POST("/:codigo/estorno", h.Estorno)
	}

	return r
}

func badRequest(err error) error {
	return apperror.Validation("payload inválido: " + err.Error())
}
