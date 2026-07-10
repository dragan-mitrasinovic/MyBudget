import { Component, OnInit, computed, inject, input, signal } from '@angular/core';
import { HttpErrorResponse } from '@angular/common/http';
import { Router } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatMenuModule } from '@angular/material/menu';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { firstValueFrom } from 'rxjs';

import { Budget } from '../../../core/budgets/budget.models';
import { BudgetService } from '../../../core/budgets/budget.service';
import { Category } from '../../../core/budgets/category.models';
import { CategoryService } from '../../../core/budgets/category.service';
import { Expense } from '../../../core/budgets/expense.models';
import { ExpenseService } from '../../../core/budgets/expense.service';
import { MoneyPipe } from '../../../core/budgets/money.pipe';
import { ConfirmData, ConfirmDialog } from '../../../shared/confirm-dialog/confirm-dialog';
import { CategoryFormDialog } from '../category-form-dialog';
import { ExpenseDialogData, ExpenseFormDialog } from '../expense-form-dialog';

@Component({
  selector: 'app-budget-detail',
  imports: [
    MatButtonModule,
    MatCardModule,
    MatDialogModule,
    MatIconModule,
    MatListModule,
    MatMenuModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
    MoneyPipe,
  ],
  templateUrl: './budget-detail.html',
  styleUrl: './budget-detail.scss',
})
export class BudgetDetail implements OnInit {
  private budgetService = inject(BudgetService);
  private categoryService = inject(CategoryService);
  private expenseService = inject(ExpenseService);
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);
  private router = inject(Router);

  /** Bound from the `:id` route param (withComponentInputBinding). */
  readonly id = input.required<string>();

  readonly budget = signal<Budget | null>(null);
  readonly categories = this.categoryService.categories;
  readonly expenses = this.expenseService.expenses;
  readonly categoryNames = computed(() => new Map(this.categories().map((c) => [c.id, c.name])));
  readonly totalMinor = computed(() =>
    this.expenses().reduce((sum, expense) => sum + expense.amountMinor, 0),
  );
  readonly loading = signal(true);
  readonly notFound = signal(false);

  async ngOnInit(): Promise<void> {
    const budgetId = Number(this.id());
    try {
      const [budget] = await Promise.all([
        this.budgetService.get(budgetId),
        this.categoryService.load(budgetId),
        this.expenseService.load(budgetId),
      ]);
      this.budget.set(budget);
    } catch {
      this.notFound.set(true);
    } finally {
      this.loading.set(false);
    }
  }

  back(): void {
    this.router.navigate(['/budgets']);
  }

  async addCategory(): Promise<void> {
    const name = await firstValueFrom(
      this.dialog
        .open<CategoryFormDialog, Category | null, string>(CategoryFormDialog, { data: null })
        .afterClosed(),
    );
    if (name) {
      await this.runCategoryAction(() => this.categoryService.create(Number(this.id()), name));
    }
  }

  async editCategory(category: Category): Promise<void> {
    const name = await firstValueFrom(
      this.dialog
        .open<CategoryFormDialog, Category | null, string>(CategoryFormDialog, { data: category })
        .afterClosed(),
    );
    if (name && name !== category.name) {
      await this.runCategoryAction(() =>
        this.categoryService.update(Number(this.id()), category.id, name),
      );
    }
  }

  async removeCategory(category: Category): Promise<void> {
    const data: ConfirmData = {
      title: 'Delete category?',
      message: `Expenses in “${category.name}” will be moved to Uncategorized.`,
      confirmText: 'Delete',
    };
    const confirmed = await firstValueFrom(
      this.dialog.open<ConfirmDialog, ConfirmData, boolean>(ConfirmDialog, { data }).afterClosed(),
    );
    if (confirmed) {
      await this.runCategoryAction(() =>
        this.categoryService.remove(Number(this.id()), category.id),
      );
    }
  }

  private async runCategoryAction(action: () => Promise<void>): Promise<void> {
    try {
      await action();
    } catch (err) {
      const conflict = err instanceof HttpErrorResponse && err.status === 409;
      const message = conflict
        ? 'A category with that name already exists.'
        : 'Something went wrong. Please try again.';
      this.snackBar.open(message, 'Dismiss', { duration: 4000 });
    }
  }

  addExpense(): void {
    this.dialog.open<ExpenseFormDialog, ExpenseDialogData, boolean>(ExpenseFormDialog, {
      data: this.expenseDialogData(null),
    });
  }

  editExpense(expense: Expense): void {
    this.dialog.open<ExpenseFormDialog, ExpenseDialogData, boolean>(ExpenseFormDialog, {
      data: this.expenseDialogData(expense),
    });
  }

  async removeExpense(expense: Expense): Promise<void> {
    const data: ConfirmData = {
      title: 'Delete expense?',
      message: 'This expense will be permanently deleted.',
      confirmText: 'Delete',
    };
    const confirmed = await firstValueFrom(
      this.dialog.open<ConfirmDialog, ConfirmData, boolean>(ConfirmDialog, { data }).afterClosed(),
    );
    if (confirmed) {
      try {
        await this.expenseService.remove(Number(this.id()), expense.id);
      } catch {
        this.snackBar.open('Could not delete the expense. Please try again.', 'Dismiss', {
          duration: 4000,
        });
      }
    }
  }

  formatDate(iso: string): string {
    const [y, m, d] = iso.split('-').map(Number);
    return new Date(y, m - 1, d).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }

  private expenseDialogData(expense: Expense | null): ExpenseDialogData {
    return {
      budgetId: Number(this.id()),
      currency: this.budget()?.currency ?? '',
      categories: this.categories(),
      expense,
      defaultCategoryId: null,
    };
  }
}
