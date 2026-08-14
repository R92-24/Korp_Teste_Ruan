import { Service, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of, timer } from 'rxjs';
import {
  catchError,
  distinctUntilChanged,
  map,
  shareReplay,
  startWith,
  switchMap,
  timeout,
} from 'rxjs/operators';

import { environment } from '../../../environments/environment';

export type StatusServico = 'verificando' | 'online' | 'offline';

export interface Servico {
  nome: string;
  endereco: string;
  status$: Observable<StatusServico>;
}

/**
 * Monitora a disponibilidade dos microsserviços em tempo real.
 *
 * Existe para tornar a arquitetura visível na interface: cada tela mostra de
 * qual serviço ela depende e se ele está no ar. Quando um dos serviços cai, o
 * usuário vê o indicador mudar antes mesmo de tentar uma operação, em vez de
 * descobrir só ao receber um erro.
 */
@Service()
export class HealthService {
  private readonly http = inject(HttpClient);

  readonly estoque: Servico = {
    nome: 'Serviço de Estoque',
    endereco: hostDe(environment.estoqueApiUrl),
    status$: this.monitorar(environment.estoqueApiUrl),
  };

  readonly faturamento: Servico = {
    nome: 'Serviço de Faturamento',
    endereco: hostDe(environment.faturamentoApiUrl),
    status$: this.monitorar(environment.faturamentoApiUrl),
  };

  private monitorar(baseUrl: string): Observable<StatusServico> {
    return timer(0, 5000).pipe(
      // switchMap descarta uma verificação ainda em voo quando o próximo
      // tick chega, evitando respostas fora de ordem.
      switchMap(() =>
        this.http.get(`${baseUrl}/health`, { responseType: 'text' }).pipe(
          timeout(2000),
          map((): StatusServico => 'online'),
          // O catchError fica *dentro* do switchMap de propósito: se ficasse
          // fora, o primeiro erro encerraria o fluxo e o monitoramento nunca
          // mais se recuperaria quando o serviço voltasse.
          catchError(() => of<StatusServico>('offline')),
        ),
      ),
      startWith<StatusServico>('verificando'),
      distinctUntilChanged(),
      // Um único fluxo de polling é compartilhado por todos os componentes
      // que exibem este serviço; refCount encerra o timer quando ninguém
      // mais está observando.
      shareReplay({ bufferSize: 1, refCount: true }),
    );
  }
}

function hostDe(url: string): string {
  try {
    return new URL(url).host;
  } catch {
    return url;
  }
}
