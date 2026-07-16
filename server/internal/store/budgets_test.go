package store

import (
	"context"
	"testing"
)

func TestCreateBudgetStartsWithoutCategories(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "budget-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")

	categories, err := st.ListCategories(context.Background(), user.ID, budget.ID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	if len(categories) != 0 {
		t.Fatalf("category count = %d, want 0", len(categories))
	}
}

func TestBudgetOperationsAreScopedToOwner(t *testing.T) {
	st := openTestStore(t)
	owner := createStoreTestUser(t, st, "budget-owner")
	otherUser := createStoreTestUser(t, st, "other-user")
	budget := createStoreTestBudget(t, st, owner.ID, "Household")
	ctx := context.Background()

	listed, err := st.ListBudgets(ctx, otherUser.ID)
	if err != nil {
		t.Fatalf("list budgets as other user: %v", err)
	}

	if len(listed) != 0 {
		t.Fatalf("other user's budget count = %d, want 0", len(listed))
	}

	got, err := st.GetBudget(ctx, otherUser.ID, budget.ID)
	if err != nil {
		t.Fatalf("get budget as other user: %v", err)
	}

	if got != nil {
		t.Fatalf("get budget as other user = %#v, want nil", got)
	}

	updated, err := st.UpdateBudget(ctx, otherUser.ID, budget.ID, "Changed", "USD", true)
	if err != nil {
		t.Fatalf("update budget as other user: %v", err)
	}

	if updated != nil {
		t.Fatalf("update budget as other user = %#v, want nil", updated)
	}

	deleted, err := st.DeleteBudget(ctx, otherUser.ID, budget.ID)
	if err != nil {
		t.Fatalf("delete budget as other user: %v", err)
	}

	if deleted {
		t.Fatal("delete budget as other user = true, want false")
	}

	unchanged, err := st.GetBudget(ctx, owner.ID, budget.ID)
	if err != nil {
		t.Fatalf("get budget as owner: %v", err)
	}

	if unchanged == nil || unchanged.Name != "Household" || unchanged.Currency != "EUR" || unchanged.Archived {
		t.Fatalf("owner's budget after foreign mutations = %#v, want unchanged", unchanged)
	}
}

func TestDeleteBudgetCascadesCategoriesAndExpenses(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "budget-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	category, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	if _, err := st.CreateExpense(ctx, user.ID, budget.ID, category.ID, 1599, "Lunch", "2026-07-11"); err != nil {
		t.Fatalf("create expense: %v", err)
	}

	deleted, err := st.DeleteBudget(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("delete budget: %v", err)
	}

	if !deleted {
		t.Fatal("delete budget = false, want true")
	}

	var categoryCount int
	if err := st.db.QueryRow("SELECT COUNT(*) FROM categories WHERE budget_id = ?", budget.ID).Scan(&categoryCount); err != nil {
		t.Fatalf("count remaining categories: %v", err)
	}

	if categoryCount != 0 {
		t.Errorf("remaining categories = %d, want 0", categoryCount)
	}

	var expenseCount int
	if err := st.db.QueryRow("SELECT COUNT(*) FROM expenses WHERE category_id = ?", category.ID).Scan(&expenseCount); err != nil {
		t.Fatalf("count remaining expenses: %v", err)
	}

	if expenseCount != 0 {
		t.Errorf("remaining expenses = %d, want 0", expenseCount)
	}
}
