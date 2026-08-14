import { Routes } from '@angular/router';

import { ProdutoList } from './features/produtos/produto-list/produto-list';
import { NotaList } from './features/notas/nota-list/nota-list';
import { NotaDetail } from './features/notas/nota-detail/nota-detail';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'notas' },
  { path: 'produtos', component: ProdutoList },
  { path: 'notas', component: NotaList },
  { path: 'notas/:numero', component: NotaDetail },
  { path: '**', redirectTo: 'notas' },
];
