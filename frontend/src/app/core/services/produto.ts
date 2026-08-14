import { Service, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { CreateProdutoInput, Produto, UpdateProdutoInput } from '../models/produto.model';

@Service()
export class ProdutoService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.estoqueApiUrl}/produtos`;

  list(): Observable<Produto[]> {
    return this.http.get<Produto[]>(this.baseUrl);
  }

  create(input: CreateProdutoInput): Observable<Produto> {
    return this.http.post<Produto>(this.baseUrl, input);
  }

  update(codigo: string, input: UpdateProdutoInput): Observable<Produto> {
    return this.http.put<Produto>(`${this.baseUrl}/${codigo}`, input);
  }

  remove(codigo: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${codigo}`);
  }
}
