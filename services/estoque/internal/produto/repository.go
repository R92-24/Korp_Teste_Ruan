package produto

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("produto não encontrado")
var ErrDuplicado = errors.New("já existe um produto com este código")
var ErrSaldoInsuficiente = errors.New("saldo insuficiente")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, in CreateInput) (*Produto, error) {
	const q = `
		INSERT INTO produtos (codigo, descricao, saldo)
		VALUES ($1, $2, $3)
		RETURNING id, codigo, descricao, saldo, created_at, updated_at`
	row := r.pool.QueryRow(ctx, q, in.Codigo, in.Descricao, in.Saldo)
	p, err := scanProduto(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicado
		}
		return nil, err
	}
	return p, nil
}

func (r *Repository) List(ctx context.Context) ([]*Produto, error) {
	const q = `SELECT id, codigo, descricao, saldo, created_at, updated_at FROM produtos ORDER BY codigo`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var produtos []*Produto
	for rows.Next() {
		p, err := scanProduto(rows)
		if err != nil {
			return nil, err
		}
		produtos = append(produtos, p)
	}
	return produtos, rows.Err()
}

func (r *Repository) GetByCodigo(ctx context.Context, codigo string) (*Produto, error) {
	const q = `SELECT id, codigo, descricao, saldo, created_at, updated_at FROM produtos WHERE codigo = $1`
	row := r.pool.QueryRow(ctx, q, codigo)
	p, err := scanProduto(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *Repository) Update(ctx context.Context, codigo string, in UpdateInput) (*Produto, error) {
	const q = `
		UPDATE produtos SET descricao = $2, saldo = $3, updated_at = now()
		WHERE codigo = $1
		RETURNING id, codigo, descricao, saldo, created_at, updated_at`
	row := r.pool.QueryRow(ctx, q, codigo, in.Descricao, in.Saldo)
	p, err := scanProduto(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *Repository) Delete(ctx context.Context, codigo string) error {
	const q = `DELETE FROM produtos WHERE codigo = $1`
	tag, err := r.pool.Exec(ctx, q, codigo)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Baixa debita saldo de forma atômica: a condição "saldo >= $2" garante que,
// sob concorrência, apenas requisições cujo saldo ainda comporte a baixa no
// momento exato do UPDATE terão sucesso — o Postgres serializa updates
// concorrentes na mesma linha, então não há necessidade de locks explícitos
// nem de uma transação com SELECT ... FOR UPDATE separado.
func (r *Repository) Baixa(ctx context.Context, codigo string, quantidade int) (*Produto, error) {
	const q = `
		UPDATE produtos SET saldo = saldo - $2, updated_at = now()
		WHERE codigo = $1 AND saldo >= $2
		RETURNING id, codigo, descricao, saldo, created_at, updated_at`
	row := r.pool.QueryRow(ctx, q, codigo, quantidade)
	p, err := scanProduto(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if _, getErr := r.GetByCodigo(ctx, codigo); errors.Is(getErr, ErrNotFound) {
				return nil, ErrNotFound
			}
			return nil, ErrSaldoInsuficiente
		}
		return nil, err
	}
	return p, nil
}

// Estorno credita saldo de volta (compensação usada pelo Faturamento quando
// uma impressão de nota falha parcialmente).
func (r *Repository) Estorno(ctx context.Context, codigo string, quantidade int) (*Produto, error) {
	const q = `
		UPDATE produtos SET saldo = saldo + $2, updated_at = now()
		WHERE codigo = $1
		RETURNING id, codigo, descricao, saldo, created_at, updated_at`
	row := r.pool.QueryRow(ctx, q, codigo, quantidade)
	p, err := scanProduto(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProduto(row scanner) (*Produto, error) {
	var p Produto
	err := row.Scan(&p.ID, &p.Codigo, &p.Descricao, &p.Saldo, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
