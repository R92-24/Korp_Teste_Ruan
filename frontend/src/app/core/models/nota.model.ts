export type StatusNota = 'Aberta' | 'Fechada';

export interface ItemNota {
  id: number;
  codigo: string;
  descricao: string;
  quantidade: number;
}

export interface NotaFiscal {
  numero: number;
  status: StatusNota;
  createdAt: string;
  closedAt?: string;
  itens: ItemNota[];
}

export interface AddItemInput {
  codigo: string;
  quantidade: number;
}
