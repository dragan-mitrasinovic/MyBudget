import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../environments/environment';
import { Category, CategoryColorInput } from './category.models';

@Injectable({ providedIn: 'root' })
export class CategoryService {
  private http = inject(HttpClient);
  private readonly api = environment.apiBase;

  /** Categories of the budget currently being viewed (user categories first, then
   * Uncategorized). Drives change detection (zoneless). */
  readonly categories = signal<Category[]>([]);

  /** Load the categories for a budget, replacing the current list. */
  async load(budgetId: number): Promise<void> {
    const categories = await firstValueFrom(
      this.http.get<Category[]>(`${this.api}/budgets/${budgetId}/categories`),
    );
    this.categories.set(categories);
  }

  // Mutations re-fetch so ordering stays authoritative (the server sorts them).
  async create(budgetId: number, name: string): Promise<void> {
    await firstValueFrom(
      this.http.post<Category>(`${this.api}/budgets/${budgetId}/categories`, { name }),
    );
    await this.load(budgetId);
  }

  async update(budgetId: number, categoryId: number, name: string): Promise<void> {
    await firstValueFrom(
      this.http.patch<Category>(`${this.api}/categories/${categoryId}`, { name }),
    );
    await this.load(budgetId);
  }

  async updateColor(
    budgetId: number,
    categoryId: number,
    input: CategoryColorInput,
  ): Promise<void> {
    await firstValueFrom(
      this.http.put<Category>(`${this.api}/categories/${categoryId}/color`, input),
    );
    await this.load(budgetId);
  }

  async remove(budgetId: number, categoryId: number): Promise<void> {
    await firstValueFrom(this.http.delete(`${this.api}/categories/${categoryId}`));
    await this.load(budgetId);
  }
}
