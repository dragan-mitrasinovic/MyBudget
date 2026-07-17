export const dashboardLayoutVersion = 8;
export const categoryExpensesWidgetType = 'category-expenses';
export const categoryBreakdownWidgetType = 'category-breakdown';
export const categoryExpenseDisplays = [
  'number-only',
  'amount-only',
  'title-amount',
  'title-amount-date',
] as const;
export const categoryBreakdownVisualizations = ['pie', 'donut', 'bar', 'treemap'] as const;

export type CategoryExpenseDisplay = (typeof categoryExpenseDisplays)[number];
export type CategoryBreakdownVisualization = (typeof categoryBreakdownVisualizations)[number];

interface DashboardWidgetGeometry {
  id: string;
  x: number;
  y: number;
  cols: number;
  rows: number;
}

export interface CategoryExpensesDashboardWidget extends DashboardWidgetGeometry {
  type: typeof categoryExpensesWidgetType;
  categoryId: number;
  expenseDisplay: CategoryExpenseDisplay;
}

export interface CategoryBreakdownDashboardWidget extends DashboardWidgetGeometry {
  type: typeof categoryBreakdownWidgetType;
  visualization: CategoryBreakdownVisualization;
}

export type DashboardWidget = CategoryExpensesDashboardWidget | CategoryBreakdownDashboardWidget;

export function isCategoryExpensesWidget(
  widget: DashboardWidget,
): widget is CategoryExpensesDashboardWidget {
  return widget.type === categoryExpensesWidgetType;
}

export function isCategoryBreakdownWidget(
  widget: DashboardWidget,
): widget is CategoryBreakdownDashboardWidget {
  return widget.type === categoryBreakdownWidgetType;
}

export interface DashboardLayout {
  version: typeof dashboardLayoutVersion;
  widgets: DashboardWidget[];
}

const dashboardInitialGridColumns = 36;
const categoryBreakdownDefaultColumns = 18;
const categoryBreakdownDefaultRows = 12;
const dashboardMaximumGridRows = 20000;

export function createCategoryBreakdownWidget(
  widgets: readonly DashboardWidget[],
  uuid: string,
  visualization: CategoryBreakdownVisualization,
): CategoryBreakdownDashboardWidget | null {
  const widget: CategoryBreakdownDashboardWidget = {
    id: `${categoryBreakdownWidgetType}:${uuid}`,
    type: categoryBreakdownWidgetType,
    visualization,
    x: 0,
    y: 0,
    cols: categoryBreakdownDefaultColumns,
    rows: categoryBreakdownDefaultRows,
  };

  for (let y = 0; y <= dashboardMaximumGridRows - widget.rows; y++) {
    for (let x = 0; x <= dashboardInitialGridColumns - widget.cols; x++) {
      widget.x = x;
      widget.y = y;
      if (!widgets.some((existing) => dashboardWidgetsOverlap(widget, existing))) {
        return widget;
      }
    }
  }

  return null;
}

function dashboardWidgetsOverlap(first: DashboardWidget, second: DashboardWidget): boolean {
  return (
    first.x < second.x + second.cols &&
    first.x + first.cols > second.x &&
    first.y < second.y + second.rows &&
    first.y + first.rows > second.y
  );
}
