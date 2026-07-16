import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnDestroy, OnInit, computed, inject, input, signal } from '@angular/core';
import { Router } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import {
  CompactType,
  DisplayGrid,
  Gridster,
  GridsterConfig,
  GridsterItem,
  GridType,
} from 'angular-gridster2';
import { firstValueFrom } from 'rxjs';

import { Budget } from '../../../core/budgets/budget.models';
import { BudgetService } from '../../../core/budgets/budget.service';
import { Category } from '../../../core/budgets/category.models';
import { CategoryService } from '../../../core/budgets/category.service';
import { aggregateCategorySpending } from '../../../core/budgets/category-spending';
import {
  CategoryBreakdownVisualization,
  DashboardLayout,
  DashboardWidget,
  createCategoryBreakdownWidget,
  dashboardLayoutVersion,
  isCategoryBreakdownWidget,
  isCategoryExpensesWidget,
} from '../../../core/budgets/dashboard.models';
import { DashboardService } from '../../../core/budgets/dashboard.service';
import { Expense } from '../../../core/budgets/expense.models';
import { ExpenseService } from '../../../core/budgets/expense.service';
import { ConfirmData, ConfirmDialog } from '../../../shared/confirm-dialog/confirm-dialog';
import { CategoryColorDialog } from '../category-color-dialog';
import {
  CategoryDeleteDialog,
  CategoryDeleteDialogData,
  CategoryDeleteDialogResult,
} from '../category-delete-dialog';
import { CategoryFormDialog } from '../category-form-dialog';
import { ExpenseDialogData, ExpenseFormDialog } from '../expense-form-dialog';
import { CategoryBreakdownWidget } from './category-breakdown-widget';
import { CategoryExpenseWidget } from './category-expense-widget';

type LayoutSaveState = 'saved' | 'saving' | 'error';

const dashboardGridColumnScale = 3;
const dashboardGridRowScale = 2;
const dashboardMaximumGridColumns = 20000;

@Component({
  selector: 'app-budget-detail',
  imports: [
    CategoryBreakdownWidget,
    CategoryExpenseWidget,
    Gridster,
    GridsterItem,
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatMenuModule,
    MatProgressSpinnerModule,
    MatSnackBarModule,
  ],
  templateUrl: './budget-detail.html',
  styleUrl: './budget-detail.scss',
})
export class BudgetDetail implements OnInit, OnDestroy {
  private budgetService = inject(BudgetService);
  private categoryService = inject(CategoryService);
  private dashboardService = inject(DashboardService);
  private expenseService = inject(ExpenseService);
  private dialog = inject(MatDialog);
  private snackBar = inject(MatSnackBar);
  private router = inject(Router);
  private dashboardSaveTimeout: ReturnType<typeof setTimeout> | undefined;
  private isDashboardSaveInFlight = false;
  private hasQueuedDashboardSave = false;

  /** Bound from the `:id` route param (withComponentInputBinding). */
  readonly id = input.required<string>();

  readonly budget = signal<Budget | null>(null);
  readonly categories = this.categoryService.categories;
  readonly expenses = this.expenseService.expenses;
  readonly dashboardWidgets = signal<DashboardWidget[]>([]);
  readonly categoryById = computed(() => new Map(this.categories().map((c) => [c.id, c])));
  readonly expensesByCategory = computed(() => {
    const grouped = new Map<number, Expense[]>();
    for (const expense of this.expenses()) {
      const categoryExpenses = grouped.get(expense.categoryId) ?? [];
      categoryExpenses.push(expense);
      grouped.set(expense.categoryId, categoryExpenses);
    }

    return grouped;
  });
  readonly categorySpending = computed(() =>
    aggregateCategorySpending(this.categories(), this.expenses()),
  );
  readonly loading = signal(true);
  readonly notFound = signal(false);
  readonly layoutSaveState = signal<LayoutSaveState>('saved');
  readonly isCategoryBreakdownWidget = isCategoryBreakdownWidget;
  readonly isCategoryExpensesWidget = isCategoryExpensesWidget;
  readonly gridOptions: GridsterConfig = {
    gridType: GridType.Fixed,
    compactType: CompactType.None,
    minCols: 12 * dashboardGridColumnScale,
    maxCols: dashboardMaximumGridColumns,
    minRows: 1,
    maxRows: 10000 * dashboardGridRowScale,
    fixedColWidth: 32,
    fixedRowHeight: 32,
    margin: 0,
    outerMargin: false,
    mobileBreakpoint: 700,
    keepFixedHeightInMobile: true,
    minItemCols: 3 * dashboardGridColumnScale,
    maxItemCols: 12 * dashboardGridColumnScale,
    minItemRows: 3 * dashboardGridRowScale,
    maxItemRows: 12 * dashboardGridRowScale,
    pushItems: true,
    pushResizeItems: true,
    disablePushOnResize: true,
    displayGrid: DisplayGrid.OnDragAndResize,
    enableBoundaryControl: true,
    draggable: {
      enabled: true,
      ignoreContent: true,
      dragHandleClass: 'dashboard-drag-handle',
      stop: () => this.scheduleDashboardSave(),
    },
    resizable: {
      enabled: true,
      handles: {
        s: true,
        e: true,
        n: true,
        w: true,
        se: true,
        ne: true,
        sw: true,
        nw: true,
      },
      stop: () => this.scheduleDashboardSave(),
    },
  };

