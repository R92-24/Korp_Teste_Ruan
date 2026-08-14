// Package estoqueclient consulta o serviço de Estoque para obter os saldos
// usados na conferência. A indisponibilidade do Estoque não invalida a
// conferência: os itens afetados são apenas marcados como não conferidos.
package estoqueclient

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Produto struct {
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
	Saldo     int    `json:"saldo"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
}

// GetProduto devolve (nil, nil) quando o produto não existe, para que o
// chamador distinga "não encontrado" de "falha ao consultar".
func (c *Client) GetProduto(ctx context.Context, codigo string) (*Produto, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/produtos/"+codigo, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errStatus{resp.StatusCode}
	}

	var produto Produto
	if err := json.NewDecoder(resp.Body).Decode(&produto); err != nil {
		return nil, err
	}
	return &produto, nil
}

type errStatus struct{ code int }

func (e errStatus) Error() string {
	return "estoque respondeu com status inesperado"
}
