import { Component, OnDestroy, OnInit, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { GoogleSigninButtonDirective, SocialAuthService } from '@abacritt/angularx-social-login';
import { Subscription } from 'rxjs';

import { AuthService } from '../../core/auth/auth.service';

@Component({
  selector: 'app-login',
  imports: [GoogleSigninButtonDirective],
  templateUrl: './login.html',
  styleUrl: './login.scss',
})
export class Login implements OnInit, OnDestroy {
  private social = inject(SocialAuthService);
  private auth = inject(AuthService);
  private router = inject(Router);

  readonly error = signal<string | null>(null);
  private sub?: Subscription;

  ngOnInit(): void {
    if (this.auth.user()) {
      this.router.navigate(['/']);
      return;
    }

    // The library emits a SocialUser (carrying the Google ID token) once the user
    // signs in via the rendered button.
    this.sub = this.social.authState.subscribe((socialUser) => {
      if (socialUser?.idToken) {
        void this.completeLogin(socialUser.idToken);
      }
    });
  }

  ngOnDestroy(): void {
    this.sub?.unsubscribe();
  }

  private async completeLogin(idToken: string): Promise<void> {
    try {
      await this.auth.loginWithGoogle(idToken);
      this.router.navigate(['/']);
    } catch {
      this.error.set('Sign-in failed. Please try again.');
    }
  }
}
