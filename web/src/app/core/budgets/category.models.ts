export interface Category {
  id: number;
  budgetId: number;
  name: string;
  isSystem: boolean;
  colorKey: string;
  colorHex: string | null;
  createdAt: string;
}

export interface CategoryColorInput {
  colorKey: string;
  colorHex: string | null;
}
