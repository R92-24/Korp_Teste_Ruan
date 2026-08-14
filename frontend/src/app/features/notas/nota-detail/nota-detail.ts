import { Component, DestroyRef, OnDestroy, OnInit, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule, DatePipe } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { FormControl, ReactiveFormsModule, Validators } from '@angular/forms';
import { Subject, forkJoin, of, timer } from 'rxjs';
import { debounceTime, distinctUntilChanged, map, switchMap, takeUntil } from 'rxjs/operators';

import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatAutocompleteModule, MatAutocompleteSelectedEvent } from '@angular/material/autocomplete';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatExpansionModule } from '@angular/material/expansion';

import { NotaService } from '../../../core/services/nota';
import { ProdutoService } from '../../../core/services/produto';
import { NotificationService } from '../../../core/services/notification';
import { HealthService } from '../../../core/services/health';
import { ConferenciaService } from '../../../core/services/conferencia';
import { ItemNota, NotaFiscal } from '../../../core/models/nota.model';
import { Produto } from '../../../core/models/produto.model';
import { ConferenciaResultado } from '../../../core/models/conferencia.model';
import { ServiceBadge } from '../../../shared/service-badge/service-badge';

@Component({
  selector: 'app-nota-detail',
  imports: [
    CommonModule,
    DatePipe,
    RouterLink,
    ReactiveFormsModule,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    MatInputModule,
    MatAutocompleteModule,
    MatProgressSpinnerModule,
    MatExpansionModule,
    ServiceBadge,
  ],
  templateUrl: './nota-detail.html',
  styleUrl: './nota-detail.scss',
})
export class NotaDetail implements OnInit, OnDestroy {
  private readonly route = inject(ActivatedRoute);
  private readonly notaService = inject(NotaService);
  private readonly produtoService = inject(ProdutoService);
  private readonly notification = inject(NotificationService);
  private readonly destroyRef = inject(DestroyRef);
  private readonly conferenciaService = inject(ConferenciaService);

  // Imprimir depende dos dois serviços: o Faturamento fecha a nota e o
  // Estoque debita o saldo. Por isso a tela exibe o estado de ambos.
  private readonly health = inject(HealthService);
  readonly servicoFaturamento = this.health.faturamento;
  readonly servicoEstoque = this.health.estoque;
  readonly servicoAssistente = this.health.assistente;

  readonly conferindo = signal(false);
  readonly resultadoConferencia = signal<ConferenciaResultado | null>(null);

  // Subject + takeUntil usados deliberadamente aqui (em conjunto com o hook
  // ngOnDestroy) para demonstrar o padrão clássico de cancelamento de
  // inscrições RxJS, complementando o takeUntilDestroyed usado no restante
  // da aplicação.
  private readonly destroy$ = new Subject<void>();

  readonly nota = signal<NotaFiscal | null>(null);
  readonly loading = signal(false);
  readonly imprimindo = signal(false);
  readonly addingItem = signal(false);
  readonly produtosDisponiveis = signal<Produto[]>([]);
  readonly produtosFiltrados = signal<Produto[]>([]);

  readonly displayedColumns = ['codigo', 'descricao', 'quantidade', 'acoes'];

  readonly buscaProdutoControl = new FormControl('', { nonNullable: true });
  readonly quantidadeControl = new FormControl(1, {
    nonNullable: true,
    validators: [Validators.required, Validators.min(1)],
  });
  selectedProduto: Produto | null = null;

