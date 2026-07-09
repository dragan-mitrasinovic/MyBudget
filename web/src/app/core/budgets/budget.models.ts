export interface Budget {
  id: number;
  name: string;
  currency: string;
  targetMinor: number | null;
  archived: boolean;
  createdAt: string;
  updatedAt: string;
}

/** Payload for creating or updating a budget. Amounts are in minor units (e.g. cents). */
export interface BudgetInput {
  name: string;
  currency: string;
  targetMinor: number | null;
  archived?: boolean;
}
