import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../environments/environment';
import { DashboardLayout } from './dashboard.models';

@Injectable({ providedIn: 'root' })
export class DashboardService {
  private http = inject(HttpClient);
  private readonly api = environment.apiBase;

  get(budgetId: number): Promise<DashboardLayout> {
    return firstValueFrom(
      this.http.get<DashboardLayout>(`${this.api}/budgets/${budgetId}/dashboard`),
    );
  }

  save(budgetId: number, layout: DashboardLayout): Promise<DashboardLayout> {
    return firstValueFrom(
      this.http.put<DashboardLayout>(`${this.api}/budgets/${budgetId}/dashboard`, layout),
    );
  }
}
