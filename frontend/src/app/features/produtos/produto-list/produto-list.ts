import { Component, DestroyRef, OnInit, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatTableModule } from '@angular/material/table';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatDialog } from '@angular/material/dialog';

import { ProdutoService } from '../../../core/services/produto';
import { NotificationService } from '../../../core/services/notification';
import { Produto } from '../../../core/models/produto.model';
import { ConfirmDialog } from '../../../shared/confirm-dialog/confirm-dialog';

@Component({
  selector: 'app-produto-list',
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatTableModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './produto-list.html',
  styleUrl: './produto-list.scss',
})
export class ProdutoList implements OnInit {
  private readonly produtoService = inject(ProdutoService);
  private readonly notification = inject(NotificationService);
  private readonly dialog = inject(MatDialog);
  private readonly fb = inject(FormBuilder);
  private readonly destroyRef = inject(DestroyRef);

  readonly displayedColumns = ['codigo', 'descricao', 'saldo', 'acoes'];
  readonly produtos = signal<Produto[]>([]);
  readonly loading = signal(false);
  readonly saving = signal(false);
  readonly editingCodigo = signal<string | null>(null);

  readonly form = this.fb.nonNullable.group({
    codigo: ['', [Validators.required, Validators.maxLength(50)]],
    descricao: ['', [Validators.required, Validators.maxLength(200)]],
    saldo: [0, [Validators.required, Validators.min(0)]],
  });

  ngOnInit(): void {
    this.carregar();
  }

  carregar(): void {
    this.loading.set(true);
    this.produtoService
      .list()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (produtos) => {
          this.produtos.set(produtos);
          this.loading.set(false);
        },
        error: (err) => {
          this.loading.set(false);
          this.notification.errorFromResponse(err, 'Falha ao carregar produtos.');
        },
      });
  }

  editar(produto: Produto): void {
    this.editingCodigo.set(produto.codigo);
    this.form.setValue({ codigo: produto.codigo, descricao: produto.descricao, saldo: produto.saldo });
    this.form.controls.codigo.disable();
  }

  cancelarEdicao(): void {
    this.editingCodigo.set(null);
    this.form.reset({ codigo: '', descricao: '', saldo: 0 });
    this.form.controls.codigo.enable();
  }

  salvar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    this.saving.set(true);
    const value = this.form.getRawValue();
    const editingCodigo = this.editingCodigo();

    const request$ = editingCodigo
      ? this.produtoService.update(editingCodigo, { descricao: value.descricao, saldo: value.saldo })
      : this.produtoService.create(value);

    request$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: () => {
        this.saving.set(false);
        this.notification.success(editingCodigo ? 'Produto atualizado com sucesso.' : 'Produto cadastrado com sucesso.');
        this.cancelarEdicao();
        this.carregar();
      },
      error: (err) => {
        this.saving.set(false);
        this.notification.errorFromResponse(err, 'Falha ao salvar produto.');
      },
    });
  }

  excluir(produto: Produto): void {
    const ref = this.dialog.open(ConfirmDialog, {
      data: {
        title: 'Excluir produto',
        message: `Tem certeza que deseja excluir o produto "${produto.descricao}" (${produto.codigo})?`,
        confirmLabel: 'Excluir',
      },
    });
    ref
      .afterClosed()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((confirmed) => {
        if (!confirmed) return;
        this.produtoService.remove(produto.codigo).subscribe({
          next: () => {
            this.notification.success('Produto removido.');
            this.carregar();
          },
          error: (err) => this.notification.errorFromResponse(err, 'Falha ao remover produto.'),
        });
      });
  }
}
