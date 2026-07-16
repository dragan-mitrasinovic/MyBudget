import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';

import { Category } from '../../core/budgets/category.models';

export interface CategoryDeleteDialogData {
  category: Category;
  expenseCount: number;
  targetCategories: Category[];
}

export interface CategoryDeleteDialogResult {
  reassignToCategoryId: number | null;
}

@Component({
  selector: 'app-category-delete-dialog',
  imports: [
    ReactiveFormsModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatSelectModule,
  ],
  templateUrl: './category-delete-dialog.html',
  styleUrl: './category-delete-dialog.scss',
})
export class CategoryDeleteDialog {
  private fb = inject(FormBuilder);
  private ref = inject(MatDialogRef<CategoryDeleteDialog, CategoryDeleteDialogResult | undefined>);
  readonly data = inject<CategoryDeleteDialogData>(MAT_DIALOG_DATA);

  readonly hasExpenses = this.data.expenseCount > 0;
  readonly hasTargets = this.data.targetCategories.length > 0;
  readonly form = this.fb.group({
    reassignToCategoryId: this.fb.control<number | null>(
      null,
      this.hasExpenses ? [Validators.required] : [],
    ),
  });

  cancel(): void {
    this.ref.close();
  }

  confirm(): void {
    if (this.hasExpenses && (!this.hasTargets || this.form.invalid)) {
      this.form.markAllAsTouched();
      return;
    }

    this.ref.close({
      reassignToCategoryId: this.hasExpenses ? this.form.controls.reassignToCategoryId.value : null,
    });
  }
}
