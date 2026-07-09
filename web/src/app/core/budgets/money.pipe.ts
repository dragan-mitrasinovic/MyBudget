import { Pipe, PipeTransform } from '@angular/core';

/**
 * Formats an integer minor-unit amount (e.g. cents) as a localized currency string.
 * Assumes two-decimal currencies for now (see ADR-0011; multi-currency is deferred).
 */
@Pipe({ name: 'money' })
export class MoneyPipe implements PipeTransform {
  transform(minor: number | null | undefined, currency: string): string {
    if (minor == null) {
      return '';
    }

    return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(minor / 100);
  }
}
