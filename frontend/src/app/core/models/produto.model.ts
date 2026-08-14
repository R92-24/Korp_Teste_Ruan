export interface Produto {
  id: number;
  codigo: string;
  descricao: string;
  saldo: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateProdutoInput {
  codigo: string;
  descricao: string;
  saldo: number;
}

export interface UpdateProdutoInput {
  descricao: string;
  saldo: number;
}
