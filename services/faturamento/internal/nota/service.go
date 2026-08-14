package nota

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"korp/faturamento/internal/apperror"
	"korp/faturamento/internal/estoqueclient"
)

type Service struct {
	repo    *Repository
	estoque *estoqueclient.Client
}

func NewService(repo *Repository, estoque *estoqueclient.Client) *Service {
	return &Service{repo: repo, estoque: estoque}
}

func (s *Service) Create(ctx context.Context) (*NotaFiscal, error) {
	n, err := s.repo.Create(ctx)
	if err != nil {
		return nil, apperror.Internal("falha ao criar nota fiscal")
	}
	return n, nil
}

func (s *Service) List(ctx context.Context) ([]*NotaFiscal, error) {
	notas, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("falha ao listar notas fiscais")
	}
	return notas, nil
}

func (s *Service) GetByNumero(ctx context.Context, numero int64) (*NotaFiscal, error) {
	n, err := s.repo.GetByNumero(ctx, numero)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, apperror.NotFound("nota fiscal não encontrada")
		}
		return nil, apperror.Internal("falha ao buscar nota fiscal")
	}
	return n, nil
}

func (s *Service) AddItem(ctx context.Context, numero int64, in AddItemInput) (*Item, error) {
	codigo := strings.TrimSpace(in.Codigo)
	if codigo == "" || in.Quantidade <= 0 {
		return nil, apperror.Validation("código do produto e quantidade (> 0) são obrigatórios")
	}

	produto, err := s.estoque.GetProduto(ctx, codigo)
	if err != nil {
		return nil, mapEstoqueError(err, codigo)
	}

	item, err := s.repo.AddItem(ctx, numero, produto.Codigo, produto.Descricao, in.Quantidade)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, apperror.NotFound("nota fiscal não encontrada")
		}
		if errors.Is(err, ErrNaoAberta) {
			return nil, apperror.Conflict("NOTA_NAO_ABERTA", "só é possível incluir itens em notas com status Aberta")
		}
		return nil, apperror.Internal("falha ao incluir item na nota")
	}
	return item, nil
}

func (s *Service) RemoveItem(ctx context.Context, numero, itemID int64) error {
	err := s.repo.RemoveItem(ctx, numero, itemID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperror.NotFound("item ou nota fiscal não encontrado")
		}
		if errors.Is(err, ErrNaoAberta) {
			return apperror.Conflict("NOTA_NAO_ABERTA", "só é possível remover itens de notas com status Aberta")
		}
		return apperror.Internal("falha ao remover item da nota")
	}
	return nil
}

// Imprimir é o fluxo obrigatório do teste: fecha a nota, debita o estoque de
// cada item e, se algo falhar no meio do caminho (estoque indisponível ou
// saldo insuficiente), compensa o que já foi debitado e reabre a nota,
// devolvendo um erro claro para o usuário.
func (s *Service) Imprimir(ctx context.Context, numero int64) (*NotaFiscal, error) {
	atual, err := s.repo.GetByNumero(ctx, numero)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, apperror.NotFound("nota fiscal não encontrada")
		}
		return nil, apperror.Internal("falha ao buscar nota fiscal")
	}
	if atual.Status != StatusAberta {
		return nil, apperror.Conflict("NOTA_NAO_ABERTA", "só é possível imprimir notas com status Aberta")
	}
	if len(atual.Itens) == 0 {
		return nil, apperror.Validation("a nota não possui itens para impressão")
	}

	// Fechamento otimista e atômico: garante que, mesmo com duas requisições
	// de impressão simultâneas para esta nota, só uma prossiga.
	if err := s.repo.FecharSeAberta(ctx, numero); err != nil {
		return nil, apperror.Conflict("NOTA_NAO_ABERTA", "esta nota já está sendo ou já foi impressa")
	}

	var baixados []Item
	for _, item := range atual.Itens {
		_, err := s.estoque.Baixa(ctx, item.Codigo, item.Quantidade)
		if err != nil {
			s.compensar(ctx, numero, baixados)
			return nil, mapEstoqueError(err, item.Codigo)
		}
		baixados = append(baixados, item)
	}

	fechada, err := s.repo.GetByNumero(ctx, numero)
	if err != nil {
		return nil, apperror.Internal("nota impressa, mas falha ao recarregar dados atualizados")
	}
	return fechada, nil
}

// compensar estorna (em ordem reversa) os itens já debitados nesta tentativa
// de impressão e reabre a nota. É best-effort: se o próprio estorno falhar
// (ex.: estoque também ficou indisponível durante a compensação), a falha é
// registrada em log para correção manual, já que a alternativa (deixar o
// usuário sem qualquer feedback) é pior.
func (s *Service) compensar(ctx context.Context, numero int64, baixados []Item) {
	for i := len(baixados) - 1; i >= 0; i-- {
		item := baixados[i]
		if _, err := s.estoque.Estorno(ctx, item.Codigo, item.Quantidade); err != nil {
			slog.Error("falha ao estornar item durante compensação",
				"nota", numero, "produto", item.Codigo, "quantidade", item.Quantidade, "erro", err)
		}
	}
	if err := s.repo.Reabrir(ctx, numero); err != nil {
		slog.Error("falha ao reabrir nota após compensação", "nota", numero, "erro", err)
	}
}

func mapEstoqueError(err error, codigo string) error {
	switch {
	case errors.Is(err, estoqueclient.ErrIndisponivel):
		return apperror.Unavailable("serviço de estoque indisponível no momento. A nota permanece aberta — tente novamente em instantes")
	case errors.Is(err, estoqueclient.ErrSaldoInsuficiente):
		return apperror.Conflict("SALDO_INSUFICIENTE", "saldo insuficiente para o produto "+codigo)
	case errors.Is(err, estoqueclient.ErrProdutoNaoEncontrado):
		return apperror.NotFound("produto " + codigo + " não encontrado no estoque")
	default:
		return apperror.Internal("falha inesperada ao comunicar com o serviço de estoque")
	}
}
