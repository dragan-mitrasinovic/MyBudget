import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideNativeDateAdapter } from '@angular/material/core';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';

import { Category } from '../../core/budgets/category.models';
import { Expense, ExpenseInput } from '../../core/budgets/expense.models';
import { ExpenseService } from '../../core/budgets/expense.service';
import { ExpenseDialogData, ExpenseFormDialog } from './expense-form-dialog';

describe('ExpenseFormDialog', () => {
  const category: Category = {
    id: 34,
    budgetId: 12,
    name: 'Eating out',
    isSystem: false,
    createdAt: '2026-07-01T00:00:00Z',
  };

  const existingExpense: Expense = {
    id: 56,
    categoryId: category.id,
    amountMinor: 1599,
    description: 'Burgers',
    spentAt: '2026-07-11',
    createdAt: '2026-07-11T08:32:00Z',
    updatedAt: '2026-07-11T08:32:00Z',
  };

  afterEach(() => TestBed.resetTestingModule());

  it('keeps description optional and counts Unicode code points', async () => {
    const { component, fixture } = await setup();
    const control = component.form.controls.description;

    expect(control.valid).toBe(true);

    control.setValue('🍔'.repeat(120));
    expect(control.valid).toBe(true);

    control.setValue('🍔'.repeat(121));
    control.markAsTouched();
    expect(control.hasError('maxlength')).toBe(true);
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain(
      'Description must be at most 120 characters',
    );
  });

  it('trims and submits a new description', async () => {
    const { component, create, close } = await setup();
    component.form.patchValue({
      amount: 15.99,
      description: '  Burgers  ',
      date: new Date(2026, 6, 11),
    });

    await component.submit();

    expect(create).toHaveBeenCalledWith(12, {
      categoryId: category.id,
      amountMinor: 1599,
      description: 'Burgers',
      spentAt: '2026-07-11',
    });
    expect(close).toHaveBeenCalledWith(true);
  });

  it('prefills and updates an existing description', async () => {
    const { component, update } = await setup(existingExpense);

    expect(component.form.controls.description.value).toBe('Burgers');
    component.form.controls.description.setValue('Dinner');

    await component.submit();

    expect(update).toHaveBeenCalledWith(12, existingExpense.id, {
      categoryId: category.id,
      amountMinor: 1599,
      description: 'Dinner',
      spentAt: '2026-07-11',
    });
  });

  it('clears amount and description after saving another expense', async () => {
    const { component, create, close } = await setup();
    const date = new Date(2026, 6, 11);
    component.form.patchValue({ amount: 15.99, description: 'Burgers', date });

    await component.submitAndAddAnother();

    expect(create).toHaveBeenCalledOnce();
    expect(component.form.getRawValue()).toEqual({
      categoryId: category.id,
      amount: null,
      date,
      description: '',
    });
    expect(close).not.toHaveBeenCalled();
  });

  async function setup(expense: Expense | null = null): Promise<{
    fixture: ComponentFixture<ExpenseFormDialog>;
    component: ExpenseFormDialog;
    create: ReturnType<typeof vi.fn<(budgetId: number, input: ExpenseInput) => Promise<void>>>;
    update: ReturnType<
      typeof vi.fn<(budgetId: number, expenseId: number, input: ExpenseInput) => Promise<void>>
    >;
    close: ReturnType<typeof vi.fn<(result?: boolean) => void>>;
  }> {
    const create = vi.fn<(budgetId: number, input: ExpenseInput) => Promise<void>>();
    const update =
      vi.fn<(budgetId: number, expenseId: number, input: ExpenseInput) => Promise<void>>();
    const close = vi.fn<(result?: boolean) => void>();
    create.mockResolvedValue(undefined);
    update.mockResolvedValue(undefined);

    const data: ExpenseDialogData = {
      budgetId: 12,
      currency: 'EUR',
      categories: [category],
      expense,
      defaultCategoryId: category.id,
    };

    await TestBed.configureTestingModule({
      imports: [ExpenseFormDialog],
      providers: [
        provideNativeDateAdapter(),
        { provide: MAT_DIALOG_DATA, useValue: data },
        { provide: MatDialogRef, useValue: { close } },
        { provide: ExpenseService, useValue: { create, update } },
        { provide: MatSnackBar, useValue: { open: vi.fn() } },
      ],
    }).compileComponents();

    const fixture = TestBed.createComponent(ExpenseFormDialog);
    fixture.detectChanges();
    return { fixture, component: fixture.componentInstance, create, update, close };
  }
});
