package produto

import (
	"context"
	"errors"
	"strings"

	"korp/estoque/internal/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Produto, error) {
	in.Codigo = strings.TrimSpace(in.Codigo)
	in.Descricao = strings.TrimSpace(in.Descricao)
	if in.Codigo == "" || in.Descricao == "" {
		return nil, apperror.Validation("código e descrição são obrigatórios")
	}
	if in.Saldo < 0 {
		return nil, apperror.Validation("saldo não pode ser negativo")
	}

	p, err := s.repo.Create(ctx, in)
	if err != nil {
		if errors.Is(err, ErrDuplicado) {
			return nil, apperror.Conflict("PRODUTO_DUPLICADO", "já existe um produto com este código")
		}
		return nil, apperror.Internal("falha ao criar produto")
	}
	return p, nil
}

func (s *Service) List(ctx context.Context) ([]*Produto, error) {
	produtos, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("falha ao listar produtos")
	}
	if produtos == nil {
		produtos = []*Produto{}
	}
	return produtos, nil
}

func (s *Service) GetByCodigo(ctx context.Context, codigo string) (*Produto, error) {
	p, err := s.repo.GetByCodigo(ctx, codigo)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, apperror.NotFound("produto não encontrado")
		}
		return nil, apperror.Internal("falha ao buscar produto")
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, codigo string, in UpdateInput) (*Produto, error) {
	in.Descricao = strings.TrimSpace(in.Descricao)
	if in.Descricao == "" {
		return nil, apperror.Validation("descrição é obrigatória")
	}
	if in.Saldo < 0 {
		return nil, apperror.Validation("saldo não pode ser negativo")
	}
	p, err := s.repo.Update(ctx, codigo, in)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, apperror.NotFound("produto não encontrado")
		}
		return nil, apperror.Internal("falha ao atualizar produto")
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, codigo string) error {
	err := s.repo.Delete(ctx, codigo)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("produto não encontrado")
		}
		return apperror.Internal("falha ao remover produto")
	}
	return nil
}

func (s *Service) Baixa(ctx context.Context, codigo string, in MovimentoInput) (*Produto, error) {
	if in.Quantidade <= 0 {
		return nil, apperror.Validation("quantidade deve ser maior que zero")
	}
	p, err := s.repo.Baixa(ctx, codigo, in.Quantidade)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, apperror.NotFound("produto não encontrado")
		}
		if errors.Is(err, ErrSaldoInsuficiente) {
			return nil, apperror.Conflict("SALDO_INSUFICIENTE", "saldo insuficiente para o produto "+codigo)
		}
		return nil, apperror.Internal("falha ao debitar saldo")
	}
	return p, nil
}

func (s *Service) Estorno(ctx context.Context, codigo string, in MovimentoInput) (*Produto, error) {
	if in.Quantidade <= 0 {
		return nil, apperror.Validation("quantidade deve ser maior que zero")
	}
	p, err := s.repo.Estorno(ctx, codigo, in.Quantidade)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, apperror.NotFound("produto não encontrado")
		}
		return nil, apperror.Internal("falha ao estornar saldo")
	}
	return p, nil
}
