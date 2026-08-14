package estoqueclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// novoClienteRapido devolve um client apontado para o servidor de teste, com
// o atraso entre tentativas reduzido para não deixar a suíte lenta.
func novoClienteRapido(baseURL string) *Client {
	c := New(baseURL)
	c.retryDelay = time.Millisecond
	return c
}

func TestBaixa_Sucesso(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("método = %s, esperado POST", r.Method)
		}
		if r.URL.Path != "/produtos/P001/baixa" {
			t.Errorf("path = %s, esperado /produtos/P001/baixa", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"codigo":"P001","descricao":"Mouse","saldo":8}`))
	}))
	defer srv.Close()

	produto, err := novoClienteRapido(srv.URL).Baixa(context.Background(), "P001", 2)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if produto.Saldo != 8 {
		t.Errorf("saldo = %d, esperado 8", produto.Saldo)
	}
}

// Uma indisponibilidade momentânea do Estoque não deve derrubar a impressão:
// o cliente precisa tentar de novo e seguir adiante quando o serviço responde.
func TestBaixa_RetentaEmFalhaTransitoriaEDepoisSucede(t *testing.T) {
	var chamadas int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&chamadas, 1) <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte(`{"codigo":"P001","descricao":"Mouse","saldo":8}`))
	}))
	defer srv.Close()

	_, err := novoClienteRapido(srv.URL).Baixa(context.Background(), "P001", 2)
	if err != nil {
		t.Fatalf("deveria ter se recuperado após as falhas, mas retornou: %v", err)
	}
	if got := atomic.LoadInt32(&chamadas); got != 3 {
		t.Errorf("chamadas = %d, esperado 3 (1 tentativa + 2 retries)", got)
	}
}

// Se o Estoque continua fora do ar, o erro precisa chegar como
// ErrIndisponivel — é ele que o Faturamento traduz em "503, tente novamente"
// para o usuário, mantendo a nota aberta.
func TestBaixa_DesisteEDevolveIndisponivel(t *testing.T) {
	var chamadas int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&chamadas, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := novoClienteRapido(srv.URL).Baixa(context.Background(), "P001", 2)
	if !errors.Is(err, ErrIndisponivel) {
		t.Fatalf("erro = %v, esperado ErrIndisponivel", err)
	}
	if got := atomic.LoadInt32(&chamadas); got != 3 {
		t.Errorf("chamadas = %d, esperado 3 (não deve tentar indefinidamente)", got)
	}
}

// Saldo insuficiente é erro de negócio: repetir a requisição não mudaria o
// resultado, então o cliente não deve gastar tentativas com isso.
func TestBaixa_NaoRetentaSaldoInsuficiente(t *testing.T) {
	var chamadas int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&chamadas, 1)
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":{"code":"SALDO_INSUFICIENTE","message":"saldo insuficiente"}}`))
	}))
	defer srv.Close()

	_, err := novoClienteRapido(srv.URL).Baixa(context.Background(), "P001", 99)
	if !errors.Is(err, ErrSaldoInsuficiente) {
		t.Fatalf("erro = %v, esperado ErrSaldoInsuficiente", err)
	}
	if got := atomic.LoadInt32(&chamadas); got != 1 {
		t.Errorf("chamadas = %d, esperado 1 (erro de negócio não se repete)", got)
	}
}

func TestGetProduto_NaoEncontradoNaoRetenta(t *testing.T) {
	var chamadas int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&chamadas, 1)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"code":"NOT_FOUND","message":"produto não encontrado"}}`))
	}))
	defer srv.Close()

	_, err := novoClienteRapido(srv.URL).GetProduto(context.Background(), "INEXISTENTE")
	if !errors.Is(err, ErrProdutoNaoEncontrado) {
		t.Fatalf("erro = %v, esperado ErrProdutoNaoEncontrado", err)
	}
	if got := atomic.LoadInt32(&chamadas); got != 1 {
		t.Errorf("chamadas = %d, esperado 1", got)
	}
}

// Serviço completamente fora do ar (conexão recusada), que é o cenário do
// 'docker compose stop estoque' exigido no teste.
func TestBaixa_ServicoForaDoAr(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close() // derruba antes de chamar: ninguém escutando nessa porta

	_, err := novoClienteRapido(url).Baixa(context.Background(), "P001", 1)
	if !errors.Is(err, ErrIndisponivel) {
		t.Fatalf("erro = %v, esperado ErrIndisponivel", err)
	}
}

func TestEstorno_ChamaEndpointCorreto(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Write([]byte(`{"codigo":"P001","descricao":"Mouse","saldo":10}`))
	}))
	defer srv.Close()

	if _, err := novoClienteRapido(srv.URL).Estorno(context.Background(), "P001", 2); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if path != "/produtos/P001/estorno" {
		t.Errorf("path = %s, esperado /produtos/P001/estorno", path)
	}
}
