import { TestBed } from '@angular/core/testing';

import { Category } from '../../../core/budgets/category.models';
import { CategoryBreakdownWidget } from './category-breakdown-widget';

describe('CategoryBreakdownWidget', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('builds shared legend and Pie data from category spending', () => {
    TestBed.configureTestingModule({ imports: [CategoryBreakdownWidget] });
    const fixture = TestBed.createComponent(CategoryBreakdownWidget);
    const component = fixture.componentInstance;
    const food = category(1, 'Food', '#0072B2');
    const transport = category(2, 'Transport', '#E69F00');

    fixture.componentRef.setInput('currency', 'EUR');
    fixture.componentRef.setInput('visualization', 'pie');
    fixture.componentRef.setInput('spending', [
      { category: food, amountMinor: 7500 },
      { category: transport, amountMinor: 2500 },
    ]);

    expect(component.totalMinor()).toBe(10000);
    expect(component.legendItems()).toEqual([
      { category: food, amountMinor: 7500, color: '#0072B2', percentage: 0.75 },
      { category: transport, amountMinor: 2500, color: '#E69F00', percentage: 0.25 },
    ]);
    expect(component.pieData().datasets[0].data).toEqual([7500, 2500]);
    expect(component.chartAriaLabel()).toContain('Pie chart of spending by category');
    expect(component.chartAriaLabel()).toContain('Food: €75.00, 75%');
    expect(component.chartAriaLabel()).toContain('Transport: €25.00, 25%');
  });

  it('builds Donut, horizontal Bar, and Treemap data from the same totals', () => {
    TestBed.configureTestingModule({ imports: [CategoryBreakdownWidget] });
    const fixture = TestBed.createComponent(CategoryBreakdownWidget);
    const component = fixture.componentInstance;
    const food = category(1, 'Food', '#FF4D6D');
    const transport = category(2, 'Transport', '#FF9F1C');
    fixture.componentRef.setInput('currency', 'EUR');
    fixture.componentRef.setInput('visualization', 'donut');
    fixture.componentRef.setInput('spending', [
      { category: food, amountMinor: 7500 },
      { category: transport, amountMinor: 2500 },
    ]);

    expect(component.donutData().datasets[0].data).toEqual([7500, 2500]);
    expect(component.donutOptions.cutout).toBe('58%');
    expect(component.barData().datasets[0].data).toEqual([7500, 2500]);
    expect(component.barOptions().indexAxis).toBe('y');
    expect(component.barChartHeight()).toBe(160);
    expect(component.treemapData().datasets[0].tree).toEqual([
      { name: 'Food', amountMinor: 7500, color: '#FF4D6D' },
      { name: 'Transport', amountMinor: 2500, color: '#FF9F1C' },
    ]);
    expect(component.treemapData().datasets[0].key).toBe('amountMinor');
  });

  it('emits only actual visualization changes', () => {
    TestBed.configureTestingModule({ imports: [CategoryBreakdownWidget] });
    const fixture = TestBed.createComponent(CategoryBreakdownWidget);
    const component = fixture.componentInstance;
    fixture.componentRef.setInput('currency', 'EUR');
    fixture.componentRef.setInput('visualization', 'bar');
    fixture.componentRef.setInput('spending', []);
    const emitted = vi.fn();
    component.visualizationChangeRequested.subscribe(emitted);

    component.selectVisualization('bar');
    component.selectVisualization('treemap');

    expect(emitted).toHaveBeenCalledOnce();
    expect(emitted).toHaveBeenCalledWith('treemap');
  });
});

function category(id: number, name: string, colorHex: string): Category {
  return {
    id,
    budgetId: 12,
    name,
    isSystem: false,
    colorKey: `chart-${id}`,
    colorHex,
    createdAt: '2026-07-01T00:00:00Z',
  };
}
