import { Component, inject } from '@angular/core';
import { Router } from '@angular/router';
import { SocialAuthService } from '@abacritt/angularx-social-login';

import { AuthService } from '../../core/auth/auth.service';

@Component({
  selector: 'app-home',
  templateUrl: './home.html',
  styleUrl: './home.scss',
})
export class Home {
  private auth = inject(AuthService);
  private social = inject(SocialAuthService);
  private router = inject(Router);

  readonly user = this.auth.user;

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
