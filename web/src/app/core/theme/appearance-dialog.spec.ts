import { signal } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MatButtonToggleChange } from '@angular/material/button-toggle';

import { AppTheme, AppThemeService } from './app-theme.service';
import { AppearanceDialog } from './appearance-dialog';
import { ThemePreference, ThemeService } from './theme.service';

describe('AppearanceDialog', () => {
  afterEach(() => TestBed.resetTestingModule());

  it('offers eight visual styles and delegates a live selection', () => {
    const { fixture, setAppTheme } = setup();
    fixture.detectChanges();

    const host = fixture.nativeElement as HTMLElement;
    const styleOptions = Array.from(
      host.querySelectorAll<HTMLButtonElement>('.style-option[role="radio"]'),
    );
    expect(styleOptions).toHaveLength(8);
    expect(
      styleOptions.map((option) => option.querySelector('.style-label')?.textContent?.trim()),
    ).toEqual([
      'Soft Ledger',
      'Color-Coded Studio',
      'Bright Utility',
      'Signal Red',
      'Midnight Terminal',
      'Candy Pop',
      'Velvet Night',
      'Vaporwave',
    ]);

    styleOptions[7]?.click();

    expect(setAppTheme).toHaveBeenCalledWith('vaporwave');
  });

  it('delegates System, Light, and Dark independently from style', () => {
    const { component, setPreference, setAppTheme } = setup();

    component.setThemePreference({ value: 'dark' } as MatButtonToggleChange);

    expect(setPreference).toHaveBeenCalledWith('dark');
    expect(setAppTheme).not.toHaveBeenCalled();
  });

  function setup(): {
    component: AppearanceDialog;
    fixture: ComponentFixture<AppearanceDialog>;
    setAppTheme: ReturnType<typeof vi.fn>;
    setPreference: ReturnType<typeof vi.fn>;
  } {
    const appTheme = signal<AppTheme>('soft-ledger');
    const preference = signal<ThemePreference>('system');
    const setAppTheme = vi.fn((theme: AppTheme) => appTheme.set(theme));
    const setPreference = vi.fn((value: ThemePreference) => preference.set(value));

    TestBed.configureTestingModule({
      imports: [AppearanceDialog],
      providers: [
        {
          provide: AppThemeService,
          useValue: { current: appTheme.asReadonly(), set: setAppTheme },
        },
        {
          provide: ThemeService,
          useValue: { preference: preference.asReadonly(), setPreference },
        },
      ],
    });

    const fixture = TestBed.createComponent(AppearanceDialog);
    return { component: fixture.componentInstance, fixture, setAppTheme, setPreference };
  }
});
