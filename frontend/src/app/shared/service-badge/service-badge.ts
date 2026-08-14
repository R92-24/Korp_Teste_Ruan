import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';

import { Servico } from '../../core/services/health';

/**
 * Mostra de qual microsserviço a tela depende e se ele está no ar.
 */
@Component({
  selector: 'app-service-badge',
  imports: [CommonModule],
  templateUrl: './service-badge.html',
  styleUrl: './service-badge.scss',
})
export class ServiceBadge {
  @Input({ required: true }) servico!: Servico;
}
