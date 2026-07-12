import { MediaMatcher } from '@angular/cdk/layout';
import { TestBed } from '@angular/core/testing';

import { THEME_STORAGE_KEY, ThemeService } from './theme.service';

describe('ThemeService', () => {
  let mediaQuery: MediaQueryList;
  let mediaChangeListener: (event: MediaQueryListEvent) => void;

  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');

    mediaQuery = {
      matches: false,
      media: '(prefers-color-scheme: dark)',
      onchange: null,
      addEventListener: vi.fn((_type: string, listener: (event: MediaQueryListEvent) => void) => {
        mediaChangeListener = listener;
      }),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    } as unknown as MediaQueryList;

    TestBed.configureTestingModule({
      providers: [
        {
          provide: MediaMatcher,
          useValue: { matchMedia: vi.fn(() => mediaQuery) },
        },
      ],
    });
  });

  afterEach(() => {
    TestBed.resetTestingModule();
    vi.restoreAllMocks();
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  it('defaults to System and follows the current operating-system scheme', () => {
    Object.defineProperty(mediaQuery, 'matches', { value: true });

    const service = initializeService();

    expect(service.preference()).toBe('system');
    expect(service.effectiveColorScheme()).toBe('dark');
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
  });

  it('restores valid explicit preferences and ignores unrecognized values', () => {
    localStorage.setItem(THEME_STORAGE_KEY, 'dark');
    const service = initializeService();

    expect(service.preference()).toBe('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');

    TestBed.resetTestingModule();
    document.documentElement.setAttribute('data-theme', 'dark');
    localStorage.setItem(THEME_STORAGE_KEY, 'sepia');
    TestBed.configureTestingModule({
      providers: [{ provide: MediaMatcher, useValue: { matchMedia: () => mediaQuery } }],
    });
    const fallbackService = initializeService();

    expect(fallbackService.preference()).toBe('system');
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
  });

  it('persists explicit preferences and removes storage when returning to System', () => {
    const service = initializeService();

    service.setPreference('light');

    expect(service.preference()).toBe('light');
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');

    service.setPreference('system');

    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBeNull();
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
  });

  it('reacts to operating-system changes only while using System', () => {
    const service = initializeService();

    mediaChangeListener({ matches: true } as MediaQueryListEvent);
    expect(service.effectiveColorScheme()).toBe('dark');

    service.setPreference('light');
    mediaChangeListener({ matches: false } as MediaQueryListEvent);
    mediaChangeListener({ matches: true } as MediaQueryListEvent);

    expect(service.effectiveColorScheme()).toBe('light');
  });

  it('applies theme changes made in another tab', () => {
    const service = initializeService();

    window.dispatchEvent(
      new StorageEvent('storage', {
        key: THEME_STORAGE_KEY,
        newValue: 'dark',
        storageArea: localStorage,
      }),
    );

    expect(service.preference()).toBe('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('still applies the current-tab preference when storage is unavailable', () => {
    const service = initializeService();
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new DOMException('Storage blocked');
    });

    service.setPreference('dark');

    expect(service.preference()).toBe('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  function initializeService(): ThemeService {
    const service = TestBed.inject(ThemeService);
    service.initialize();
    return service;
  }
});
