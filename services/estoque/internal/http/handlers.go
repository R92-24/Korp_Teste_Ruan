package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"korp/estoque/internal/produto"
)

type ProdutoHandler struct {
	service *produto.Service
}

func NewProdutoHandler(service *produto.Service) *ProdutoHandler {
	return &ProdutoHandler{service: service}
}

func (h *ProdutoHandler) Create(c *gin.Context) {
	var in produto.CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, badRequest(err))
		return
	}
	p, err := h.service.Create(c.Request.Context(), in)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *ProdutoHandler) List(c *gin.Context) {
	produtos, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, produtos)
}

func (h *ProdutoHandler) Get(c *gin.Context) {
	p, err := h.service.GetByCodigo(c.Request.Context(), c.Param("codigo"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProdutoHandler) Update(c *gin.Context) {
	var in produto.UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, badRequest(err))
		return
	}
	p, err := h.service.Update(c.Request.Context(), c.Param("codigo"), in)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProdutoHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("codigo")); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProdutoHandler) Baixa(c *gin.Context) {
	var in produto.MovimentoInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, badRequest(err))
		return
	}
	p, err := h.service.Baixa(c.Request.Context(), c.Param("codigo"), in)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProdutoHandler) Estorno(c *gin.Context) {
	var in produto.MovimentoInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, badRequest(err))
		return
	}
	p, err := h.service.Estorno(c.Request.Context(), c.Param("codigo"), in)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}
