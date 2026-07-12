import { BreakpointObserver } from '@angular/cdk/layout';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { SocialAuthService } from '@abacritt/angularx-social-login';
import { of } from 'rxjs';

import { AuthService } from '../auth/auth.service';
import { Budget } from '../budgets/budget.models';
import { BudgetService } from '../budgets/budget.service';
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
          provide: BudgetService,
          useValue: { budgets: signal<Budget[]>([]).asReadonly(), load: vi.fn() },
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

  it('shows ten budget shortcuts and puts the rest in an overflow menu', async () => {
    const budgets = signal<Budget[]>(
      Array.from({ length: 12 }, (_, index) => ({
        id: index + 1,
        name: `Budget ${index + 1}`,
        currency: 'EUR',
        archived: false,
        createdAt: '2026-07-12T00:00:00Z',
        updatedAt: '2026-07-12T00:00:00Z',
      })),
    );
    const load = vi.fn().mockResolvedValue(undefined);

    await TestBed.configureTestingModule({
      imports: [Shell],
      providers: [
        provideRouter([]),
        {
          provide: AuthService,
          useValue: { user: signal({ id: 1 }) },
        },
        {
          provide: BudgetService,
          useValue: { budgets: budgets.asReadonly(), load },
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
            preference: signal<ThemePreference>('system').asReadonly(),
            effectiveColorScheme: signal<'light' | 'dark'>('light').asReadonly(),
            setPreference: vi.fn(),
          },
        },
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(Shell);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    const host = fixture.nativeElement as HTMLElement;
    const directLinks = Array.from(host.querySelectorAll<HTMLAnchorElement>('.budget-link'));
    expect(load).toHaveBeenCalledOnce();
    expect(directLinks).toHaveLength(10);
    expect(directLinks.map((link) => link.textContent?.trim())).toEqual(
      Array.from({ length: 10 }, (_, index) => `Budget ${index + 1}`),
    );

    const overflowTrigger = host.querySelector(
      '[aria-label="Show 2 more budgets"]',
    ) as HTMLButtonElement | null;
    expect(overflowTrigger).not.toBeNull();
    overflowTrigger?.click();
    fixture.detectChanges();
    await fixture.whenStable();

    const overflowLinks = Array.from(
      document.body.querySelectorAll<HTMLAnchorElement>('.overflow-budget-link'),
    );
    expect(overflowLinks.map((link) => link.textContent?.trim())).toEqual([
      'Budget 11',
      'Budget 12',
    ]);
  });
});
