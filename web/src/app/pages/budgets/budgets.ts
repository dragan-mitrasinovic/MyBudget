import { Component, OnInit, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { firstValueFrom } from 'rxjs';

import { Budget, BudgetInput } from '../../core/budgets/budget.models';
import { BudgetService } from '../../core/budgets/budget.service';
import { MoneyPipe } from '../../core/budgets/money.pipe';
import { ConfirmDialog, ConfirmData } from '../../shared/confirm-dialog/confirm-dialog';
import { BudgetFormDialog } from './budget-form-dialog';

@Component({
  selector: 'app-budgets',
  imports: [
    MatButtonModule,
    MatCardModule,
    MatDialogModule,
    MatIconModule,
    MatMenuModule,
    MatProgressSpinnerModule,
    MoneyPipe,
  ],
  templateUrl: './budgets.html',
  styleUrl: './budgets.scss',
})
export class Budgets implements OnInit {
  private service = inject(BudgetService);
  private dialog = inject(MatDialog);
  private router = inject(Router);

  readonly budgets = this.service.budgets;
  readonly loading = signal(true);

  async ngOnInit(): Promise<void> {
    try {
      await this.service.load();
    } finally {
      this.loading.set(false);
    }
  }

  open(budget: Budget): void {
    this.router.navigate(['/budgets', budget.id]);
  }

  async create(): Promise<void> {
    const result = await firstValueFrom(
      this.dialog
        .open<BudgetFormDialog, Budget | null, BudgetInput>(BudgetFormDialog, {
          data: null,
        })
        .afterClosed(),
    );
    if (result) {
      await this.service.create(result);
    }
  }

  async edit(budget: Budget): Promise<void> {
    const result = await firstValueFrom(
      this.dialog
        .open<BudgetFormDialog, Budget | null, BudgetInput>(BudgetFormDialog, {
          data: budget,
        })
        .afterClosed(),
    );
    if (result) {
      await this.service.update(budget.id, result);
    }
  }

  async remove(budget: Budget): Promise<void> {
    const data: ConfirmData = {
      title: 'Delete budget?',
      message: `“${budget.name}” and all of its categories and expenses will be permanently deleted.`,
      confirmText: 'Delete',
    };
    const confirmed = await firstValueFrom(
      this.dialog.open<ConfirmDialog, ConfirmData, boolean>(ConfirmDialog, { data }).afterClosed(),
    );
    if (confirmed) {
      await this.service.remove(budget.id);
    }
  }
}
