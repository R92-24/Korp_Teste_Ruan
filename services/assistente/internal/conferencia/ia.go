package conferencia

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const promptSistema = `Você revisa notas fiscais antes da impressão, num sistema de emissão de notas.

A impressão é irreversível: ela fecha a nota e debita o estoque. Seu papel é ser
o segundo par de olhos de quem vai clicar em imprimir.

Você recebe os itens da nota, o saldo atual de cada produto e as observações que
verificações determinísticas já produziram. Não repita essas observações — elas
já serão exibidas ao usuário. Acrescente apenas o que a leitura do conjunto
revela e que uma regra fixa não pegaria, por exemplo:

- quantidades que destoam do padrão dos outros itens da mesma nota
- produtos que parecem trocados entre si pela descrição
- uma composição de itens que sugere que a nota está incompleta
- indícios de erro de digitação na quantidade (um dígito a mais, por exemplo)

Responda SOMENTE com um objeto JSON válido, sem cercas de código e sem texto
fora dele, nesta forma:

{
  "resumo": "uma frase dizendo se a nota parece pronta para impressão",
  "observacoes": [
    {"severidade": "info|atencao|alerta", "titulo": "curto e direto", "detalhe": "uma ou duas frases"}
  ]
}

Use "alerta" apenas para algo que provavelmente impede a impressão ou causa
prejuízo, "atencao" para o que merece revisão, e "info" para contexto útil.
Se nada além das verificações automáticas chamar atenção, devolva a lista
"observacoes" vazia e diga isso no resumo. Não invente problemas para preencher
a resposta.`

// Analisador encapsula a chamada ao modelo. Quando não há chave configurada,
// Disponivel() é falso e o serviço segue apenas com as regras determinísticas.
type Analisador struct {
	client    anthropic.Client
	modelo    string
	habilitado bool
}

func NovoAnalisador(apiKey, modelo string) *Analisador {
	if strings.TrimSpace(apiKey) == "" {
		return &Analisador{habilitado: false}
	}
	return &Analisador{
		client:     anthropic.NewClient(option.WithAPIKey(apiKey)),
		modelo:     modelo,
		habilitado: true,
	}
}

func (a *Analisador) Disponivel() bool { return a.habilitado }

type respostaIA struct {
	Resumo      string `json:"resumo"`
	Observacoes []struct {
		Severidade string `json:"severidade"`
		Titulo     string `json:"titulo"`
		Detalhe    string `json:"detalhe"`
	} `json:"observacoes"`
}

// Analisar pede ao modelo uma revisão da nota. Devolve o resumo e as
// observações adicionais; erros de comunicação são devolvidos ao chamador, que
// decide seguir apenas com as regras.
func (a *Analisador) Analisar(ctx context.Context, itens []Item, saldos map[string]saldoConhecido, jaObservado []Observacao) (string, []Observacao, error) {
	if !a.habilitado {
		return "", nil, fmt.Errorf("análise por IA não configurada")
	}

	resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(a.modelo),
		MaxTokens: 8000,
		System: []anthropic.TextBlockParam{{
			Text: promptSistema,
		}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(montarPrompt(itens, saldos, jaObservado))),
		},
	})
	if err != nil {
		return "", nil, err
	}

	var texto strings.Builder
	for _, bloco := range resp.Content {
		if b, ok := bloco.AsAny().(anthropic.TextBlock); ok {
			texto.WriteString(b.Text)
		}
	}

	return interpretarResposta(texto.String())
}

// interpretarResposta é deliberadamente tolerante: se o modelo devolver algo
// que não seja o JSON esperado, o texto ainda é aproveitado como resumo em vez
// de a conferência inteira falhar.
func interpretarResposta(texto string) (string, []Observacao, error) {
	limpo := strings.TrimSpace(texto)
	if limpo == "" {
		return "", nil, fmt.Errorf("resposta vazia do modelo")
	}

	if inicio := strings.Index(limpo, "{"); inicio >= 0 {
		if fim := strings.LastIndex(limpo, "}"); fim > inicio {
			var parsed respostaIA
			if err := json.Unmarshal([]byte(limpo[inicio:fim+1]), &parsed); err == nil {
				obs := make([]Observacao, 0, len(parsed.Observacoes))
				for _, o := range parsed.Observacoes {
					obs = append(obs, Observacao{
						Severidade: normalizarSeveridade(o.Severidade),
						Titulo:     o.Titulo,
						Detalhe:    o.Detalhe,
						Origem:     OrigemIA,
					})
				}
				return parsed.Resumo, obs, nil
			}
		}
	}

	return limpo, nil, nil
}

func normalizarSeveridade(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case SeveridadeAlerta:
		return SeveridadeAlerta
	case SeveridadeAtencao, "atenção":
		return SeveridadeAtencao
	default:
		return SeveridadeInfo
	}
}

func montarPrompt(itens []Item, saldos map[string]saldoConhecido, jaObservado []Observacao) string {
	var b strings.Builder

	b.WriteString("Itens da nota:\n")
	for _, it := range itens {
		saldoTexto := "saldo desconhecido"
		if s, ok := saldos[it.Codigo]; ok && s.Conhecido {
			saldoTexto = fmt.Sprintf("saldo atual %d", s.Saldo)
		}
		descricao := it.Descricao
		if descricao == "" {
			descricao = "(sem descrição)"
		}
		fmt.Fprintf(&b, "- %s | %s | quantidade %d | %s\n", it.Codigo, descricao, it.Quantidade, saldoTexto)
	}

	if len(jaObservado) == 0 {
		b.WriteString("\nAs verificações determinísticas não encontraram problemas.\n")
	} else {
		b.WriteString("\nJá apontado pelas verificações determinísticas (não repita):\n")
		for _, o := range jaObservado {
			fmt.Fprintf(&b, "- [%s] %s\n", o.Severidade, o.Titulo)
		}
	}

	return b.String()
}