  ngOnInit(): void {
    this.route.paramMap
      .pipe(
        map((params) => Number(params.get('numero'))),
        takeUntil(this.destroy$),
      )
      .subscribe((numero) => this.carregar(numero));

    this.produtoService
      .list()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((produtos) => {
        this.produtosDisponiveis.set(produtos);
        this.produtosFiltrados.set(produtos);
      });

    this.buscaProdutoControl.valueChanges
      .pipe(
        debounceTime(250),
        distinctUntilChanged(),
        switchMap((termo) => of(this.filtrarProdutos(termo))),
        takeUntil(this.destroy$),
      )
      .subscribe((produtos) => this.produtosFiltrados.set(produtos));
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private filtrarProdutos(termo: string): Produto[] {
    const termoNormalizado = (termo ?? '').trim().toLowerCase();
    if (!termoNormalizado) return this.produtosDisponiveis();
    return this.produtosDisponiveis().filter(
      (p) =>
        p.codigo.toLowerCase().includes(termoNormalizado) ||
        p.descricao.toLowerCase().includes(termoNormalizado),
    );
  }

  private carregar(numero: number): void {
    if (!numero) return;
    this.loading.set(true);
    this.notaService
      .getByNumero(numero)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (nota) => {
          this.nota.set(nota);
          this.loading.set(false);
        },
        error: (err) => {
          this.loading.set(false);
          this.notification.errorFromResponse(err, 'Falha ao carregar a nota fiscal.');
        },
      });
  }

  exibirProduto(produto: Produto | string | null): string {
    if (!produto || typeof produto === 'string') return '';
    return `${produto.codigo} - ${produto.descricao}`;
  }

  selecionarProduto(event: MatAutocompleteSelectedEvent): void {
    this.selectedProduto = event.option.value as Produto;
  }

  podeAdicionarItem(): boolean {
    return this.nota()?.status === 'Aberta' && !!this.selectedProduto && this.quantidadeControl.valid;
  }

  adicionarItem(): void {
    const notaAtual = this.nota();
    if (!notaAtual || !this.selectedProduto || this.quantidadeControl.invalid) {
      return;
    }
    this.addingItem.set(true);
    this.notaService
      .addItem(notaAtual.numero, {
        codigo: this.selectedProduto.codigo,
        quantidade: this.quantidadeControl.value,
      })
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.addingItem.set(false);
          this.notification.success('Item incluído na nota.');
          this.selectedProduto = null;
          this.buscaProdutoControl.reset('');
          this.quantidadeControl.reset(1);
          this.resultadoConferencia.set(null);
          this.carregar(notaAtual.numero);
        },
        error: (err) => {
          this.addingItem.set(false);
          this.notification.errorFromResponse(err, 'Falha ao incluir item na nota.');
        },
      });
  }

  removerItem(item: ItemNota): void {
    const notaAtual = this.nota();
    if (!notaAtual) return;
    this.notaService
      .removeItem(notaAtual.numero, item.id)
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: () => {
          this.notification.success('Item removido da nota.');
          this.resultadoConferencia.set(null);
          this.carregar(notaAtual.numero);
        },
        error: (err) => this.notification.errorFromResponse(err, 'Falha ao remover item da nota.'),
      });
  }

  imprimir(): void {
    const notaAtual = this.nota();
    if (!notaAtual) return;
    this.imprimindo.set(true);

    // A impressão costuma responder em algumas dezenas de milissegundos, o que
    // faria o indicador de processamento piscar e sumir antes de o usuário
    // perceber. O forkJoin com um timer garante um tempo mínimo de exibição,
    // evitando esse "flash" de estado de carregamento. Em caso de erro o
    // forkJoin propaga na hora, sem esperar o timer.
    forkJoin([this.notaService.imprimir(notaAtual.numero), timer(700)])
      .pipe(
        map(([nota]) => nota),
        takeUntil(this.destroy$),
      )
      .subscribe({
        next: (nota) => {
          this.imprimindo.set(false);
          this.nota.set(nota);
          this.notification.success(`Nota nº ${nota.numero} impressa e fechada com sucesso.`);
        },
        error: (err) => {
          this.imprimindo.set(false);
          this.notification.errorFromResponse(err, 'Falha ao imprimir a nota. A nota permanece aberta.');
        },
      });
  }

  // Requisito opcional de uso de IA: antes de imprimir, o usuário pode pedir
  // uma conferência da nota. O backend sempre roda verificações
  // determinísticas (duplicidade, saldo); quando há uma chave de API
  // configurada, complementa com uma análise por IA. O botão funciona nos
  // dois cenários — a interface só muda o que exibe conforme
  // `resultado.iaDisponivel`.
  conferir(): void {
    const notaAtual = this.nota();
    if (!notaAtual || notaAtual.itens.length === 0) return;

    this.conferindo.set(true);
    this.resultadoConferencia.set(null);

    this.conferenciaService
      .conferir({
        numero: notaAtual.numero,
        itens: notaAtual.itens.map((item) => ({
          codigo: item.codigo,
          descricao: item.descricao,
          quantidade: item.quantidade,
        })),
      })
      .pipe(takeUntil(this.destroy$))
      .subscribe({
        next: (resultado) => {
          this.conferindo.set(false);
          this.resultadoConferencia.set(resultado);
        },
        error: (err) => {
          this.conferindo.set(false);
          this.notification.errorFromResponse(err, 'Falha ao conferir a nota.');
        },
      });
  }
}
