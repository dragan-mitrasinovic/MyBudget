import { Component, inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatButtonToggleChange, MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';

import { APP_THEMES, AppTheme, AppThemeService } from './app-theme.service';
import { ThemePreference, ThemeService } from './theme.service';

@Component({
  selector: 'app-appearance-dialog',
  imports: [MatButtonModule, MatButtonToggleModule, MatDialogModule, MatIconModule],
  templateUrl: './appearance-dialog.html',
  styleUrl: './appearance-dialog.scss',
})
export class AppearanceDialog {
  private readonly appThemeService = inject(AppThemeService);
  private readonly themeService = inject(ThemeService);

  readonly appThemes = APP_THEMES;
  readonly appTheme = this.appThemeService.current;
  readonly themePreference = this.themeService.preference;

  setThemePreference(change: MatButtonToggleChange): void {
    this.themeService.setPreference(change.value as ThemePreference);
  }

  setAppTheme(theme: AppTheme): void {
    this.appThemeService.set(theme);
  }
}
