package conferencia

import (
	"fmt"
	"strings"
)

// aplicarRegras executa as verificações determinísticas da conferência.
//
// Elas rodam sempre, mesmo sem IA configurada: são checagens exatas, baratas e
// que não dependem de nenhum serviço externo além do saldo já consultado. A
// camada de IA vem depois e interpreta o conjunto — ela não substitui isto.
func aplicarRegras(itens []Item, saldos map[string]saldoConhecido) []Observacao {
	obs := []Observacao{}

	obs = append(obs, verificarDuplicados(itens)...)
	obs = append(obs, verificarSaldos(itens, saldos)...)

	return obs
}

// verificarDuplicados aponta o mesmo produto lançado em mais de um item. Não é
// necessariamente um erro, mas quase sempre é digitação repetida — e depois de
// impressa a nota não pode mais ser corrigida.
func verificarDuplicados(itens []Item) []Observacao {
	ocorrencias := map[string]int{}
	totais := map[string]int{}
	ordem := []string{}

	for _, it := range itens {
		if ocorrencias[it.Codigo] == 0 {
			ordem = append(ordem, it.Codigo)
		}
		ocorrencias[it.Codigo]++
		totais[it.Codigo] += it.Quantidade
	}

	obs := []Observacao{}
	for _, codigo := range ordem {
		if ocorrencias[codigo] > 1 {
			obs = append(obs, Observacao{
				Severidade: SeveridadeAtencao,
				Titulo:     fmt.Sprintf("Produto %s aparece em %d itens", codigo, ocorrencias[codigo]),
				Detalhe: fmt.Sprintf(
					"Somando os lançamentos, a nota baixa %d unidades deste produto. Se a intenção era um único lançamento, remova os itens repetidos antes de imprimir.",
					totais[codigo]),
				Origem: OrigemRegra,
			})
		}
	}
	return obs
}

// verificarSaldos compara o total pedido por produto com o saldo atual. A
// impressão falharia de qualquer forma se o saldo fosse insuficiente, mas
// avisar antes evita a viagem perdida.
func verificarSaldos(itens []Item, saldos map[string]saldoConhecido) []Observacao {
	pedidoPorProduto := map[string]int{}
	ordem := []string{}
	for _, it := range itens {
		if pedidoPorProduto[it.Codigo] == 0 {
			ordem = append(ordem, it.Codigo)
		}
		pedidoPorProduto[it.Codigo] += it.Quantidade
	}

	obs := []Observacao{}
	for _, codigo := range ordem {
		pedido := pedidoPorProduto[codigo]
		saldo, ok := saldos[codigo]
		if !ok || !saldo.Conhecido {
			obs = append(obs, Observacao{
				Severidade: SeveridadeInfo,
				Titulo:     "Saldo de " + codigo + " não pôde ser consultado",
				Detalhe:    "O serviço de Estoque não respondeu a tempo, então este item não foi conferido contra o saldo disponível.",
				Origem:     OrigemRegra,
			})
			continue
		}

		switch {
		case pedido > saldo.Saldo:
			obs = append(obs, Observacao{
				Severidade: SeveridadeAlerta,
				Titulo:     fmt.Sprintf("Saldo insuficiente para %s", codigo),
				Detalhe: fmt.Sprintf(
					"A nota pede %d unidades e o saldo atual é %d. A impressão será recusada nesta condição.",
					pedido, saldo.Saldo),
				Origem: OrigemRegra,
			})
		case pedido == saldo.Saldo:
			obs = append(obs, Observacao{
				Severidade: SeveridadeAtencao,
				Titulo:     fmt.Sprintf("Esta nota zera o estoque de %s", codigo),
				Detalhe: fmt.Sprintf(
					"São %d unidades pedidas para um saldo de %d. Após a impressão o produto fica sem saldo.",
					pedido, saldo.Saldo),
				Origem: OrigemRegra,
			})
		}
	}
	return obs
}

// resumoDeterministico é o texto usado quando a IA não está disponível: sem
// interpretação, apenas a contagem do que foi verificado.
func resumoDeterministico(itens []Item, obs []Observacao) string {
	alertas := 0
	for _, o := range obs {
		if o.Severidade == SeveridadeAlerta {
			alertas++
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Nota com %s.", pluralizar(len(itens), "item", "itens"))
	switch {
	case alertas == 1:
		b.WriteString(" Há 1 problema que impede a impressão.")
	case alertas > 1:
		fmt.Fprintf(&b, " Há %d problemas que impedem a impressão.", alertas)
	case len(obs) > 0:
		fmt.Fprintf(&b, " %s para revisar antes de imprimir.", pluralizar(len(obs), "Há 1 ponto", fmt.Sprintf("Há %d pontos", len(obs))))
	default:
		b.WriteString(" Nenhum problema encontrado nas verificações automáticas.")
	}
	return b.String()
}

func pluralizar(n int, singular, plural string) string {
	if n == 1 {
		if strings.HasPrefix(singular, "Há ") {
			return singular
		}
		return "1 " + singular
	}
	if strings.HasPrefix(plural, "Há ") {
		return plural
	}
	return fmt.Sprintf("%d %s", n, plural)
}
