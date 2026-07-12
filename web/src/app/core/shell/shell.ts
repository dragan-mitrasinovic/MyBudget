import { Component, OnInit, computed, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { NavigationEnd, Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatMenuModule } from '@angular/material/menu';
import { MatSidenav, MatSidenavModule } from '@angular/material/sidenav';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { SocialAuthService } from '@abacritt/angularx-social-login';
import { filter, map } from 'rxjs';

import { AuthService } from '../auth/auth.service';
import { BudgetService } from '../budgets/budget.service';
import { AppearanceDialog } from '../theme/appearance-dialog';

@Component({
  selector: 'app-shell',
  imports: [
    RouterLink,
    RouterLinkActive,
    RouterOutlet,
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatListModule,
    MatMenuModule,
    MatSidenavModule,
    MatToolbarModule,
    MatTooltipModule,
  ],
  templateUrl: './shell.html',
  styleUrl: './shell.scss',
})
export class Shell implements OnInit {
  private auth = inject(AuthService);
  private budgetService = inject(BudgetService);
  private social = inject(SocialAuthService);
  private router = inject(Router);
  private breakpointObserver = inject(BreakpointObserver);
  private dialog = inject(MatDialog);

  readonly user = this.auth.user;
  readonly budgets = this.budgetService.budgets;
  readonly sidebarBudgets = computed(() => this.budgets().slice(0, 10));
  readonly overflowBudgets = computed(() => this.budgets().slice(10));
  private readonly currentUrl = toSignal(
    this.router.events.pipe(
      filter((event): event is NavigationEnd => event instanceof NavigationEnd),
      map((event) => event.urlAfterRedirects),
    ),
    { initialValue: this.router.url },
  );
  readonly activeBudgetId = computed(() => {
    const match = /^\/budgets\/(\d+)(?:[/?#]|$)/.exec(this.currentUrl());
    return match ? Number(match[1]) : null;
  });
  readonly overflowBudgetActive = computed(() =>
    this.overflowBudgets().some((budget) => budget.id === this.activeBudgetId()),
  );
  readonly isHandset = toSignal(
    this.breakpointObserver.observe(Breakpoints.Handset).pipe(map((result) => result.matches)),
    { initialValue: false },
  );

  async ngOnInit(): Promise<void> {
    await this.budgetService.load();
  }

  closeNavigationOnHandset(sideNav: MatSidenav): void {
    if (this.isHandset()) {
      void sideNav.close();
    }
  }

  openAppearance(): void {
    this.dialog.open(AppearanceDialog, {
      autoFocus: false,
      maxWidth: 'min(56rem, 96vw)',
    });
  }

  async logout(): Promise<void> {
    // Reset the library's state so its replayed authState can't re-trigger a login.
    try {
      await this.social.signOut();
    } catch {
      // Not signed into GIS this session (e.g. after a reload) — nothing to reset.
    }

    await this.auth.logout();
    this.router.navigate(['/login']);
  }
}
