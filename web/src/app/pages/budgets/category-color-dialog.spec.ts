import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';

import { Category } from '../../core/budgets/category.models';
import { CategoryColorDialog } from './category-color-dialog';

describe('CategoryColorDialog', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('saves a selected palette color without a custom override', () => {
    const { component, fixture, close } = setup();

    component.selectPalette('chart-5');
    fixture.detectChanges();
    const submit = new Event('submit', { cancelable: true });
    fixture.nativeElement.querySelector('form').dispatchEvent(submit);

    expect(submit.defaultPrevented).toBe(true);
    expect(close).toHaveBeenCalledWith({ colorKey: 'chart-5', colorHex: null });
  });

  it('validates and normalizes a custom hex override', () => {
    const { component, close } = setup();
    component.mode.set('custom');
    component.form.controls.colorHex.setValue('#12345');

    component.save();
    expect(close).not.toHaveBeenCalled();

    component.form.controls.colorHex.setValue('#a1b2c3');
    component.save();
    expect(close).toHaveBeenCalledWith({ colorKey: 'chart-1', colorHex: '#A1B2C3' });
  });
});

function setup(): {
  component: CategoryColorDialog;
  fixture: ComponentFixture<CategoryColorDialog>;
  close: ReturnType<typeof vi.fn>;
} {
  const close = vi.fn();
  const category: Category = {
    id: 34,
    budgetId: 12,
    name: 'Food',
    isSystem: false,
    colorKey: 'chart-1',
    colorHex: null,
    createdAt: '2026-07-01T00:00:00Z',
  };

  TestBed.configureTestingModule({
    imports: [CategoryColorDialog],
    providers: [
      { provide: MAT_DIALOG_DATA, useValue: category },
      { provide: MatDialogRef, useValue: { close } },
    ],
  });

  const fixture = TestBed.createComponent(CategoryColorDialog);
  return { component: fixture.componentInstance, fixture, close };
}
