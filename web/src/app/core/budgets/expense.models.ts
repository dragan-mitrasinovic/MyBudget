export interface Expense {
  id: number;
  categoryId: number;
  amountMinor: number;
  description: string;
  spentAt: string;
  createdAt: string;
  updatedAt: string;
}

/** Payload for creating or updating an expense. `amountMinor` is in minor units;
 * `spentAt` is a calendar date in YYYY-MM-DD form. */
export interface ExpenseInput {
  categoryId: number;
  amountMinor: number;
  description: string;
  spentAt: string;
}
