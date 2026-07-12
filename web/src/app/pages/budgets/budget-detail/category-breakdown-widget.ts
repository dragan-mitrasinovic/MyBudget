import { DOCUMENT } from '@angular/common';
import { Component, computed, inject, input, output } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatTooltipModule } from '@angular/material/tooltip';
import {
  ArcElement,
  BarController,
  BarElement,
  CategoryScale,
  ChartComponentLike,
  ChartData,
  ChartOptions,
  DoughnutController,
  LinearScale,
  PieController,
  Tooltip,
} from 'chart.js';
import { TreemapController, TreemapDataPoint, TreemapElement } from 'chartjs-chart-treemap';
import { BaseChartDirective, provideCharts } from 'ng2-charts';

import { Category } from '../../../core/budgets/category.models';
import { CategorySpending } from '../../../core/budgets/category-spending';
import { CategoryBreakdownVisualization } from '../../../core/budgets/dashboard.models';
import { MoneyPipe } from '../../../core/budgets/money.pipe';
import { AppThemeService } from '../../../core/theme/app-theme.service';
import { ThemeService } from '../../../core/theme/theme.service';

interface CategoryBreakdownLegendItem extends CategorySpending {
  color: string;
  percentage: number;
}

interface CategoryTreemapSource extends Record<string, unknown> {
  name: string;
  amountMinor: number;
  color: string;
}

@Component({
  selector: 'app-category-breakdown-widget',
  imports: [
    BaseChartDirective,
    MatButtonModule,
    MatCardModule,
    MatIconModule,
    MatMenuModule,
    MatTooltipModule,
    MoneyPipe,
  ],
  providers: [
    provideCharts({
      registerables: [
        PieController,
        DoughnutController,
        BarController,
        ArcElement,
        BarElement,
        CategoryScale,
        LinearScale,
        Tooltip,
        TreemapController as unknown as ChartComponentLike,
        TreemapElement as unknown as ChartComponentLike,
      ],
    }),
  ],
  templateUrl: './category-breakdown-widget.html',
  styleUrl: './category-breakdown-widget.scss',
})
export class CategoryBreakdownWidget {
  private document = inject(DOCUMENT);
  private appTheme = inject(AppThemeService);
  private theme = inject(ThemeService);

  readonly spending = input.required<readonly CategorySpending[]>();
  readonly currency = input.required<string>();
  readonly visualization = input.required<CategoryBreakdownVisualization>();
  readonly colorEditRequested = output<Category>();
  readonly visualizationChangeRequested = output<CategoryBreakdownVisualization>();
  readonly removeRequested = output<void>();

