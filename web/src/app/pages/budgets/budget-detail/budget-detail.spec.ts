import { HttpErrorResponse } from '@angular/common/http';
import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { MatDialog } from '@angular/material/dialog';
import { MatSnackBar } from '@angular/material/snack-bar';
import { GridType } from 'angular-gridster2';
import { of } from 'rxjs';

import { Budget } from '../../../core/budgets/budget.models';
import { BudgetService } from '../../../core/budgets/budget.service';
import { Category } from '../../../core/budgets/category.models';
import { CategoryService } from '../../../core/budgets/category.service';
import { DashboardLayout, dashboardLayoutVersion } from '../../../core/budgets/dashboard.models';
import { DashboardService } from '../../../core/budgets/dashboard.service';
import { Expense } from '../../../core/budgets/expense.models';
import { ExpenseService } from '../../../core/budgets/expense.service';
import {
  CategoryDeleteDialog,
  CategoryDeleteDialogData,
  CategoryDeleteDialogResult,
} from '../category-delete-dialog';
import { CategoryFormDialog } from '../category-form-dialog';
import { BudgetDetail } from './budget-detail';

describe('BudgetDetail dashboard', () => {
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
      colorKey: 'chart-1',
      colorHex: null,
      createdAt: '2026-07-01T00:00:00Z',
    },
    {
      id: 35,
      budgetId: budget.id,
      name: 'Groceries',
      colorKey: 'chart-2',
      colorHex: null,
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

  const layout: DashboardLayout = {
    version: dashboardLayoutVersion,
    widgets: categories.map((category, index) => ({
      id: `category-expenses:${category.id}`,
      type: 'category-expenses',
      categoryId: category.id,
      x: index * 12,
      y: 0,
      cols: 12,
      rows: 10,
    })),
  };

  afterEach(() => {
    vi.useRealTimers();
    TestBed.resetTestingModule();
  });

  it('uses description as the primary label and category as the blank fallback', async () => {
    const { fixture } = await setup();
    const root = fixture.nativeElement as HTMLElement;
    const titles = Array.from(root.querySelectorAll<HTMLElement>('.expense-title'), normalizedText);
    const details = Array.from(root.querySelectorAll<HTMLElement>('.expense-sub'), normalizedText);

    expect(titles).toEqual(['Burgers', 'Groceries']);
    expect(details).toEqual([formatDate('2026-07-11'), formatDate('2026-07-10')]);
    expect(Array.from(root.querySelectorAll<HTMLElement>('.widget-title'), normalizedText)).toEqual(
      ['Eating out', 'Groceries'],
    );
  });

  it('identifies described expenses in delete confirmation copy', async () => {
    const { component, dialogMessage } = await setup();

    await component.removeExpense(expenses[0]);
    expect(dialogMessage()).toBe('“Burgers” will be permanently deleted.');

    await component.removeExpense(expenses[1]);
    expect(dialogMessage()).toBe('This expense will be permanently deleted.');
  });

  it('adds a category from the budget header', async () => {
    const { fixture, createCategory } = await setup({ categoryDialogResult: 'Utilities' });
    const root = fixture.nativeElement as HTMLElement;
    const addCategoryButton = Array.from(root.querySelectorAll<HTMLButtonElement>('button')).find(
      (button) => normalizedText(button).includes('Add category'),
    );

    expect(addCategoryButton).toBeDefined();
    addCategoryButton!.click();

    await vi.waitFor(() => expect(createCategory).toHaveBeenCalledWith(budget.id, 'Utilities'));
  });

  it('guides an empty budget to its first category and disables expense creation', async () => {
    const emptyLayout: DashboardLayout = { version: dashboardLayoutVersion, widgets: [] };
    const { fixture } = await setup({ categories: [], expenses: [], layout: emptyLayout });
    const root = fixture.nativeElement as HTMLElement;
    const addExpenseButton = Array.from(root.querySelectorAll<HTMLButtonElement>('button')).find(
      (button) => normalizedText(button) === 'add Add expense',
    );

    expect(root.textContent).toContain('Create your first category');
    expect(root.textContent).toContain('Expenses need a category');
    expect(addExpenseButton?.disabled).toBe(true);
    expect(root.querySelector('gridster')).toBeNull();
  });

  it('opens global expense entry without a preselected category', async () => {
    const { component, dialogData } = await setup();

    component.addExpense();

    expect(dialogData()).toMatchObject({ defaultCategoryId: null });
  });

  it('reassigns a category expenses and refreshes expenses and layout', async () => {
    const { component, deleteCategory, loadExpenses, getDashboard, dialogData } = await setup({
      categoryDeleteDialogResult: { reassignToCategoryId: categories[1].id },
    });

    await component.removeCategory(categories[0]);

    expect(dialogData()).toEqual({
      category: categories[0],
      expenseCount: 1,
      targetCategories: [categories[1]],
    } satisfies CategoryDeleteDialogData);
    expect(deleteCategory).toHaveBeenCalledWith(budget.id, categories[0].id, categories[1].id);
    expect(loadExpenses).toHaveBeenCalledTimes(2);
    expect(getDashboard).toHaveBeenCalledTimes(2);
  });

  it('reloads dashboard data when deletion is rejected because category state changed', async () => {
    const { component, loadCategories, loadExpenses, getDashboard, snackBarOpen } = await setup({
      categoryDeleteDialogResult: { reassignToCategoryId: categories[1].id },
      deleteCategoryError: new HttpErrorResponse({ status: 409 }),
    });

    await component.removeCategory(categories[0]);

    expect(loadCategories).toHaveBeenCalledTimes(2);
    expect(loadExpenses).toHaveBeenCalledTimes(2);
    expect(getDashboard).toHaveBeenCalledTimes(2);
    expect(snackBarOpen).toHaveBeenCalledWith(
      'Categories changed. Review the latest expenses and try again.',
      'Dismiss',
      { duration: 4000 },
    );
  });

  it('persists changed widget geometry after dragging stops', async () => {
    const { component, saveDashboard } = await setup();
    vi.useFakeTimers();
    const widget = component.dashboardWidgets()[0];
    widget.x = 48;
    widget.y = 12;

    component.gridOptions.draggable?.stop?.(widget, null as never, new MouseEvent('mouseup'));

    expect(saveDashboard).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(300);

    expect(saveDashboard).toHaveBeenCalledWith(12, {
      version: dashboardLayoutVersion,
      widgets: [{ ...layout.widgets[0], x: 48, y: 12 }, layout.widgets[1]],
    });
    expect(component.layoutSaveState()).toBe('saved');
  });

  it('uses a bidirectionally extensible fixed-cell grid without changing widget size', async () => {
    const { component } = await setup();

    expect(component.gridOptions.gridType).toBe(GridType.Fixed);
    expect(component.gridOptions.minCols).toBe(36);
    expect(component.gridOptions.maxCols).toBe(20000);
    expect(component.gridOptions.fixedColWidth).toBe(32);
    expect(component.gridOptions.fixedRowHeight).toBe(32);
    expect(component.gridOptions.minItemCols).toBe(9);
    expect(component.gridOptions.minItemRows).toBe(6);
    expect(component.dashboardWidgets()[0]).toMatchObject({ cols: 12, rows: 10 });
  });

  it('resizes from edges and corners and shrinks adjacent widgets instead of pushing them', async () => {
    const { component } = await setup();

    expect(component.gridOptions.resizable?.handles).toEqual({
      n: true,
      e: true,
      s: true,
      w: true,
      ne: true,
      nw: true,
      se: true,
      sw: true,
    });
    expect(component.gridOptions.pushResizeItems).toBe(true);
    expect(component.gridOptions.disablePushOnResize).toBe(true);
  });

  it('adds and removes duplicate category breakdown charts through layout persistence', async () => {
    const { component, saveDashboard } = await setup();
    vi.useFakeTimers();

    component.addCategoryBreakdownChart('pie');
    component.addCategoryBreakdownChart('donut');
    component.addCategoryBreakdownChart('bar');
    component.addCategoryBreakdownChart('treemap');
    const charts = component
      .dashboardWidgets()
      .filter((widget) => widget.type === 'category-breakdown');

    expect(charts).toHaveLength(4);
    expect(charts[0]).toMatchObject({ visualization: 'pie', x: 0, y: 10, cols: 18, rows: 12 });
    expect(charts[1]).toMatchObject({
      visualization: 'donut',
      x: 18,
      y: 10,
      cols: 18,
      rows: 12,
    });
    expect(charts[2]).toMatchObject({ visualization: 'bar', x: 0, y: 22, cols: 18, rows: 12 });
    expect(charts[3]).toMatchObject({
      visualization: 'treemap',
      x: 18,
      y: 22,
      cols: 18,
      rows: 12,
    });

    await vi.advanceTimersByTimeAsync(300);
    expect(saveDashboard).toHaveBeenLastCalledWith(
      budget.id,
      expect.objectContaining({ widgets: expect.arrayContaining(charts) }),
    );

    component.removeDashboardWidget(charts[0].id);
    await vi.advanceTimersByTimeAsync(300);
    expect(component.dashboardWidgets()).not.toContainEqual(charts[0]);
  });

  it('changes chart visualization without changing its identity or geometry', async () => {
    const { component, saveDashboard } = await setup();
    vi.useFakeTimers();
    component.addCategoryBreakdownChart('pie');
    const chart = component
      .dashboardWidgets()
      .find((widget) => widget.type === 'category-breakdown');
    expect(chart).toBeDefined();

    component.changeCategoryBreakdownVisualization(chart!.id, 'treemap');
    await vi.advanceTimersByTimeAsync(300);

    const changed = component.dashboardWidgets().find((widget) => widget.id === chart!.id);
    expect(changed).toEqual({ ...chart, visualization: 'treemap' });
    expect(saveDashboard).toHaveBeenLastCalledWith(
      budget.id,
      expect.objectContaining({ widgets: expect.arrayContaining([changed]) }),
    );
  });

  async function setup(
    options: {
      categoryDialogResult?: string;
      categoryDeleteDialogResult?: CategoryDeleteDialogResult;
      categories?: Category[];
      expenses?: Expense[];
      layout?: DashboardLayout;
      deleteCategoryError?: unknown;
    } = {},
  ): Promise<{
    component: BudgetDetail;
    fixture: ComponentFixture<BudgetDetail>;
    dialogMessage: () => string;
    dialogData: () => unknown;
    createCategory: ReturnType<typeof vi.fn<(budgetId: number, name: string) => Promise<void>>>;
    deleteCategory: ReturnType<
      typeof vi.fn<
        (
          budgetId: number,
          categoryId: number,
          reassignToCategoryId?: number | null,
        ) => Promise<void>
      >
    >;
    loadCategories: ReturnType<typeof vi.fn<(budgetId: number) => Promise<void>>>;
    loadExpenses: ReturnType<typeof vi.fn<(budgetId: number) => Promise<void>>>;
    getDashboard: ReturnType<typeof vi.fn<(budgetId: number) => Promise<DashboardLayout>>>;
    snackBarOpen: ReturnType<typeof vi.fn>;
    saveDashboard: ReturnType<
      typeof vi.fn<(budgetId: number, value: DashboardLayout) => Promise<DashboardLayout>>
    >;
  }> {
    const categorySignal = signal(options.categories ?? categories);
    const expenseSignal = signal(options.expenses ?? expenses);
    const get = vi.fn<(id: number) => Promise<Budget>>();
    const loadCategories = vi.fn<(budgetId: number) => Promise<void>>();
    const loadExpenses = vi.fn<(budgetId: number) => Promise<void>>();
    const getDashboard = vi.fn<(budgetId: number) => Promise<DashboardLayout>>();
    const createCategory = vi.fn<(budgetId: number, name: string) => Promise<void>>();
    const updateCategory =
      vi.fn<(budgetId: number, categoryId: number, name: string) => Promise<void>>();
    const deleteCategory =
      vi.fn<
        (
          budgetId: number,
          categoryId: number,
          reassignToCategoryId?: number | null,
        ) => Promise<void>
      >();
    const saveDashboard =
      vi.fn<(budgetId: number, value: DashboardLayout) => Promise<DashboardLayout>>();
    const updateColor = vi.fn();
    const snackBarOpen = vi.fn();
    let lastDialogConfig: { data: unknown } | undefined;
    const dialog = {
      open: (component: unknown, config: unknown) => {
        lastDialogConfig = config as { data: unknown };
        let result: string | CategoryDeleteDialogResult | boolean = false;
        if (component === CategoryFormDialog) {
          result = options.categoryDialogResult ?? false;
        } else if (component === CategoryDeleteDialog) {
          result = options.categoryDeleteDialogResult ?? false;
        }

        return { afterClosed: () => of(result) };
      },
    };
    get.mockResolvedValue(budget);
    loadCategories.mockResolvedValue(undefined);
    loadExpenses.mockResolvedValue(undefined);
    getDashboard.mockResolvedValue(options.layout ?? layout);
    createCategory.mockResolvedValue(undefined);
    updateCategory.mockResolvedValue(undefined);
    if (options.deleteCategoryError) {
      deleteCategory.mockRejectedValue(options.deleteCategoryError);
    } else {
      deleteCategory.mockResolvedValue(undefined);
    }
    saveDashboard.mockResolvedValue(layout);

    TestBed.configureTestingModule({
      imports: [BudgetDetail],
      providers: [
        { provide: BudgetService, useValue: { get } },
        {
          provide: CategoryService,
          useValue: {
            categories: categorySignal,
            load: loadCategories,
            create: createCategory,
            update: updateCategory,
            remove: deleteCategory,
            updateColor,
          },
        },
        {
          provide: ExpenseService,
          useValue: { expenses: expenseSignal, load: loadExpenses, remove: vi.fn() },
        },
        { provide: DashboardService, useValue: { get: getDashboard, save: saveDashboard } },
        { provide: MatSnackBar, useValue: { open: snackBarOpen } },
        { provide: Router, useValue: { navigate: vi.fn() } },
      ],
    });
    TestBed.overrideComponent(BudgetDetail, {
      add: {
        providers: [
          { provide: MatDialog, useValue: dialog },
          { provide: MatSnackBar, useValue: { open: snackBarOpen } },
        ],
      },
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
      dialogMessage: () =>
        (lastDialogConfig?.data as { message?: string } | undefined)?.message ?? '',
      dialogData: () => lastDialogConfig?.data,
      createCategory,
      deleteCategory,
      loadCategories,
      loadExpenses,
      getDashboard,
      snackBarOpen,
      saveDashboard,
    };
  }

  function normalizedText(element: HTMLElement): string {
    return element.textContent?.replace(/\s+/g, ' ').trim() ?? '';
  }

  function formatDate(iso: string): string {
    const [year, month, day] = iso.split('-').map(Number);
    return new Date(year, month - 1, day).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }
});
