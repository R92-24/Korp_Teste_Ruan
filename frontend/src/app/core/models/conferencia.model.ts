export type Severidade = 'info' | 'atencao' | 'alerta';
export type OrigemObservacao = 'regra' | 'ia';

export interface Observacao {
  severidade: Severidade;
  titulo: string;
  detalhe: string;
  origem: OrigemObservacao;
}

export interface ConferenciaItemInput {
  codigo: string;
  descricao: string;
  quantidade: number;
}

export interface ConferenciaRequest {
  numero: number;
  itens: ConferenciaItemInput[];
}

export interface ConferenciaResultado {
  numero: number;
  iaDisponivel: boolean;
  motivoIa?: string;
  resumo: string;
  observacoes: Observacao[];
}
