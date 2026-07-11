import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { of } from 'rxjs';

import { Budget } from '../../../core/budgets/budget.models';
import { BudgetService } from '../../../core/budgets/budget.service';
import { Category } from '../../../core/budgets/category.models';
import { CategoryService } from '../../../core/budgets/category.service';
import { Expense } from '../../../core/budgets/expense.models';
import { ExpenseService } from '../../../core/budgets/expense.service';
import { BudgetDetail } from './budget-detail';

describe('BudgetDetail expense descriptions', () => {
  const budget: Budget = {
    id: 12,
    name: 'Household',
    currency: 'EUR',
    archived: false,
    createdAt: '2026-07-01T00:00:00Z',
    updatedAt: '2026-07-01T00:00:00Z',
  };

  const categories: Category[] = [
    {
      id: 34,
      budgetId: budget.id,
      name: 'Eating out',
      isSystem: false,
      createdAt: '2026-07-01T00:00:00Z',
    },
    {
      id: 35,
      budgetId: budget.id,
      name: 'Groceries',
      isSystem: false,
      createdAt: '2026-07-01T00:00:00Z',
    },
  ];

  const expenses: Expense[] = [
    {
      id: 56,
      categoryId: categories[0].id,
      amountMinor: 1599,
      description: 'Burgers',
      spentAt: '2026-07-11',
      createdAt: '2026-07-11T08:32:00Z',
      updatedAt: '2026-07-11T08:32:00Z',
    },
    {
      id: 57,
      categoryId: categories[1].id,
      amountMinor: 4200,
      description: '',
      spentAt: '2026-07-10',
      createdAt: '2026-07-10T08:32:00Z',
      updatedAt: '2026-07-10T08:32:00Z',
    },
  ];

  afterEach(() => TestBed.resetTestingModule());

  it('uses description as the primary label and category as the blank fallback', async () => {
    const { component, fixture } = await setup();
    const root = fixture.nativeElement as HTMLElement;
    const titles = Array.from(root.querySelectorAll<HTMLElement>('.expense-title'), normalizedText);
    const details = Array.from(root.querySelectorAll<HTMLElement>('.expense-sub'), normalizedText);

    expect(titles).toEqual(['Burgers', 'Groceries']);
    expect(details).toEqual([
      `Eating out · ${component.formatDate('2026-07-11')}`,
      component.formatDate('2026-07-10'),
    ]);
  });

  it('identifies described expenses in delete confirmation copy', async () => {
    const { component, dialogMessage } = await setup();

    await component.removeExpense(expenses[0]);
    expect(dialogMessage()).toBe('“Burgers” will be permanently deleted.');

    await component.removeExpense(expenses[1]);
    expect(dialogMessage()).toBe('This expense will be permanently deleted.');
  });

  async function setup(): Promise<{
    component: BudgetDetail;
    fixture: ComponentFixture<BudgetDetail>;
    dialogMessage: () => string;
  }> {
    const categorySignal = signal(categories);
    const expenseSignal = signal(expenses);
    const get = vi.fn<(id: number) => Promise<Budget>>();
    const loadCategories = vi.fn<(budgetId: number) => Promise<void>>();
    const loadExpenses = vi.fn<(budgetId: number) => Promise<void>>();
    let lastDialogConfig: { data: { message: string } } | undefined;
    const dialog = {
      open: (component: unknown, config: unknown) => {
        void component;
        lastDialogConfig = config as { data: { message: string } };
        return { afterClosed: () => of(false) };
      },
    };
    get.mockResolvedValue(budget);
    loadCategories.mockResolvedValue(undefined);
    loadExpenses.mockResolvedValue(undefined);

    TestBed.configureTestingModule({
      imports: [BudgetDetail],
      providers: [
        { provide: BudgetService, useValue: { get } },
        {
          provide: CategoryService,
          useValue: { categories: categorySignal, load: loadCategories },
        },
        {
          provide: ExpenseService,
          useValue: { expenses: expenseSignal, load: loadExpenses, remove: vi.fn() },
        },
        { provide: MatSnackBar, useValue: { open: vi.fn() } },
        { provide: Router, useValue: { navigate: vi.fn() } },
      ],
    });
    TestBed.overrideComponent(BudgetDetail, {
      add: { providers: [{ provide: MatDialog, useValue: dialog }] },
    });
    await TestBed.compileComponents();

    const fixture = TestBed.createComponent(BudgetDetail);
    fixture.componentRef.setInput('id', String(budget.id));
    fixture.detectChanges();
    await vi.waitFor(() => expect(fixture.componentInstance.loading()).toBe(false));
    fixture.detectChanges();

    return {
      component: fixture.componentInstance,
      fixture,
      dialogMessage: () => lastDialogConfig?.data.message ?? '',
    };
  }

  function normalizedText(element: HTMLElement): string {
    return element.textContent?.replace(/\s+/g, ' ').trim() ?? '';
  }
});
