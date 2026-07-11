export interface Budget {
  id: number;
  name: string;
  currency: string;
  archived: boolean;
  createdAt: string;
  updatedAt: string;
}

/** Payload for creating or updating a budget. */
export interface BudgetInput {
  name: string;
  currency: string;
  archived?: boolean;
}
