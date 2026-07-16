export interface Category {
  id: number;
  budgetId: number;
  name: string;
  colorKey: string;
  colorHex: string | null;
  createdAt: string;
}

export interface CategoryColorInput {
  colorKey: string;
  colorHex: string | null;
}
