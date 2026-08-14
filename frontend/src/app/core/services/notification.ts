import { Service, inject } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { MatSnackBar } from '@angular/material/snack-bar';

import { ApiErrorBody } from '../models/api-error.model';

@Service()
export class NotificationService {
  private readonly snackBar = inject(MatSnackBar);

  success(message: string): void {
    this.snackBar.open(message, 'Fechar', { duration: 4000, panelClass: 'snackbar-success' });
  }

  error(message: string): void {
    this.snackBar.open(message, 'Fechar', { duration: 7000, panelClass: 'snackbar-error' });
  }

  errorFromResponse(err: unknown, fallback = 'Ocorreu um erro inesperado. Tente novamente.'): void {
    if (err instanceof HttpErrorResponse) {
      const body = err.error as ApiErrorBody | undefined;
      this.error(body?.error?.message ?? fallback);
      return;
    }
    this.error(fallback);
  }
}
