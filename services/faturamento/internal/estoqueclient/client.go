// Package estoqueclient encapsula toda a comunicação HTTP do serviço de
// Faturamento com o serviço de Estoque. É o único ponto onde o cenário de
// falha de microsserviço (obrigatório no teste) é tratado: timeout curto,
// retries em erros transitórios (rede/timeout/5xx) e um erro de domínio
// claro (ErrIndisponivel) quando o Estoque não responde mesmo após as
// tentativas — o serviço de Faturamento usa isso para devolver feedback
// apropriado ao usuário e manter a nota em aberto para nova tentativa.
package estoqueclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrIndisponivel      = errors.New("serviço de estoque indisponível")
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado no estoque")
	ErrSaldoInsuficiente = errors.New("saldo insuficiente")
)

type Produto struct {
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
	Saldo     int    `json:"saldo"`
}

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 3 * time.Second},
		maxRetries: 2,
		retryDelay: 200 * time.Millisecond,
	}
}

func (c *Client) GetProduto(ctx context.Context, codigo string) (*Produto, error) {
	var produto Produto
	err := c.doWithRetry(ctx, http.MethodGet, "/produtos/"+codigo, nil, &produto)
	if err != nil {
		return nil, err
	}
	return &produto, nil
}

func (c *Client) Baixa(ctx context.Context, codigo string, quantidade int) (*Produto, error) {
	body := map[string]int{"quantidade": quantidade}
	var produto Produto
	err := c.doWithRetry(ctx, http.MethodPost, "/produtos/"+codigo+"/baixa", body, &produto)
	if err != nil {
		return nil, err
	}
	return &produto, nil
}

func (c *Client) Estorno(ctx context.Context, codigo string, quantidade int) (*Produto, error) {
	body := map[string]int{"quantidade": quantidade}
	var produto Produto
	err := c.doWithRetry(ctx, http.MethodPost, "/produtos/"+codigo+"/estorno", body, &produto)
	if err != nil {
		return nil, err
	}
	return &produto, nil
}

// doWithRetry executa a requisição e, em caso de falha de rede/timeout ou
// erro 5xx (transitório), tenta novamente algumas vezes com um pequeno
// atraso. Erros de negócio (4xx) não são reexecutados, pois reexecutá-los
// não mudaria o resultado.
func (c *Client) doWithRetry(ctx context.Context, method, path string, body any, out any) error {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		err := c.do(ctx, method, path, body, out)
		if err == nil {
			return nil
		}
		lastErr = err

		if !isTransient(err) {
			return err
		}
		if attempt < c.maxRetries {
			time.Sleep(c.retryDelay)
		}
	}
	return fmt.Errorf("%w: %v", ErrIndisponivel, lastErr)
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return transientNetworkError{err}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out == nil {
			return nil
		}
		return json.NewDecoder(resp.Body).Decode(out)
	}

	var env errorEnvelope
	_ = json.NewDecoder(resp.Body).Decode(&env)

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return ErrProdutoNaoEncontrado
	case resp.StatusCode == http.StatusConflict && env.Error.Code == "SALDO_INSUFICIENTE":
		return ErrSaldoInsuficiente
	case resp.StatusCode >= 500:
		return transientNetworkError{fmt.Errorf("estoque retornou status %d", resp.StatusCode)}
	default:
		if env.Error.Message != "" {
			return errors.New(env.Error.Message)
		}
		return fmt.Errorf("estoque retornou status %d", resp.StatusCode)
	}
}

// transientNetworkError marca falhas que valem retry (conexão recusada,
// timeout, DNS, 5xx) sem que o chamador precise conhecer detalhes de rede.
type transientNetworkError struct{ err error }

func (e transientNetworkError) Error() string { return e.err.Error() }
func (e transientNetworkError) Unwrap() error { return e.err }

func isTransient(err error) bool {
	var t transientNetworkError
	return errors.As(err, &t)
}
