import { DatePipe } from '@angular/common';
import { Component, inject } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { firstValueFrom } from 'rxjs';

import { Budget, BudgetInput } from '../../core/budgets/budget.models';
import { BudgetService } from '../../core/budgets/budget.service';
import { ConfirmDialog, ConfirmData } from '../../shared/confirm-dialog/confirm-dialog';
import { BudgetFormDialog } from './budget-form-dialog';

@Component({
  selector: 'app-budgets',
  imports: [
    DatePipe,
    RouterLink,
    MatButtonModule,
    MatCardModule,
    MatDialogModule,
    MatIconModule,
    MatMenuModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './budgets.html',
  styleUrl: './budgets.scss',
})
export class Budgets {
  private service = inject(BudgetService);
  private dialog = inject(MatDialog);
  private router = inject(Router);

  readonly budgets = this.service.budgets;
  readonly loading = this.service.loading;

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
      const budget = await this.service.create(result);
      this.open(budget);
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
