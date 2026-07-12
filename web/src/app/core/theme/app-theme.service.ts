import { DOCUMENT } from '@angular/common';
import { DestroyRef, Injectable, inject, signal } from '@angular/core';

export const APP_THEME_STORAGE_KEY = 'mybudget.app-theme';

export const APP_THEMES = [
  { value: 'soft-ledger', label: 'Soft Ledger' },
  { value: 'color-coded-studio', label: 'Color-Coded Studio' },
  { value: 'bright-utility', label: 'Bright Utility' },
  { value: 'signal-red', label: 'Signal Red' },
  { value: 'midnight-terminal', label: 'Midnight Terminal' },
  { value: 'candy-pop', label: 'Candy Pop' },
  { value: 'velvet-night', label: 'Velvet Night' },
  { value: 'vaporwave', label: 'Vaporwave' },
] as const;

export type AppTheme = (typeof APP_THEMES)[number]['value'];

const DEFAULT_APP_THEME: AppTheme = 'soft-ledger';
const APP_THEME_ATTRIBUTE = 'data-app-theme';

@Injectable({ providedIn: 'root' })
export class AppThemeService {
  private readonly document = inject(DOCUMENT);
  private readonly destroyRef = inject(DestroyRef);
  private readonly currentState = signal<AppTheme>(DEFAULT_APP_THEME);
  private initialized = false;

  readonly current = this.currentState.asReadonly();

  initialize(): void {
    if (this.initialized) {
      return;
    }

    this.initialized = true;
    this.apply(this.readStoredTheme());
    this.document.defaultView?.addEventListener('storage', this.handleStorageChange);
    this.destroyRef.onDestroy(() => {
      this.document.defaultView?.removeEventListener('storage', this.handleStorageChange);
    });
  }

  set(theme: AppTheme): void {
    this.apply(theme);

    try {
      this.document.defaultView?.localStorage.setItem(APP_THEME_STORAGE_KEY, theme);
    } catch {
      // Browser policy can disable storage; the in-memory theme still applies.
    }
  }

  private readonly handleStorageChange = (event: StorageEvent): void => {
    if (event.key !== APP_THEME_STORAGE_KEY && event.key !== null) {
      return;
    }

    try {
      if (event.storageArea !== this.document.defaultView?.localStorage) {
        return;
      }
    } catch {
      return;
    }

    const theme = event.key === null ? this.readStoredTheme() : parseStoredTheme(event.newValue);
    this.apply(theme);
  };

  private readStoredTheme(): AppTheme {
    try {
      return parseStoredTheme(
        this.document.defaultView?.localStorage.getItem(APP_THEME_STORAGE_KEY) ?? null,
      );
    } catch {
      return DEFAULT_APP_THEME;
    }
  }

  private apply(theme: AppTheme): void {
    this.currentState.set(theme);
    this.document.documentElement.setAttribute(APP_THEME_ATTRIBUTE, theme);
  }
}

function parseStoredTheme(value: string | null): AppTheme {
  const theme = APP_THEMES.find((option) => option.value === value);
  return theme?.value ?? DEFAULT_APP_THEME;
}
