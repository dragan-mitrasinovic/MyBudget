import { TestBed } from '@angular/core/testing';

import { APP_THEMES, APP_THEME_STORAGE_KEY, AppThemeService } from './app-theme.service';

describe('AppThemeService', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-app-theme');
  });

  afterEach(() => {
    TestBed.resetTestingModule();
    vi.restoreAllMocks();
    localStorage.clear();
    document.documentElement.removeAttribute('data-app-theme');
    document.documentElement.removeAttribute('data-theme');
  });

  it('defaults to Soft Ledger and restores a stored theme', () => {
    const defaultService = initializeService();

    expect(defaultService.current()).toBe('soft-ledger');
    expect(document.documentElement.getAttribute('data-app-theme')).toBe('soft-ledger');

    TestBed.resetTestingModule();
    localStorage.setItem(APP_THEME_STORAGE_KEY, 'midnight-terminal');
    const storedService = initializeService();

    expect(storedService.current()).toBe('midnight-terminal');
    expect(document.documentElement.getAttribute('data-app-theme')).toBe('midnight-terminal');
  });

  it('persists a selection without changing the light or dark preference', () => {
    document.documentElement.setAttribute('data-theme', 'dark');
    const service = initializeService();

    service.set('signal-red');

    expect(service.current()).toBe('signal-red');
    expect(localStorage.getItem(APP_THEME_STORAGE_KEY)).toBe('signal-red');
    expect(document.documentElement.getAttribute('data-app-theme')).toBe('signal-red');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('falls back from unrecognized stored values and follows changes from another tab', () => {
    localStorage.setItem(APP_THEME_STORAGE_KEY, 'unknown');
    const service = initializeService();

    expect(service.current()).toBe('soft-ledger');

    window.dispatchEvent(
      new StorageEvent('storage', {
        key: APP_THEME_STORAGE_KEY,
        newValue: 'bright-utility',
        storageArea: localStorage,
      }),
    );

    expect(service.current()).toBe('bright-utility');
    expect(document.documentElement.getAttribute('data-app-theme')).toBe('bright-utility');
  });

  it('still applies the current selection when storage is unavailable', () => {
    const service = initializeService();
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new DOMException('Storage blocked');
    });

    service.set('color-coded-studio');

    expect(service.current()).toBe('color-coded-studio');
    expect(document.documentElement.getAttribute('data-app-theme')).toBe('color-coded-studio');
  });

  it.each(APP_THEMES)('persists the $label theme', ({ value }) => {
    const service = initializeService();

    service.set(value);

    expect(service.current()).toBe(value);
    expect(localStorage.getItem(APP_THEME_STORAGE_KEY)).toBe(value);
    expect(document.documentElement.getAttribute('data-app-theme')).toBe(value);
  });

  function initializeService(): AppThemeService {
    const service = TestBed.inject(AppThemeService);
    service.initialize();
    return service;
  }
});
