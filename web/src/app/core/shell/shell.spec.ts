import { BreakpointObserver } from '@angular/cdk/layout';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { SocialAuthService } from '@abacritt/angularx-social-login';
import { of } from 'rxjs';

import { AuthService } from '../auth/auth.service';
import { ThemePreference, ThemeService } from '../theme/theme.service';
import { Shell } from './shell';

describe('Shell', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('offers System, Light, and Dark preferences and delegates the selection', async () => {
    const preference = signal<ThemePreference>('system');
    const effectiveColorScheme = signal<'light' | 'dark'>('light');
    const setPreference = vi.fn((value: ThemePreference) => {
      preference.set(value);
      effectiveColorScheme.set(value === 'dark' ? 'dark' : 'light');
    });

    await TestBed.configureTestingModule({
      imports: [Shell],
      providers: [
        provideRouter([]),
        {
          provide: AuthService,
          useValue: { user: signal({ id: 1 }) },
        },
        {
          provide: SocialAuthService,
          useValue: { signOut: vi.fn() },
        },
        {
          provide: BreakpointObserver,
          useValue: { observe: () => of({ matches: false }) },
        },
        {
          provide: ThemeService,
          useValue: {
            preference: preference.asReadonly(),
            effectiveColorScheme: effectiveColorScheme.asReadonly(),
            setPreference,
          },
        },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(Shell);
    fixture.detectChanges();

    const trigger = fixture.nativeElement.querySelector(
      '[aria-label="Choose color theme"]',
    ) as HTMLButtonElement | null;
    expect(trigger).not.toBeNull();

    trigger?.click();
    fixture.detectChanges();
    await fixture.whenStable();

    const menuItems = Array.from(
      document.body.querySelectorAll<HTMLButtonElement>('[role="menuitemradio"]'),
    );
    expect(menuItems).toHaveLength(3);
    expect(menuItems.map((item) => item.textContent)).toEqual([
      expect.stringContaining('System'),
      expect.stringContaining('Light'),
      expect.stringContaining('Dark'),
    ]);
    expect(menuItems.map((item) => item.getAttribute('aria-checked'))).toEqual([
      'true',
      'false',
      'false',
    ]);

    menuItems[2]?.click();

    expect(setPreference).toHaveBeenCalledOnce();
    expect(setPreference).toHaveBeenCalledWith('dark');
  });
});
