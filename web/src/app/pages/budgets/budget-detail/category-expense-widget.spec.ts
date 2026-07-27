import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Category } from '../../../core/budgets/category.models';
import { Expense } from '../../../core/budgets/expense.models';
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
  const expense: Expense = {
    id: 56,
    categoryId: category.id,
    amountMinor: 1599,
    description: 'Lunch',
    spentAt: '2026-07-11',
    createdAt: '2026-07-11T08:32:00Z',
    updatedAt: '2026-07-11T08:32:00Z',
  };
  const wholeExpense: Expense = {
    ...expense,
    id: 57,
    amountMinor: 900,
    description: 'Coffee',
  };

  afterEach(() => TestBed.resetTestingModule());

  it('emits distinct category management actions from its menu', async () => {
    const fixture = TestBed.createComponent(CategoryExpenseWidget);
    fixture.componentRef.setInput('category', category);
    fixture.componentRef.setInput('expenses', []);
    fixture.componentRef.setInput('currency', 'EUR');
    fixture.componentRef.setInput('expenseDisplay', 'title-amount');
    const component = fixture.componentInstance;
    const rename = vi.fn();
    const changeColor = vi.fn();
    const remove = vi.fn();
    component.renameCategoryRequested.subscribe(rename);
    component.colorEditRequested.subscribe(changeColor);
    component.deleteCategoryRequested.subscribe(remove);
    fixture.detectChanges();

    await clickCategoryMenuItem(fixture, 'Rename category');
    expect(rename).toHaveBeenCalledWith(category);
    expect(changeColor).not.toHaveBeenCalled();
    expect(remove).not.toHaveBeenCalled();

    await clickCategoryMenuItem(fixture, 'Change color');
    expect(changeColor).toHaveBeenCalledWith(category);
    expect(remove).not.toHaveBeenCalled();

    await clickCategoryMenuItem(fixture, 'Delete category');
    expect(remove).toHaveBeenCalledWith(category);
  });

  it('renders one or two lines according to the selected expense display', () => {
    const fixture = TestBed.createComponent(CategoryExpenseWidget);
    fixture.componentRef.setInput('category', category);
    fixture.componentRef.setInput('expenses', [expense, wholeExpense]);
    fixture.componentRef.setInput('currency', 'EUR');
    fixture.componentRef.setInput('expenseDisplay', 'title-amount');
    fixture.detectChanges();

    expect(text(fixture, '.expense-title')).toBe('Lunch');
    expect(text(fixture, '.expense-amount')).toContain('15.99');
    expect(fixture.nativeElement.querySelector('.expense-sub')).toBeNull();
    expect(fixture.nativeElement.querySelector('.expense-actions')).not.toBeNull();
    expect(fixture.nativeElement.querySelector('.expense-list')?.dataset['display']).toBe(
      'title-amount',
    );
    expect(fixture.nativeElement.querySelector('.expense-row')).not.toBeNull();

    fixture.componentRef.setInput('expenseDisplay', 'number-only');
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('.expense-title')).toBeNull();
    expect(
      Array.from(
        (fixture.nativeElement as HTMLElement).querySelectorAll<HTMLElement>('.expense-number'),
        (element) => element.textContent?.trim(),
      ),
    ).toEqual(['15.99', '9']);
    expect(fixture.nativeElement.querySelector('.expense-sub')).toBeNull();
    expect(fixture.nativeElement.querySelector('.expense-list')?.dataset['display']).toBe(
      'number-only',
    );

    fixture.componentRef.setInput('expenseDisplay', 'amount-only');
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('.expense-title')).toBeNull();
    expect(text(fixture, '.expense-amount')).toContain('15.99');
    expect(fixture.nativeElement.querySelector('.expense-sub')).toBeNull();

    fixture.componentRef.setInput('expenseDisplay', 'title-amount-date');
    fixture.detectChanges();
    expect(text(fixture, '.expense-title')).toBe('Lunch');
    expect(text(fixture, '.expense-sub')).toBe(formatDate(expense.spentAt));
  });

  it('emits only actual expense display changes', () => {
    const fixture = TestBed.createComponent(CategoryExpenseWidget);
    fixture.componentRef.setInput('category', category);
    fixture.componentRef.setInput('expenses', []);
    fixture.componentRef.setInput('currency', 'EUR');
    fixture.componentRef.setInput('expenseDisplay', 'title-amount');
    const emitted = vi.fn();
    fixture.componentInstance.expenseDisplayChangeRequested.subscribe(emitted);

    fixture.componentInstance.selectExpenseDisplay('title-amount');
    fixture.componentInstance.selectExpenseDisplay('number-only');

    expect(emitted).toHaveBeenCalledOnce();
    expect(emitted).toHaveBeenCalledWith('number-only');
  });

  it('clearly marks the active expense display choice', async () => {
    const fixture = TestBed.createComponent(CategoryExpenseWidget);
    fixture.componentRef.setInput('category', category);
    fixture.componentRef.setInput('expenses', []);
    fixture.componentRef.setInput('currency', 'EUR');
    fixture.componentRef.setInput('expenseDisplay', 'title-amount');
    fixture.detectChanges();

    await openExpenseDisplayMenu(fixture);

    const options = Array.from(
      document.querySelectorAll<HTMLButtonElement>('[role="menuitemradio"]'),
    );
    const selectedOptions = options.filter((option) => option.ariaChecked === 'true');
    expect(options).toHaveLength(4);
    expect(selectedOptions).toHaveLength(1);
    expect(selectedOptions[0].textContent).toContain('Titles and amounts');
    expect(selectedOptions[0].classList).toContain('selected-display-option');
    expect(selectedOptions[0].querySelector('mat-icon')?.textContent?.trim()).toBe('check_circle');
  });
});

function text(fixture: ComponentFixture<CategoryExpenseWidget>, selector: string): string {
  return (fixture.nativeElement as HTMLElement).querySelector(selector)?.textContent?.trim() ?? '';
}

function formatDate(iso: string): string {
  const [year, month, day] = iso.split('-').map(Number);
  return new Date(year, month - 1, day).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

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

async function openExpenseDisplayMenu(
  fixture: ComponentFixture<CategoryExpenseWidget>,
): Promise<void> {
  const root = fixture.nativeElement as HTMLElement;
  root.querySelector<HTMLButtonElement>('[aria-label="Food category actions"]')?.click();
  fixture.detectChanges();
  await fixture.whenStable();

  const displayMenuItem = Array.from(
    document.querySelectorAll<HTMLButtonElement>('.mat-mdc-menu-item'),
  ).find((button) => button.textContent?.includes('Expense display'));
  expect(displayMenuItem).toBeDefined();
  displayMenuItem?.click();
  fixture.detectChanges();
  await fixture.whenStable();
}
