import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';

import { Category } from '../../core/budgets/category.models';
import { Expense, ExpenseInput } from '../../core/budgets/expense.models';
import { ExpenseService } from '../../core/budgets/expense.service';

export interface ExpenseDialogData {
  budgetId: number;
  currency: string;
  categories: Category[];
  expense: Expense | null;
  defaultCategoryId: number | null;
}

/** Parse a YYYY-MM-DD string into a local-midnight Date (avoids UTC off-by-one). */
function parseDate(iso: string): Date {
  const [y, m, d] = iso.split('-').map(Number);
  return new Date(y, m - 1, d);
}

/** Format a Date as a local YYYY-MM-DD string. */
function formatDate(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

@Component({
  selector: 'app-expense-form-dialog',
  imports: [
    ReactiveFormsModule,
    MatButtonModule,
    MatDatepickerModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    MatSnackBarModule,
  ],
  templateUrl: './expense-form-dialog.html',
  styleUrl: './expense-form-dialog.scss',
})
export class ExpenseFormDialog {
  private fb = inject(FormBuilder);
  private ref = inject(MatDialogRef<ExpenseFormDialog, boolean>);
  private expenseService = inject(ExpenseService);
  private snackBar = inject(MatSnackBar);
  readonly data = inject<ExpenseDialogData>(MAT_DIALOG_DATA);

  readonly isEdit = this.data.expense !== null;
  readonly categories = this.data.categories;

  private changed = false;

  readonly form = this.fb.group({
    categoryId: this.fb.nonNullable.control(
      this.data.expense?.categoryId ?? this.data.defaultCategoryId ?? this.categories[0]?.id ?? 0,
      [Validators.required],
    ),
    amount: this.fb.control<number | null>(
      this.data.expense ? this.data.expense.amountMinor / 100 : null,
      [Validators.required, Validators.min(0.01)],
    ),
    date: this.fb.nonNullable.control<Date>(
      this.data.expense ? parseDate(this.data.expense.spentAt) : new Date(),
      [Validators.required],
    ),
    note: this.fb.nonNullable.control(this.data.expense?.note ?? '', [Validators.maxLength(500)]),
  });

  cancel(): void {
    this.ref.close(this.changed);
  }

  async submit(): Promise<void> {
    if (await this.persist()) {
      this.ref.close(true);
    }
  }

  /** Save and keep the dialog open for the next entry (add mode only). */
  async submitAndAddAnother(): Promise<void> {
    if (await this.persist()) {
      this.form.patchValue({ amount: null, note: '' });
      this.form.controls.amount.markAsUntouched();
    }
  }

  private async persist(): Promise<boolean> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return false;
    }

    const { categoryId, amount, date, note } = this.form.getRawValue();
    const input: ExpenseInput = {
      categoryId,
      amountMinor: Math.round((amount ?? 0) * 100),
      note: note.trim(),
      spentAt: formatDate(date),
    };

    try {
      if (this.data.expense) {
        await this.expenseService.update(this.data.budgetId, this.data.expense.id, input);
      } else {
        await this.expenseService.create(this.data.budgetId, input);
      }

      this.changed = true;
      return true;
    } catch {
      this.snackBar.open('Could not save the expense. Please try again.', 'Dismiss', {
        duration: 4000,
      });
      return false;
    }
  }
}
