import { Component, computed, input, output } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatMenuModule } from '@angular/material/menu';
import { MatTooltipModule } from '@angular/material/tooltip';

import { Category } from '../../../core/budgets/category.models';
import { categoryDisplayColor } from '../../../core/budgets/category-colors';
import { Expense } from '../../../core/budgets/expense.models';
import { MoneyPipe } from '../../../core/budgets/money.pipe';

@Component({
  selector: 'app-category-expense-widget',
  imports: [
    MatButtonModule,
    MatCardModule,
    MatIconModule,
    MatListModule,
    MatMenuModule,
    MatTooltipModule,
    MoneyPipe,
  ],
  templateUrl: './category-expense-widget.html',
  styleUrl: './category-expense-widget.scss',
})
export class CategoryExpenseWidget {
  readonly category = input.required<Category>();
  readonly expenses = input.required<Expense[]>();
  readonly currency = input.required<string>();
  readonly categoryColor = computed(() => categoryDisplayColor(this.category()));

  readonly addRequested = output<number>();
  readonly editRequested = output<Expense>();
  readonly deleteRequested = output<Expense>();
  readonly renameCategoryRequested = output<Category>();
  readonly deleteCategoryRequested = output<Category>();

  readonly totalMinor = computed(() =>
    this.expenses().reduce((total, expense) => total + expense.amountMinor, 0),
  );

  formatDate(iso: string): string {
    const [year, month, day] = iso.split('-').map(Number);
    return new Date(year, month - 1, day).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }
}