  async ngOnInit(): Promise<void> {
    const budgetId = Number(this.id());
    try {
      const [budget, layout] = await Promise.all([
        this.budgetService.get(budgetId),
        this.dashboardService.get(budgetId),
        this.categoryService.load(budgetId),
        this.expenseService.load(budgetId),
      ]);
      this.budget.set(budget);
      this.setDashboardLayout(layout);
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
    const data: CategoryDeleteDialogData = {
      category,
      expenseCount: this.expensesByCategory().get(category.id)?.length ?? 0,
      targetCategories: this.categories().filter((candidate) => candidate.id !== category.id),
    };
    const result = await firstValueFrom(
      this.dialog
        .open<CategoryDeleteDialog, CategoryDeleteDialogData, CategoryDeleteDialogResult>(
          CategoryDeleteDialog,
          { data },
        )
        .afterClosed(),
    );
    if (!result) {
      return;
    }

    const budgetId = Number(this.id());
    try {
      await this.categoryService.remove(budgetId, category.id, result.reassignToCategoryId);
      await Promise.all([this.expenseService.load(budgetId), this.loadDashboardLayout()]);
    } catch (err) {
      await Promise.allSettled([
        this.categoryService.load(budgetId),
        this.expenseService.load(budgetId),
        this.loadDashboardLayout(),
      ]);
      const isStaleCategoryState =
        err instanceof HttpErrorResponse && (err.status === 400 || err.status === 409);
      const message = isStaleCategoryState
        ? 'Categories changed. Review the latest expenses and try again.'
        : 'Could not delete the category. Please try again.';
      this.snackBar.open(message, 'Dismiss', { duration: 4000 });
    }
  }

  private async runCategoryAction(action: () => Promise<void>): Promise<void> {
    try {
      await action();
      await this.loadDashboardLayout();
    } catch (err) {
      const conflict = err instanceof HttpErrorResponse && err.status === 409;
      const message = conflict
        ? 'A category with that name already exists.'
        : 'Something went wrong. Please try again.';
      this.snackBar.open(message, 'Dismiss', { duration: 4000 });
    }
  }

  addExpense(defaultCategoryId: number | null = null): void {
    if (this.categories().length === 0) {
      this.snackBar.open('Create a category before adding an expense.', 'Dismiss', {
        duration: 4000,
      });
      return;
    }

    this.dialog.open<ExpenseFormDialog, ExpenseDialogData, boolean>(ExpenseFormDialog, {
      data: this.expenseDialogData(null, defaultCategoryId),
    });
  }

  editExpense(expense: Expense): void {
    this.dialog.open<ExpenseFormDialog, ExpenseDialogData, boolean>(ExpenseFormDialog, {
      data: this.expenseDialogData(expense, null),
    });
  }

  async removeExpense(expense: Expense): Promise<void> {
    const data: ConfirmData = {
      title: 'Delete expense?',
      message: expense.description
        ? `“${expense.description}” will be permanently deleted.`
        : 'This expense will be permanently deleted.',
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

  addCategoryBreakdownChart(visualization: CategoryBreakdownVisualization): void {
    const widget = createCategoryBreakdownWidget(
      this.dashboardWidgets(),
      globalThis.crypto.randomUUID(),
      visualization,
    );
    if (widget === null) {
      this.snackBar.open('There is no room for another chart.', 'Dismiss', { duration: 4000 });
      return;
    }

    this.dashboardWidgets.update((widgets) => [...widgets, widget]);
    this.scheduleDashboardSave();
  }

  changeCategoryBreakdownVisualization(
    widgetID: string,
    visualization: CategoryBreakdownVisualization,
  ): void {
    this.dashboardWidgets.update((widgets) =>
      widgets.map((widget) =>
        isCategoryBreakdownWidget(widget) && widget.id === widgetID
          ? { ...widget, visualization }
          : widget,
      ),
    );
    this.scheduleDashboardSave();
  }

  removeDashboardWidget(widgetID: string): void {
    this.dashboardWidgets.update((widgets) => widgets.filter((widget) => widget.id !== widgetID));
    this.scheduleDashboardSave();
  }

  async editCategoryColor(category: Category): Promise<void> {
    const input = await firstValueFrom(
      this.dialog.open(CategoryColorDialog, { data: category }).afterClosed(),
    );
    if (!input) {
      return;
    }

    try {
      await this.categoryService.updateColor(Number(this.id()), category.id, input);
    } catch {
      this.snackBar.open('Could not save the category color. Please try again.', 'Dismiss', {
        duration: 4000,
      });
    }
  }

  ngOnDestroy(): void {
    if (this.dashboardSaveTimeout !== undefined) {
      clearTimeout(this.dashboardSaveTimeout);
      this.dashboardSaveTimeout = undefined;
      void this.persistDashboardLayout();
    }
  }

  private expenseDialogData(
    expense: Expense | null,
    defaultCategoryId: number | null,
  ): ExpenseDialogData {
    return {
      budgetId: Number(this.id()),
      currency: this.budget()?.currency ?? '',
      categories: this.categories(),
      expense,
      defaultCategoryId,
    };
  }

  private async loadDashboardLayout(): Promise<void> {
    const layout = await this.dashboardService.get(Number(this.id()));
    this.setDashboardLayout(layout);
  }

  private setDashboardLayout(layout: DashboardLayout): void {
    this.dashboardWidgets.set(layout.widgets.map((widget) => ({ ...widget })));
  }

  private scheduleDashboardSave(): void {
    this.cancelScheduledDashboardSave();
    this.layoutSaveState.set('saving');
    this.dashboardSaveTimeout = setTimeout(() => {
      this.dashboardSaveTimeout = undefined;
      void this.persistDashboardLayout();
    }, 300);
  }

  private cancelScheduledDashboardSave(): void {
    if (this.dashboardSaveTimeout !== undefined) {
      clearTimeout(this.dashboardSaveTimeout);
      this.dashboardSaveTimeout = undefined;
    }
  }

  private async persistDashboardLayout(): Promise<void> {
    if (this.isDashboardSaveInFlight) {
      this.hasQueuedDashboardSave = true;
      return;
    }

    this.isDashboardSaveInFlight = true;
    let didSave = false;
    try {
      await this.dashboardService.save(Number(this.id()), this.dashboardLayoutSnapshot());
      didSave = true;
    } catch {
      this.layoutSaveState.set('error');
      this.snackBar.open('Could not save the dashboard layout. Please try again.', 'Dismiss', {
        duration: 4000,
      });
    } finally {
      this.isDashboardSaveInFlight = false;
      if (this.hasQueuedDashboardSave) {
        this.hasQueuedDashboardSave = false;
        void this.persistDashboardLayout();
      } else if (didSave && this.dashboardSaveTimeout === undefined) {
        this.layoutSaveState.set('saved');
      }
    }
  }

  private dashboardLayoutSnapshot(): DashboardLayout {
    return {
      version: dashboardLayoutVersion,
      widgets: this.dashboardWidgets().map((widget) => {
        const geometry = {
          id: widget.id,
          x: widget.x,
          y: widget.y,
          cols: widget.cols,
          rows: widget.rows,
        };

        return isCategoryExpensesWidget(widget)
          ? { ...geometry, type: widget.type, categoryId: widget.categoryId }
          : { ...geometry, type: widget.type, visualization: widget.visualization };
      }),
    };
  }
}
