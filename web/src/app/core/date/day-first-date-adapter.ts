import { Injectable, Provider } from '@angular/core';
import {
  DateAdapter,
  MAT_DATE_FORMATS,
  MAT_DATE_LOCALE,
  MAT_NATIVE_DATE_FORMATS,
  MatDateFormats,
  NativeDateAdapter,
} from '@angular/material/core';

const dayFirstDateInputFormat = 'DD/MM/YYYY';

export const dayFirstDateFormats: MatDateFormats = {
  parse: {
    ...MAT_NATIVE_DATE_FORMATS.parse,
    dateInput: dayFirstDateInputFormat,
  },
  display: {
    ...MAT_NATIVE_DATE_FORMATS.display,
    dateInput: dayFirstDateInputFormat,
  },
};

@Injectable()
export class DayFirstDateAdapter extends NativeDateAdapter {
  override parse(value: unknown): Date | null {
    if (value instanceof Date) {
      return this.isValid(value) ? this.clone(value) : this.invalid();
    }

    if (typeof value !== 'string') {
      return super.parse(value);
    }

    const trimmed = value.trim();
    if (trimmed === '') {
      return null;
    }

    const match = /^(\d{1,2})\/(\d{1,2})\/(\d{4})$/.exec(trimmed);
    if (match === null) {
      return this.invalid();
    }

    const [, dayText, monthText, yearText] = match;
    const day = Number(dayText);
    const month = Number(monthText);
    const year = Number(yearText);
    const date = new Date(year, month - 1, day);
    if (year >= 0 && year < 100) {
      date.setFullYear(year);
    }

    return date.getFullYear() === year && date.getMonth() === month - 1 && date.getDate() === day
      ? date
      : this.invalid();
  }

  override format(date: Date, displayFormat: unknown): string {
    if (displayFormat !== dayFirstDateInputFormat) {
      return super.format(date, displayFormat as Intl.DateTimeFormatOptions);
    }

    if (!this.isValid(date)) {
      throw new Error('DayFirstDateAdapter: Cannot format invalid date.');
    }

    const day = String(date.getDate()).padStart(2, '0');
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const year = String(date.getFullYear()).padStart(4, '0');
    return `${day}/${month}/${year}`;
  }
}

export function provideDayFirstDateAdapter(): Provider[] {
  return [
    { provide: MAT_DATE_LOCALE, useValue: 'en-GB' },
    { provide: DateAdapter, useClass: DayFirstDateAdapter },
    { provide: MAT_DATE_FORMATS, useValue: dayFirstDateFormats },
  ];
}
