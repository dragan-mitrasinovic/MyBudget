import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';

import { Category } from '../../core/budgets/category.models';
import {
  CategoryDeleteDialog,
  CategoryDeleteDialogData,
  CategoryDeleteDialogResult,
} from './category-delete-dialog';

describe('CategoryDeleteDialog', () => {
  const source = category(34, 'Food');
  const target = category(35, 'Household');

  afterEach(() => TestBed.resetTestingModule());

  it('confirms deletion of an empty category without a target', () => {
    const { component, close } = setup({
      category: source,
      expenseCount: 0,
      targetCategories: [],
    });

    component.confirm();

    expect(close).toHaveBeenCalledWith({ reassignToCategoryId: null });
  });

  it('requires a target before confirming a populated category', () => {
    const { component, fixture, close } = setup({
      category: source,
      expenseCount: 2,
      targetCategories: [target],
    });

    component.confirm();
    fixture.detectChanges();

    expect(close).not.toHaveBeenCalled();
    expect(component.form.controls.reassignToCategoryId.hasError('required')).toBe(true);
    expect(fixture.nativeElement.textContent).toContain('Choose a category');

    component.form.controls.reassignToCategoryId.setValue(target.id);
    component.confirm();

    expect(close).toHaveBeenCalledWith({ reassignToCategoryId: target.id });
  });

  it('blocks deleting a populated category when no target exists', () => {
    const { component, fixture, close } = setup({
      category: source,
      expenseCount: 1,
      targetCategories: [],
    });
    fixture.detectChanges();

    const root = fixture.nativeElement as HTMLElement;
    const deleteButton = root.querySelector<HTMLButtonElement>('.delete-button');
    expect(deleteButton?.disabled).toBe(true);
    expect(root.textContent).toContain('Create another category');

    component.confirm();
    expect(close).not.toHaveBeenCalled();
  });

  function setup(data: CategoryDeleteDialogData): {
    component: CategoryDeleteDialog;
    fixture: ComponentFixture<CategoryDeleteDialog>;
    close: ReturnType<typeof vi.fn<(result?: CategoryDeleteDialogResult | undefined) => void>>;
  } {
    const close = vi.fn<(result?: CategoryDeleteDialogResult | undefined) => void>();

    TestBed.configureTestingModule({
      imports: [CategoryDeleteDialog],
      providers: [
        { provide: MAT_DIALOG_DATA, useValue: data },
        { provide: MatDialogRef, useValue: { close } },
      ],
    });

    const fixture = TestBed.createComponent(CategoryDeleteDialog);
    fixture.detectChanges();
    return { component: fixture.componentInstance, fixture, close };
  }
});

function category(id: number, name: string): Category {
  return {
    id,
    budgetId: 12,
    name,
    colorKey: `chart-${id - 33}`,
    colorHex: null,
    createdAt: '2026-07-01T00:00:00Z',
  };
}
