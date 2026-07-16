import { Category } from './category.models';
import { aggregateCategorySpending } from './category-spending';
import { Expense } from './expense.models';

describe('aggregateCategorySpending', () => {
  it('sums all-time spending, omits zero totals, and sorts deterministically', () => {
    const categories = [category(1, 'Transport'), category(2, 'Food'), category(3, 'Unused')];
    const expenses = [expense(1, 1, 1200), expense(2, 2, 800), expense(3, 1, 300)];

    expect(aggregateCategorySpending(categories, expenses)).toEqual([
      { category: categories[0], amountMinor: 1500 },
      { category: categories[1], amountMinor: 800 },
    ]);
  });

  it('uses name then ID for equal spending totals', () => {
    const categories = [category(3, 'Food'), category(2, 'Food'), category(1, 'Transport')];
    const expenses = [expense(1, 1, 500), expense(2, 2, 500), expense(3, 3, 500)];

    expect(
      aggregateCategorySpending(categories, expenses).map(({ category }) => category.id),
    ).toEqual([2, 3, 1]);
  });
});

function category(id: number, name: string): Category {
  return {
    id,
    budgetId: 12,
    name,
    colorKey: `chart-${id}`,
    colorHex: null,
    createdAt: '2026-07-01T00:00:00Z',
  };
}

function expense(id: number, categoryId: number, amountMinor: number): Expense {
  return {
    id,
    categoryId,
    amountMinor,
    description: '',
    spentAt: '2026-07-12',
    createdAt: '2026-07-12T00:00:00Z',
    updatedAt: '2026-07-12T00:00:00Z',
  };
}
