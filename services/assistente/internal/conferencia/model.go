package conferencia

// Severidade classifica o peso de uma observação da conferência.
const (
	SeveridadeInfo     = "info"
	SeveridadeAtencao  = "atencao"
	SeveridadeAlerta   = "alerta"
)

type Item struct {
	Codigo     string `json:"codigo" binding:"required"`
	Descricao  string `json:"descricao"`
	Quantidade int    `json:"quantidade" binding:"required,gt=0"`
}

type Request struct {
	Numero int64  `json:"numero"`
	Itens  []Item `json:"itens" binding:"required,min=1"`
}

type Observacao struct {
	Severidade string `json:"severidade"`
	Titulo     string `json:"titulo"`
	Detalhe    string `json:"detalhe"`
	// Origem distingue o que veio de uma regra determinística do que veio da
	// análise por IA — o usuário precisa saber o que é verificação exata e o
	// que é sugestão sujeita a revisão.
	Origem string `json:"origem"`
}

const (
	OrigemRegra = "regra"
	OrigemIA    = "ia"
)

type Resultado struct {
	Numero int64 `json:"numero"`
	// IADisponivel indica se a análise por IA pôde ser executada. Quando falsa,
	// apenas as verificações determinísticas foram aplicadas.
	IADisponivel bool         `json:"iaDisponivel"`
	MotivoIA     string       `json:"motivoIa,omitempty"`
	Resumo       string       `json:"resumo"`
	Observacoes  []Observacao `json:"observacoes"`
}

// saldoConhecido carrega o saldo atual de um produto, quando o Estoque
// respondeu. Ausente significa que não foi possível consultar.
type saldoConhecido struct {
	Saldo    int
	Conhecido bool
}
