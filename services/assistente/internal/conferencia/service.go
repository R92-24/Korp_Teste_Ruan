package conferencia

import (
	"context"
	"log/slog"

	"korp/assistente/internal/apperror"
	"korp/assistente/internal/estoqueclient"
)

type Service struct {
	estoque    *estoqueclient.Client
	analisador *Analisador
}

func NewService(estoque *estoqueclient.Client, analisador *Analisador) *Service {
	return &Service{estoque: estoque, analisador: analisador}
}

func (s *Service) Conferir(ctx context.Context, req Request) (*Resultado, error) {
	if len(req.Itens) == 0 {
		return nil, apperror.Validation("a nota não possui itens para conferir")
	}

	saldos := s.consultarSaldos(ctx, req.Itens)
	observacoesRegras := aplicarRegras(req.Itens, saldos)

	resultado := &Resultado{
		Numero:      req.Numero,
		Observacoes: append([]Observacao{}, observacoesRegras...),
	}

	if !s.analisador.Disponivel() {
		resultado.IADisponivel = false
		resultado.MotivoIA = "Nenhuma chave de API configurada neste ambiente (ANTHROPIC_API_KEY). Exibindo apenas as verificações automáticas."
		resultado.Resumo = resumoDeterministico(req.Itens, observacoesRegras)
		return resultado, nil
	}

	resumoIA, obsIA, err := s.analisador.Analisar(ctx, req.Itens, saldos, observacoesRegras)
	if err != nil {
		// A IA é um reforço, não uma dependência crítica: se ela falhar (rede,
		// limite de uso, chave inválida), a conferência continua com o que as
		// regras determinísticas já produziram, e isso é comunicado ao usuário
		// em vez de falhar a requisição inteira.
		slog.Warn("análise por IA indisponível, seguindo apenas com regras", "erro", err)
		resultado.IADisponivel = false
		resultado.MotivoIA = "A análise por IA não respondeu neste momento. Exibindo apenas as verificações automáticas."
		resultado.Resumo = resumoDeterministico(req.Itens, observacoesRegras)
		return resultado, nil
	}

	resultado.IADisponivel = true
	resultado.Resumo = resumoIA
	resultado.Observacoes = append(resultado.Observacoes, obsIA...)
	return resultado, nil
}

func (s *Service) consultarSaldos(ctx context.Context, itens []Item) map[string]saldoConhecido {
	saldos := map[string]saldoConhecido{}
	for _, it := range itens {
		if _, ja := saldos[it.Codigo]; ja {
			continue
		}
		produto, err := s.estoque.GetProduto(ctx, it.Codigo)
		if err != nil || produto == nil {
			saldos[it.Codigo] = saldoConhecido{Conhecido: false}
			continue
		}
		saldos[it.Codigo] = saldoConhecido{Saldo: produto.Saldo, Conhecido: true}
	}
	return saldos
}
