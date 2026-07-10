import { Component, inject } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';

import { Category } from '../../core/budgets/category.models';

@Component({
  selector: 'app-category-form-dialog',
  imports: [
    ReactiveFormsModule,
    MatButtonModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
  ],
  templateUrl: './category-form-dialog.html',
  styleUrl: './category-form-dialog.scss',
})
export class CategoryFormDialog {
  private fb = inject(FormBuilder);
  private ref = inject(MatDialogRef<CategoryFormDialog, string>);
  readonly category = inject<Category | null>(MAT_DIALOG_DATA);

  readonly isEdit = this.category !== null;

  readonly form = this.fb.group({
    name: this.fb.nonNullable.control(this.category?.name ?? '', [
      Validators.required,
      Validators.maxLength(100),
    ]),
  });

  cancel(): void {
    this.ref.close();
  }

  save(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.ref.close(this.form.getRawValue().name.trim());
  }
}
