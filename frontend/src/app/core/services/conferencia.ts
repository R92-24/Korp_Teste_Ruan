import { Service, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { ConferenciaRequest, ConferenciaResultado } from '../models/conferencia.model';

@Service()
export class ConferenciaService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.assistenteApiUrl}`;

  conferir(request: ConferenciaRequest): Observable<ConferenciaResultado> {
    return this.http.post<ConferenciaResultado>(`${this.baseUrl}/conferencia`, request);
  }
}
