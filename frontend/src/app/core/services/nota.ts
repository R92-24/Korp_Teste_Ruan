import { Service, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { AddItemInput, ItemNota, NotaFiscal } from '../models/nota.model';

@Service()
export class NotaService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.faturamentoApiUrl}/notas`;

  list(): Observable<NotaFiscal[]> {
    return this.http.get<NotaFiscal[]>(this.baseUrl);
  }

  create(): Observable<NotaFiscal> {
    return this.http.post<NotaFiscal>(this.baseUrl, {});
  }

  getByNumero(numero: number): Observable<NotaFiscal> {
    return this.http.get<NotaFiscal>(`${this.baseUrl}/${numero}`);
  }

  addItem(numero: number, input: AddItemInput): Observable<ItemNota> {
    return this.http.post<ItemNota>(`${this.baseUrl}/${numero}/itens`, input);
  }

  removeItem(numero: number, itemId: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${numero}/itens/${itemId}`);
  }

  imprimir(numero: number): Observable<NotaFiscal> {
    return this.http.post<NotaFiscal>(`${this.baseUrl}/${numero}/imprimir`, {});
  }
}
