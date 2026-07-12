import { BreakpointObserver } from '@angular/cdk/layout';
import { signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { SocialAuthService } from '@abacritt/angularx-social-login';
import { of } from 'rxjs';

import { AuthService } from '../auth/auth.service';
import { Budget } from '../budgets/budget.models';
import { BudgetService } from '../budgets/budget.service';
import { Shell } from './shell';

describe('Shell', () => {
  afterEach(() => {
    TestBed.resetTestingModule();
  });

  it('opens one Appearance dialog from the toolbar', async () => {
    const fixture = await createShell();
    const openAppearance = vi
      .spyOn(fixture.componentInstance, 'openAppearance')
      .mockImplementation(() => undefined);

    const trigger = fixture.nativeElement.querySelector(
      '[aria-label="Customize appearance"]',
    ) as HTMLButtonElement | null;
    expect(trigger).not.toBeNull();
    expect(fixture.nativeElement.querySelector('[aria-label="Choose color theme"]')).toBeNull();

    trigger?.click();
    expect(openAppearance).toHaveBeenCalledOnce();
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

  async function createShell() {
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
      ],
    }).compileComponents();
    const fixture = TestBed.createComponent(Shell);
    fixture.detectChanges();
    return fixture;
  }
});
