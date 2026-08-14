package produto

import "time"

type Produto struct {
	ID        int64     `json:"id"`
	Codigo    string    `json:"codigo"`
	Descricao string    `json:"descricao"`
	Saldo     int       `json:"saldo"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateInput struct {
	Codigo    string `json:"codigo" binding:"required"`
	Descricao string `json:"descricao" binding:"required"`
	Saldo     int    `json:"saldo"`
}

type UpdateInput struct {
	Descricao string `json:"descricao" binding:"required"`
	Saldo     int    `json:"saldo"`
}

type MovimentoInput struct {
	Quantidade int `json:"quantidade" binding:"required,gt=0"`
}