  readonly totalMinor = computed(() =>
    this.spending().reduce((total, item) => total + item.amountMinor, 0),
  );
  readonly legendItems = computed<CategoryBreakdownLegendItem[]>(() => {
    const totalMinor = this.totalMinor();
    return this.spending().map((item) => ({
      ...item,
      color: this.resolveCategoryColor(item.category),
      percentage: totalMinor === 0 ? 0 : item.amountMinor / totalMinor,
    }));
  });
  readonly pieData = computed<ChartData<'pie'>>(() => ({
    labels: this.legendItems().map(({ category }) => category.name),
    datasets: [this.arcDataset()],
  }));
  readonly donutData = computed<ChartData<'doughnut'>>(() => ({
    labels: this.legendItems().map(({ category }) => category.name),
    datasets: [this.arcDataset()],
  }));
  readonly barData = computed<ChartData<'bar'>>(() => ({
    labels: this.legendItems().map(({ category }) => category.name),
    datasets: [
      {
        data: this.legendItems().map(({ amountMinor }) => amountMinor),
        backgroundColor: this.legendItems().map(({ color }) => color),
        borderRadius: 4,
        borderSkipped: false,
      },
    ],
  }));
  readonly treemapData = computed<ChartData<'treemap'>>(() => ({
    datasets: [
      {
        label: 'Category spending',
        data: [],
        tree: this.legendItems().map<CategoryTreemapSource>(({ category, amountMinor, color }) => ({
          name: category.name,
          amountMinor,
          color,
        })),
        key: 'amountMinor',
        backgroundColor: (context) => this.treemapSource(context.raw)?.color ?? 'transparent',
        borderColor: this.resolveTokenColor('--mat-sys-surface-container-lowest'),
        borderWidth: 2,
        spacing: 1,
        labels: { display: false },
      },
    ],
  }));
  readonly visualizationName = computed(() => {
    switch (this.visualization()) {
      case 'pie':
        return 'Pie chart';
      case 'donut':
        return 'Donut chart';
      case 'bar':
        return 'Horizontal bar chart';
      case 'treemap':
        return 'Treemap';
    }
  });
  readonly chartAriaLabel = computed(() => {
    const details = this.legendItems()
      .map(
        ({ category, amountMinor, percentage }) =>
          `${category.name}: ${this.formatMoney(amountMinor)}, ${this.formatPercentage(percentage)}`,
      )
      .join('; ');
    return `${this.visualizationName()} of spending by category. ${details}`;
  });
  readonly pieOptions: ChartOptions<'pie'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          label: (context) => this.formatArcTooltip(context.label, context.raw),
        },
      },
    },
  };
  readonly donutOptions: ChartOptions<'doughnut'> = {
    responsive: true,
    maintainAspectRatio: false,
    cutout: '58%',
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          label: (context) => this.formatArcTooltip(context.label, context.raw),
        },
      },
    },
  };
  readonly barOptions = computed<ChartOptions<'bar'>>(() => {
    const textColor = this.resolveTokenColor('--mat-sys-on-surface-variant');
    const gridColor = this.resolveTokenColor('--mat-sys-outline-variant');
    return {
      indexAxis: 'y',
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: {
          beginAtZero: true,
          grid: { color: gridColor },
          ticks: {
            color: textColor,
            callback: (value) => this.formatMoney(Number(value)),
          },
        },
        y: {
          grid: { display: false },
          ticks: { color: textColor },
        },
      },
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: (context) => this.formatArcTooltip(context.label, context.raw),
          },
        },
      },
    };
  });
  readonly treemapOptions: ChartOptions<'treemap'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          title: (items) => this.treemapSource(items[0]?.raw)?.name ?? '',
          label: (context) => {
            const source = this.treemapSource(context.raw);
            return source === null ? '' : this.formatMoney(source.amountMinor);
          },
        },
      },
    },
  };
  readonly barChartHeight = computed(() => Math.max(160, this.legendItems().length * 40));

  formatPercentage(value: number): string {
    return new Intl.NumberFormat(undefined, {
      style: 'percent',
      maximumFractionDigits: 1,
    }).format(value);
  }

  selectVisualization(visualization: CategoryBreakdownVisualization): void {
    if (visualization !== this.visualization()) {
      this.visualizationChangeRequested.emit(visualization);
    }
  }

  private arcDataset() {
    return {
      data: this.legendItems().map(({ amountMinor }) => amountMinor),
      backgroundColor: this.legendItems().map(({ color }) => color),
      borderColor: this.resolveTokenColor('--mat-sys-surface-container-lowest'),
      borderWidth: 2,
    };
  }

  private formatArcTooltip(label: string, raw: unknown): string {
    return `${label}: ${this.formatMoney(Number(raw))}`;
  }

  private treemapSource(raw: unknown): CategoryTreemapSource | null {
    const dataPoint = raw as TreemapDataPoint | undefined;
    const source = dataPoint?._data;
    if (
      source === undefined ||
      typeof source['name'] !== 'string' ||
      typeof source['amountMinor'] !== 'number' ||
      typeof source['color'] !== 'string'
    ) {
      return null;
    }

    return source as CategoryTreemapSource;
  }

  private formatMoney(amountMinor: number): string {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency: this.currency(),
    }).format(amountMinor / 100);
  }

  private resolveCategoryColor(category: Category): string {
    if (category.colorHex !== null) {
      return category.colorHex;
    }

    return this.resolveTokenColor(
      category.colorKey === 'neutral' ? '--chart-neutral' : `--${category.colorKey}`,
    );
  }

  private resolveTokenColor(token: string): string {
    this.appTheme.current();
    const probe = this.document.createElement('span');
    probe.style.color = `var(${token})`;
    probe.style.colorScheme = this.theme.effectiveColorScheme();
    probe.style.display = 'none';
    this.document.body.append(probe);
    const color = this.document.defaultView?.getComputedStyle(probe).color ?? '';
    probe.remove();
    return color;
  }
}
