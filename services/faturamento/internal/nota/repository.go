package nota

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound       = errors.New("nota fiscal não encontrada")
	ErrNaoAberta      = errors.New("nota fiscal não está aberta")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context) (*NotaFiscal, error) {
	const q = `
		INSERT INTO notas (status) VALUES ('Aberta')
		RETURNING numero, status, created_at, closed_at`
	row := r.pool.QueryRow(ctx, q)
	n, err := scanNota(row)
	if err != nil {
		return nil, err
	}
	n.Itens = []Item{}
	return n, nil
}

func (r *Repository) List(ctx context.Context) ([]*NotaFiscal, error) {
	const q = `SELECT numero, status, created_at, closed_at FROM notas ORDER BY numero DESC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notas []*NotaFiscal
	for rows.Next() {
		n, err := scanNota(rows)
		if err != nil {
			return nil, err
		}
		notas = append(notas, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if notas == nil {
		notas = []*NotaFiscal{}
	}
	return notas, nil
}

func (r *Repository) GetByNumero(ctx context.Context, numero int64) (*NotaFiscal, error) {
	const q = `SELECT numero, status, created_at, closed_at FROM notas WHERE numero = $1`
	row := r.pool.QueryRow(ctx, q, numero)
	n, err := scanNota(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	itens, err := r.listItens(ctx, numero)
	if err != nil {
		return nil, err
	}
	n.Itens = itens
	return n, nil
}

func (r *Repository) listItens(ctx context.Context, numero int64) ([]Item, error) {
	const q = `SELECT id, codigo, descricao, quantidade FROM itens_nota WHERE nota_numero = $1 ORDER BY id`
	rows, err := r.pool.Query(ctx, q, numero)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itens := []Item{}
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.Codigo, &it.Descricao, &it.Quantidade); err != nil {
			return nil, err
		}
		itens = append(itens, it)
	}
	return itens, rows.Err()
}

// AddItem insere um item somente se a nota existir e estiver Aberta,
// verificado e escrito dentro da mesma transação com lock de linha para
// evitar inserir itens em uma nota que está sendo fechada nesse instante.
func (r *Repository) AddItem(ctx context.Context, numero int64, codigo, descricao string, quantidade int) (*Item, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM notas WHERE numero = $1 FOR UPDATE`, numero).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if status != StatusAberta {
		return nil, ErrNaoAberta
	}

	var item Item
	const q = `
		INSERT INTO itens_nota (nota_numero, codigo, descricao, quantidade)
		VALUES ($1, $2, $3, $4)
		RETURNING id, codigo, descricao, quantidade`
	err = tx.QueryRow(ctx, q, numero, codigo, descricao, quantidade).Scan(&item.ID, &item.Codigo, &item.Descricao, &item.Quantidade)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) RemoveItem(ctx context.Context, numero, itemID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM notas WHERE numero = $1 FOR UPDATE`, numero).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if status != StatusAberta {
		return ErrNaoAberta
	}

	tag, err := tx.Exec(ctx, `DELETE FROM itens_nota WHERE id = $1 AND nota_numero = $2`, itemID, numero)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return tx.Commit(ctx)
}

// FecharSeAberta transiciona a nota para Fechada de forma atômica, apenas se
// ela ainda estiver Aberta. É o "gate" que garante que, mesmo com duas
// requisições de impressão simultâneas para a MESMA nota, apenas uma delas
// prossiga para debitar o estoque.
func (r *Repository) FecharSeAberta(ctx context.Context, numero int64) error {
	const q = `UPDATE notas SET status = 'Fechada', closed_at = now() WHERE numero = $1 AND status = 'Aberta'`
	tag, err := r.pool.Exec(ctx, q, numero)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNaoAberta
	}
	return nil
}

// Reabrir volta o status para Aberta — usado como compensação quando a
// impressão falha após o fechamento otimista (ex.: estoque indisponível ou
// saldo insuficiente em algum item).
func (r *Repository) Reabrir(ctx context.Context, numero int64) error {
	const q = `UPDATE notas SET status = 'Aberta', closed_at = NULL WHERE numero = $1`
	_, err := r.pool.Exec(ctx, q, numero)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNota(row scanner) (*NotaFiscal, error) {
	var n NotaFiscal
	if err := row.Scan(&n.Numero, &n.Status, &n.CreatedAt, &n.ClosedAt); err != nil {
		return nil, err
	}
	return &n, nil
}
