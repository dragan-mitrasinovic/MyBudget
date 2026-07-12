import { Component, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { BreakpointObserver, Breakpoints } from '@angular/cdk/layout';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatMenuModule } from '@angular/material/menu';
import { MatSidenav, MatSidenavModule } from '@angular/material/sidenav';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { SocialAuthService } from '@abacritt/angularx-social-login';
import { map } from 'rxjs';

import { AuthService } from '../auth/auth.service';
import { ThemePreference, ThemeService } from '../theme/theme.service';

@Component({
  selector: 'app-shell',
  imports: [
    RouterLink,
    RouterLinkActive,
    RouterOutlet,
    MatButtonModule,
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
export class Shell {
  private auth = inject(AuthService);
  private social = inject(SocialAuthService);
  private router = inject(Router);
  private breakpointObserver = inject(BreakpointObserver);
  private theme = inject(ThemeService);

  readonly user = this.auth.user;
  readonly themePreference = this.theme.preference;
  readonly effectiveColorScheme = this.theme.effectiveColorScheme;
  readonly isHandset = toSignal(
    this.breakpointObserver.observe(Breakpoints.Handset).pipe(map((result) => result.matches)),
    { initialValue: false },
  );

  closeNavigationOnHandset(sideNav: MatSidenav): void {
    if (this.isHandset()) {
      void sideNav.close();
    }
  }

  setThemePreference(preference: ThemePreference): void {
    this.theme.setPreference(preference);
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
