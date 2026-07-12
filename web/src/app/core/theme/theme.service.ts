import { DOCUMENT } from '@angular/common';
import { MediaMatcher } from '@angular/cdk/layout';
import { DestroyRef, Injectable, computed, inject, signal } from '@angular/core';

export const THEME_STORAGE_KEY = 'mybudget.theme';

export type ThemePreference = 'system' | 'light' | 'dark';
export type EffectiveColorScheme = Exclude<ThemePreference, 'system'>;

const PREFERS_DARK_SCHEME = '(prefers-color-scheme: dark)';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly document = inject(DOCUMENT);
  private readonly mediaMatcher = inject(MediaMatcher);
  private readonly destroyRef = inject(DestroyRef);
  private readonly darkSchemeQuery = this.mediaMatcher.matchMedia(PREFERS_DARK_SCHEME);
  private readonly preferenceState = signal<ThemePreference>('system');
  private readonly systemPrefersDark = signal(false);
  private initialized = false;

  readonly preference = this.preferenceState.asReadonly();
  readonly effectiveColorScheme = computed<EffectiveColorScheme>(() => {
    const preference = this.preferenceState();
    return preference === 'system' ? (this.systemPrefersDark() ? 'dark' : 'light') : preference;
  });

  initialize(): void {
    if (this.initialized) {
      return;
    }

    this.initialized = true;
    this.systemPrefersDark.set(this.darkSchemeQuery.matches);
    this.applyPreference(this.readStoredPreference());
    this.darkSchemeQuery.addEventListener('change', this.handleSystemPreferenceChange);
    this.document.defaultView?.addEventListener('storage', this.handleStorageChange);

    this.destroyRef.onDestroy(() => {
      this.darkSchemeQuery.removeEventListener('change', this.handleSystemPreferenceChange);
      this.document.defaultView?.removeEventListener('storage', this.handleStorageChange);
    });
  }

  setPreference(preference: ThemePreference): void {
    this.applyPreference(preference);

    try {
      const storage = this.document.defaultView?.localStorage;
      if (preference === 'system') {
        storage?.removeItem(THEME_STORAGE_KEY);
      } else {
        storage?.setItem(THEME_STORAGE_KEY, preference);
      }
    } catch {
      // Browser policy can disable storage; the in-memory preference still applies.
    }
  }

  private readonly handleSystemPreferenceChange = (event: MediaQueryListEvent): void => {
    this.systemPrefersDark.set(event.matches);
  };

  private readonly handleStorageChange = (event: StorageEvent): void => {
    if (event.key !== THEME_STORAGE_KEY && event.key !== null) {
      return;
    }

    try {
      if (event.storageArea !== this.document.defaultView?.localStorage) {
        return;
      }
    } catch {
      return;
    }

    const preference =
      event.key === null ? this.readStoredPreference() : parseStoredPreference(event.newValue);
    this.applyPreference(preference);
  };

  private readStoredPreference(): ThemePreference {
    try {
      return parseStoredPreference(
        this.document.defaultView?.localStorage.getItem(THEME_STORAGE_KEY) ?? null,
      );
    } catch {
      return 'system';
    }
  }

  private applyPreference(preference: ThemePreference): void {
    this.preferenceState.set(preference);

    if (preference === 'system') {
      this.document.documentElement.removeAttribute('data-theme');
    } else {
      this.document.documentElement.setAttribute('data-theme', preference);
    }
  }
}

function parseStoredPreference(value: string | null): ThemePreference {
  return value === 'light' || value === 'dark' ? value : 'system';
}
