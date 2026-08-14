package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"korp/faturamento/internal/apperror"
	"korp/faturamento/internal/nota"
)

type NotaHandler struct {
	service *nota.Service
}

func NewNotaHandler(service *nota.Service) *NotaHandler {
	return &NotaHandler{service: service}
}

func (h *NotaHandler) Create(c *gin.Context) {
	n, err := h.service.Create(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, n)
}

func (h *NotaHandler) List(c *gin.Context) {
	notas, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, notas)
}

func (h *NotaHandler) Get(c *gin.Context) {
	numero, err := parseNumero(c)
	if err != nil {
		respondError(c, err)
		return
	}
	n, err := h.service.GetByNumero(c.Request.Context(), numero)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, n)
}

func (h *NotaHandler) AddItem(c *gin.Context) {
	numero, err := parseNumero(c)
	if err != nil {
		respondError(c, err)
		return
	}
	var in nota.AddItemInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, apperror.Validation("payload inválido: "+err.Error()))
		return
	}
	item, err := h.service.AddItem(c.Request.Context(), numero, in)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *NotaHandler) RemoveItem(c *gin.Context) {
	numero, err := parseNumero(c)
	if err != nil {
		respondError(c, err)
		return
	}
	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		respondError(c, apperror.Validation("id de item inválido"))
		return
	}
	if err := h.service.RemoveItem(c.Request.Context(), numero, itemID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NotaHandler) Imprimir(c *gin.Context) {
	numero, err := parseNumero(c)
	if err != nil {
		respondError(c, err)
		return
	}
	n, err := h.service.Imprimir(c.Request.Context(), numero)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, n)
}

func parseNumero(c *gin.Context) (int64, error) {
	numero, err := strconv.ParseInt(c.Param("numero"), 10, 64)
	if err != nil {
		return 0, apperror.Validation("número de nota inválido")
	}
	return numero, nil
}
