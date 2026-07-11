import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';

import { Budget, BudgetInput } from '../../core/budgets/budget.models';

/** Two-decimal currencies offered for now; expand when multi-currency is revisited (ADR-0011). */
const CURRENCIES = ['RSD', 'EUR', 'USD', 'GBP', 'CHF', 'AUD', 'CAD'];

@Component({
  selector: 'app-budget-form-dialog',
  imports: [
    ReactiveFormsModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
  ],
  templateUrl: './budget-form-dialog.html',
  styleUrl: './budget-form-dialog.scss',
})
export class BudgetFormDialog {
  private fb = inject(FormBuilder);
  private ref = inject(MatDialogRef<BudgetFormDialog, BudgetInput>);
  readonly budget = inject<Budget | null>(MAT_DIALOG_DATA);

  readonly currencies = CURRENCIES;
  readonly isEdit = this.budget !== null;

  readonly form = this.fb.group({
    name: this.fb.nonNullable.control(this.budget?.name ?? '', [
      Validators.required,
      Validators.maxLength(100),
    ]),
    currency: this.fb.nonNullable.control(this.budget?.currency ?? 'RSD', [Validators.required]),
  });

  cancel(): void {
    this.ref.close();
  }

  save(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const { name, currency } = this.form.getRawValue();
    const input: BudgetInput = {
      name: name.trim(),
      currency,
      // Preserve the current archived state; it isn't edited from this form.
      archived: this.budget?.archived ?? false,
    };
    this.ref.close(input);
  }
}
