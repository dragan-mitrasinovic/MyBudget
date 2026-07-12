import { Category } from './category.models';
import { Expense } from './expense.models';

export interface CategorySpending {
  category: Category;
  amountMinor: number;
}

export function aggregateCategorySpending(
  categories: readonly Category[],
  expenses: readonly Expense[],
): CategorySpending[] {
  const totals = new Map<number, number>();
  for (const expense of expenses) {
    totals.set(expense.categoryId, (totals.get(expense.categoryId) ?? 0) + expense.amountMinor);
  }

  return categories
    .map((category) => ({ category, amountMinor: totals.get(category.id) ?? 0 }))
    .filter(({ amountMinor }) => amountMinor > 0)
    .sort((first, second) => {
      const amountDifference = second.amountMinor - first.amountMinor;
      if (amountDifference !== 0) {
        return amountDifference;
      }

      if (first.category.name !== second.category.name) {
        return first.category.name < second.category.name ? -1 : 1;
      }

      return first.category.id - second.category.id;
    });
}
