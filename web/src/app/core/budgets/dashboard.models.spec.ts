import {
  DashboardWidget,
  categoryExpensesWidgetType,
  createCategoryBreakdownWidget,
} from './dashboard.models';

describe('createCategoryBreakdownWidget', () => {
  it('persists each visualization while placing charts in the first available positions', () => {
    const categoryWidgets: DashboardWidget[] = [0, 1, 2].map((index) => ({
      id: `category-expenses:${index + 1}`,
      type: categoryExpensesWidgetType,
      categoryId: index + 1,
      x: index * 12,
      y: 0,
      cols: 12,
      rows: 10,
    }));

    const first = createCategoryBreakdownWidget(
      categoryWidgets,
      '11111111-1111-4111-8111-111111111111',
      'pie',
    );
    expect(first).toMatchObject({ visualization: 'pie', x: 0, y: 10, cols: 18, rows: 12 });

    const second = createCategoryBreakdownWidget(
      [...categoryWidgets, first!],
      '22222222-2222-4222-8222-222222222222',
      'donut',
    );
    expect(second).toMatchObject({ visualization: 'donut', x: 18, y: 10, cols: 18, rows: 12 });

    const bar = createCategoryBreakdownWidget([], '33333333-3333-4333-8333-333333333333', 'bar');
    const treemap = createCategoryBreakdownWidget(
      [],
      '44444444-4444-4444-8444-444444444444',
      'treemap',
    );
    expect(bar?.visualization).toBe('bar');
    expect(treemap?.visualization).toBe('treemap');
  });
});
