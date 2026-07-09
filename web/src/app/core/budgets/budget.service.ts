import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../environments/environment';
import { Budget, BudgetInput } from './budget.models';

@Injectable({ providedIn: 'root' })
export class BudgetService {
  private http = inject(HttpClient);
  private readonly api = environment.apiBase;

  /** The user's budgets, newest first. Drives change detection (zoneless). */
  readonly budgets = signal<Budget[]>([]);

  /** Refresh the budget list from the API. */
  async load(): Promise<void> {
    const budgets = await firstValueFrom(this.http.get<Budget[]>(`${this.api}/budgets`));
    this.budgets.set(budgets);
  }

  /** Fetch a single budget by id. */
  async get(id: number): Promise<Budget> {
    return firstValueFrom(this.http.get<Budget>(`${this.api}/budgets/${id}`));
  }

  async create(input: BudgetInput): Promise<Budget> {
    const budget = await firstValueFrom(this.http.post<Budget>(`${this.api}/budgets`, input));
    this.budgets.update((list) => [budget, ...list]);
    return budget;
  }

  async update(id: number, input: BudgetInput): Promise<Budget> {
    const budget = await firstValueFrom(
      this.http.patch<Budget>(`${this.api}/budgets/${id}`, input),
    );
    this.budgets.update((list) => list.map((b) => (b.id === id ? budget : b)));
    return budget;
  }

  async remove(id: number): Promise<void> {
    await firstValueFrom(this.http.delete(`${this.api}/budgets/${id}`));
    this.budgets.update((list) => list.filter((b) => b.id !== id));
  }
}
