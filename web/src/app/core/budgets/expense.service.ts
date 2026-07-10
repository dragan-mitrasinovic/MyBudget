import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../environments/environment';
import { Expense, ExpenseInput } from './expense.models';

@Injectable({ providedIn: 'root' })
export class ExpenseService {
  private http = inject(HttpClient);
  private readonly api = environment.apiBase;

  /** Expenses of the budget currently being viewed, newest first. */
  readonly expenses = signal<Expense[]>([]);

  /** Load the expenses for a budget, replacing the current list. */
  async load(budgetId: number): Promise<void> {
    const expenses = await firstValueFrom(
      this.http.get<Expense[]>(`${this.api}/budgets/${budgetId}/expenses`),
    );
    this.expenses.set(expenses);
  }

  // Mutations re-fetch so ordering stays authoritative (the server sorts by date).
  async create(budgetId: number, input: ExpenseInput): Promise<void> {
    await firstValueFrom(
      this.http.post<Expense>(`${this.api}/budgets/${budgetId}/expenses`, input),
    );
    await this.load(budgetId);
  }

  async update(budgetId: number, expenseId: number, input: ExpenseInput): Promise<void> {
    await firstValueFrom(this.http.patch<Expense>(`${this.api}/expenses/${expenseId}`, input));
    await this.load(budgetId);
  }

  async remove(budgetId: number, expenseId: number): Promise<void> {
    await firstValueFrom(this.http.delete(`${this.api}/expenses/${expenseId}`));
    await this.load(budgetId);
  }
}
