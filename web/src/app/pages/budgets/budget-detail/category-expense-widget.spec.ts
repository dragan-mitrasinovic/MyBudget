import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Category } from '../../../core/budgets/category.models';
import { CategoryExpenseWidget } from './category-expense-widget';

describe('CategoryExpenseWidget', () => {
  const category: Category = {
    id: 34,
    budgetId: 12,
    name: 'Food',
    colorKey: 'chart-1',
    colorHex: null,
    createdAt: '2026-07-01T00:00:00Z',
  };

  afterEach(() => TestBed.resetTestingModule());

  it('emits distinct category management actions from its menu', async () => {
    const fixture = TestBed.createComponent(CategoryExpenseWidget);
    fixture.componentRef.setInput('category', category);
    fixture.componentRef.setInput('expenses', []);
    fixture.componentRef.setInput('currency', 'EUR');
    const component = fixture.componentInstance;
    const rename = vi.fn();
    const remove = vi.fn();
    component.renameCategoryRequested.subscribe(rename);
    component.deleteCategoryRequested.subscribe(remove);
    fixture.detectChanges();

    await clickCategoryMenuItem(fixture, 'Rename category');
    expect(rename).toHaveBeenCalledWith(category);
    expect(remove).not.toHaveBeenCalled();

    await clickCategoryMenuItem(fixture, 'Delete category');
    expect(remove).toHaveBeenCalledWith(category);
  });
});

async function clickCategoryMenuItem(
  fixture: ComponentFixture<CategoryExpenseWidget>,
  label: string,
): Promise<void> {
  const root = fixture.nativeElement as HTMLElement;
  const trigger = root.querySelector<HTMLButtonElement>('[aria-label="Food category actions"]');
  trigger?.click();
  fixture.detectChanges();
  await fixture.whenStable();

  const item = Array.from(document.querySelectorAll<HTMLButtonElement>('.mat-mdc-menu-item')).find(
    (button) => button.textContent?.includes(label),
  );
  expect(item).toBeDefined();
  item?.click();
  fixture.detectChanges();
  await fixture.whenStable();
}
