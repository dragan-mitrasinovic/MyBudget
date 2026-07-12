import { Category } from './category.models';

export const categoryColorKeys = [
  'chart-1',
  'chart-2',
  'chart-3',
  'chart-4',
  'chart-5',
  'chart-6',
  'chart-7',
  'chart-8',
  'chart-9',
  'chart-10',
  'chart-11',
  'chart-12',
  'neutral',
] as const;

export type CategoryColorKey = (typeof categoryColorKeys)[number];

export function categoryColorVariable(colorKey: string): string {
  return `var(--${colorKey === 'neutral' ? 'chart-neutral' : colorKey})`;
}

export function categoryDisplayColor(category: Category): string {
  return category.colorHex ?? categoryColorVariable(category.colorKey);
}
