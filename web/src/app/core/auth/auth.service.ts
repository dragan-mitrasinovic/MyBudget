import { HttpClient } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { catchError, firstValueFrom, of } from 'rxjs';

import { environment } from '../../../environments/environment';

export interface User {
  id: number;
  email: string;
  name: string;
  picture: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);
  private readonly api = environment.apiBase;

  /** The currently signed-in user, or null. Drives change detection (zoneless). */
  readonly user = signal<User | null>(null);

  /** Hydrate auth state at startup; a 401 simply means "not signed in". */
  async init(): Promise<void> {
    const user = await firstValueFrom(
      this.http.get<User>(`${this.api}/me`).pipe(catchError(() => of(null))),
    );
    this.user.set(user);
  }

  /** Exchange a Google ID token for a session; first login registers the user. */
  async loginWithGoogle(credential: string): Promise<void> {
    const user = await firstValueFrom(
      this.http.post<User>(`${this.api}/auth/google`, { credential }),
    );
    this.user.set(user);
  }

  async logout(): Promise<void> {
    await firstValueFrom(this.http.post(`${this.api}/auth/logout`, {}));
    this.user.set(null);
  }
}
