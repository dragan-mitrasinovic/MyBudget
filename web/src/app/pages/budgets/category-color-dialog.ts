import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatButtonToggleChange, MatButtonToggleModule } from '@angular/material/button-toggle';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatTooltipModule } from '@angular/material/tooltip';

import {
  CategoryColorKey,
  categoryColorKeys,
  categoryColorVariable,
} from '../../core/budgets/category-colors';
import { Category, CategoryColorInput } from '../../core/budgets/category.models';

type ColorMode = 'palette' | 'custom';

@Component({
  selector: 'app-category-color-dialog',
  imports: [
    ReactiveFormsModule,
    MatButtonModule,
    MatButtonToggleModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatTooltipModule,
  ],
  templateUrl: './category-color-dialog.html',
  styleUrl: './category-color-dialog.scss',
})
export class CategoryColorDialog {
  private fb = inject(FormBuilder);
  private ref = inject(MatDialogRef<CategoryColorDialog, CategoryColorInput>);
  readonly category = inject<Category>(MAT_DIALOG_DATA);

  readonly paletteKeys = categoryColorKeys;
  readonly mode = signal<ColorMode>(this.category.colorHex === null ? 'palette' : 'custom');
  readonly selectedColorKey = signal<CategoryColorKey>(
    categoryColorKeys.includes(this.category.colorKey as CategoryColorKey)
      ? (this.category.colorKey as CategoryColorKey)
      : 'chart-1',
  );
  readonly form = this.fb.group({
    colorHex: this.fb.nonNullable.control(this.category.colorHex ?? '#0072B2', [
      Validators.required,
      Validators.pattern(/^#[0-9A-Fa-f]{6}$/),
    ]),
  });

  colorVariable(colorKey: CategoryColorKey): string {
    return categoryColorVariable(colorKey);
  }

  setMode(change: MatButtonToggleChange): void {
    this.mode.set(change.value as ColorMode);
  }

  selectPalette(colorKey: CategoryColorKey): void {
    this.selectedColorKey.set(colorKey);
    this.mode.set('palette');
  }

  setNativeColor(event: Event): void {
    this.form.controls.colorHex.setValue((event.target as HTMLInputElement).value.toUpperCase());
  }

  cancel(): void {
    this.ref.close();
  }

  save(): void {
    if (this.mode() === 'custom' && this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.ref.close({
      colorKey: this.selectedColorKey(),
      colorHex: this.mode() === 'custom' ? this.form.controls.colorHex.value.toUpperCase() : null,
    });
  }
}
