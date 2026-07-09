import { Component, OnInit, inject, input, signal } from '@angular/core';
import { Router } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { Budget } from '../../../core/budgets/budget.models';
import { BudgetService } from '../../../core/budgets/budget.service';
import { MoneyPipe } from '../../../core/budgets/money.pipe';

@Component({
  selector: 'app-budget-detail',
  imports: [MatButtonModule, MatIconModule, MatProgressSpinnerModule, MoneyPipe],
  templateUrl: './budget-detail.html',
  styleUrl: './budget-detail.scss',
})
export class BudgetDetail implements OnInit {
  private service = inject(BudgetService);
  private router = inject(Router);

  /** Bound from the `:id` route param (withComponentInputBinding). */
  readonly id = input.required<string>();

  readonly budget = signal<Budget | null>(null);
  readonly loading = signal(true);
  readonly notFound = signal(false);

  async ngOnInit(): Promise<void> {
    try {
      this.budget.set(await this.service.get(Number(this.id())));
    } catch {
      this.notFound.set(true);
    } finally {
      this.loading.set(false);
    }
  }

  back(): void {
    this.router.navigate(['/budgets']);
  }
}
