import { Routes } from '@angular/router';

import { authGuard } from './core/auth/auth.guard';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () => import('./pages/login/login').then((m) => m.Login),
  },
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () => import('./core/shell/shell').then((m) => m.Shell),
    children: [
      { path: '', pathMatch: 'full', redirectTo: 'budgets' },
      {
        path: 'budgets',
        loadComponent: () => import('./pages/budgets/budgets').then((m) => m.Budgets),
      },
      {
        path: 'budgets/:id',
        loadComponent: () =>
          import('./pages/budgets/budget-detail/budget-detail').then((m) => m.BudgetDetail),
      },
    ],
  },
  { path: '**', redirectTo: '' },
];
