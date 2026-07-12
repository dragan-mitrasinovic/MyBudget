export const dashboardLayoutVersion = 3;
export const categoryExpensesWidgetType = 'category-expenses';

export interface DashboardWidget {
  id: string;
  type: typeof categoryExpensesWidgetType;
  categoryId: number;
  x: number;
  y: number;
  cols: number;
  rows: number;
}

export interface DashboardLayout {
  version: typeof dashboardLayoutVersion;
  widgets: DashboardWidget[];
}
