import { Component, DestroyRef, OnInit, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CommonModule, DatePipe } from '@angular/common';
import { Router } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { NotaService } from '../../../core/services/nota';
import { NotificationService } from '../../../core/services/notification';
import { NotaFiscal } from '../../../core/models/nota.model';

@Component({
  selector: 'app-nota-list',
  imports: [CommonModule, DatePipe, MatTableModule, MatButtonModule, MatChipsModule, MatProgressSpinnerModule],
  templateUrl: './nota-list.html',
  styleUrl: './nota-list.scss',
})
export class NotaList implements OnInit {
  private readonly notaService = inject(NotaService);
  private readonly notification = inject(NotificationService);
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);

  readonly displayedColumns = ['numero', 'status', 'createdAt', 'acoes'];
  readonly notas = signal<NotaFiscal[]>([]);
  readonly loading = signal(false);
  readonly creating = signal(false);

  ngOnInit(): void {
    this.carregar();
  }

  carregar(): void {
    this.loading.set(true);
    this.notaService
      .list()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (notas) => {
          this.notas.set(notas);
          this.loading.set(false);
        },
        error: (err) => {
          this.loading.set(false);
          this.notification.errorFromResponse(err, 'Falha ao carregar notas fiscais.');
        },
      });
  }

  novaNota(): void {
    this.creating.set(true);
    this.notaService
      .create()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (nota) => {
          this.creating.set(false);
          this.notification.success(`Nota nº ${nota.numero} criada com sucesso.`);
          this.router.navigate(['/notas', nota.numero]);
        },
        error: (err) => {
          this.creating.set(false);
          this.notification.errorFromResponse(err, 'Falha ao criar nota fiscal.');
        },
      });
  }

  abrir(nota: NotaFiscal): void {
    this.router.navigate(['/notas', nota.numero]);
  }
}
