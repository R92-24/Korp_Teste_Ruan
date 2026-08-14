package produto

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Estes testes exercitam o comportamento real do PostgreSQL — em especial a
// atomicidade da baixa de saldo, que é justamente o que resolve a corrida
// entre duas notas disputando o mesmo produto. Testar isso com um repositório
// falso não provaria nada, porque a garantia vem do banco, não do código Go.
//
// Rode com:
//   TEST_DATABASE_URL=postgres://korp:korp@localhost:5433/estoque?sslmode=disable go test ./...
//
// Sem essa variável os testes são pulados, para não quebrar o build de quem
// não tem um banco à mão.
func abrirPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL não definida; pulando testes de integração")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("não foi possível conectar ao banco de teste: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("banco de teste não respondeu: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// criarProduto insere um produto com código único e agenda sua remoção,
// para que os testes não interfiram uns nos outros nem sujem o banco.
func criarProduto(t *testing.T, repo *Repository, saldo int) *Produto {
	t.Helper()

	ctx := context.Background()
	codigo := fmt.Sprintf("TEST-%s", t.Name())
	// Um nome de teste pode conter '/' em subtestes; o código é só um
	// identificador, então normalizar basta.
	for i, r := range codigo {
		if r == '/' {
			codigo = codigo[:i] + "-" + codigo[i+1:]
		}
	}

	_, _ = repo.pool.Exec(ctx, `DELETE FROM produtos WHERE codigo = $1`, codigo)

	p, err := repo.Create(ctx, CreateInput{
		Codigo:    codigo,
		Descricao: "Produto de teste",
		Saldo:     saldo,
	})
	if err != nil {
		t.Fatalf("falha ao criar produto de teste: %v", err)
	}

	t.Cleanup(func() {
		_, _ = repo.pool.Exec(context.Background(), `DELETE FROM produtos WHERE codigo = $1`, codigo)
	})
	return p
}

func TestBaixa_DebitaSaldo(t *testing.T) {
	repo := NewRepository(abrirPool(t))
	p := criarProduto(t, repo, 10)

	atualizado, err := repo.Baixa(context.Background(), p.Codigo, 2)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if atualizado.Saldo != 8 {
		t.Errorf("saldo = %d, esperado 8 (10 - 2)", atualizado.Saldo)
	}
}

func TestBaixa_RecusaQuandoSaldoInsuficiente(t *testing.T) {
	repo := NewRepository(abrirPool(t))
	p := criarProduto(t, repo, 1)

	_, err := repo.Baixa(context.Background(), p.Codigo, 5)
	if !errors.Is(err, ErrSaldoInsuficiente) {
		t.Fatalf("erro = %v, esperado ErrSaldoInsuficiente", err)
	}

	inalterado, err := repo.GetByCodigo(context.Background(), p.Codigo)
	if err != nil {
		t.Fatalf("erro ao reler produto: %v", err)
	}
	if inalterado.Saldo != 1 {
		t.Errorf("saldo = %d, esperado 1 (uma baixa recusada não pode alterar o estoque)", inalterado.Saldo)
	}
}

func TestBaixa_ProdutoInexistente(t *testing.T) {
	repo := NewRepository(abrirPool(t))

	_, err := repo.Baixa(context.Background(), "PRODUTO-QUE-NAO-EXISTE", 1)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("erro = %v, esperado ErrNotFound", err)
	}
}

func TestEstorno_DevolveSaldo(t *testing.T) {
	repo := NewRepository(abrirPool(t))
	p := criarProduto(t, repo, 5)
	ctx := context.Background()

	if _, err := repo.Baixa(ctx, p.Codigo, 3); err != nil {
		t.Fatalf("falha na baixa: %v", err)
	}
	estornado, err := repo.Estorno(ctx, p.Codigo, 3)
	if err != nil {
		t.Fatalf("falha no estorno: %v", err)
	}
	if estornado.Saldo != 5 {
		t.Errorf("saldo = %d, esperado 5 (a compensação deve restaurar o saldo original)", estornado.Saldo)
	}
}

// O cenário do requisito opcional: um produto com saldo 1 sendo disputado por
// várias notas ao mesmo tempo. Exatamente uma baixa pode vencer, e o saldo
// jamais pode ficar negativo.
func TestBaixa_ConcorrenciaSaldoUnitario(t *testing.T) {
	repo := NewRepository(abrirPool(t))
	p := criarProduto(t, repo, 1)

	const disputantes = 12
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		sucessos  int
		recusas   int
		inesperado error
	)

	largada := make(chan struct{})
	for i := 0; i < disputantes; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-largada // todos partem no mesmo instante
			_, err := repo.Baixa(context.Background(), p.Codigo, 1)

			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				sucessos++
			case errors.Is(err, ErrSaldoInsuficiente):
				recusas++
			default:
				inesperado = err
			}
		}()
	}
	close(largada)
	wg.Wait()

	if inesperado != nil {
		t.Fatalf("erro inesperado durante a disputa: %v", inesperado)
	}
	if sucessos != 1 {
		t.Errorf("baixas bem-sucedidas = %d, esperado exatamente 1", sucessos)
	}
	if recusas != disputantes-1 {
		t.Errorf("baixas recusadas = %d, esperado %d", recusas, disputantes-1)
	}

	final, err := repo.GetByCodigo(context.Background(), p.Codigo)
	if err != nil {
		t.Fatalf("erro ao reler produto: %v", err)
	}
	if final.Saldo != 0 {
		t.Errorf("saldo final = %d, esperado 0 (nunca negativo)", final.Saldo)
	}
}

func TestCreate_RecusaCodigoDuplicado(t *testing.T) {
	repo := NewRepository(abrirPool(t))
	p := criarProduto(t, repo, 1)

	_, err := repo.Create(context.Background(), CreateInput{
		Codigo:    p.Codigo,
		Descricao: "Outro produto com o mesmo código",
		Saldo:     5,
	})
	if !errors.Is(err, ErrDuplicado) {
		t.Fatalf("erro = %v, esperado ErrDuplicado", err)
	}
}
