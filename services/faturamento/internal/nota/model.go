package nota

import "time"

const (
	StatusAberta  = "Aberta"
	StatusFechada = "Fechada"
)

type NotaFiscal struct {
	Numero    int64      `json:"numero"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	ClosedAt  *time.Time `json:"closedAt,omitempty"`
	Itens     []Item     `json:"itens"`
}

type Item struct {
	ID         int64  `json:"id"`
	Codigo     string `json:"codigo"`
	Descricao  string `json:"descricao"`
	Quantidade int    `json:"quantidade"`
}

type AddItemInput struct {
	Codigo     string `json:"codigo" binding:"required"`
	Quantidade int    `json:"quantidade" binding:"required,gt=0"`
}
