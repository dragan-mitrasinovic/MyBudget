package store

import (
	"context"
	"errors"
	"testing"
)

func TestCreateCategoryRejectsDuplicateName(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "category-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	if _, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food"); err != nil {
		t.Fatalf("create first category: %v", err)
	}

	if _, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food"); !errors.Is(err, ErrDuplicateName) {
		t.Fatalf("create duplicate category error = %v, want ErrDuplicateName", err)
	}
}

func TestCreateCategoryAssignsDistinctPaletteColors(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "category-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	food, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create Food category: %v", err)
	}

	transport, err := st.CreateCategory(ctx, user.ID, budget.ID, "Transport")
	if err != nil {
		t.Fatalf("create Transport category: %v", err)
	}

	if food.ColorKey != "chart-1" || transport.ColorKey != "chart-2" {
		t.Errorf("assigned colors = %q, %q, want chart-1, chart-2", food.ColorKey, transport.ColorKey)
	}
}

func TestUpdateCategoryColorAllowsSystemCategoryAndChecksOwnership(t *testing.T) {
	st := openTestStore(t)
	owner := createStoreTestUser(t, st, "category-owner")
	other := createStoreTestUser(t, st, "other-user")
	budget := createStoreTestBudget(t, st, owner.ID, "Household")
	ctx := context.Background()

	categories, err := st.ListCategories(ctx, owner.ID, budget.ID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	systemCategory := categories[0]
	customHex := "#AABBCC"

	updated, err := st.UpdateCategoryColor(ctx, owner.ID, systemCategory.ID, "chart-3", &customHex)
	if err != nil {
		t.Fatalf("update system category color: %v", err)
	}

	if updated.ColorKey != "chart-3" || updated.ColorHex == nil || *updated.ColorHex != customHex {
		t.Errorf("updated category = %#v, want chart-3 with %s override", updated, customHex)
	}

	if _, err := st.UpdateCategoryColor(ctx, other.ID, systemCategory.ID, "chart-1", nil); !errors.Is(err, ErrNotFound) {
		t.Errorf("other user's update error = %v, want ErrNotFound", err)
	}
}

func TestSystemCategoryCannotBeModified(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "category-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	categories, err := st.ListCategories(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	systemCategory := categories[0]

	if _, err := st.UpdateCategory(ctx, user.ID, systemCategory.ID, "Other"); !errors.Is(err, ErrSystemCategory) {
		t.Errorf("update system category error = %v, want ErrSystemCategory", err)
	}

	if err := st.DeleteCategory(ctx, user.ID, systemCategory.ID); !errors.Is(err, ErrSystemCategory) {
		t.Errorf("delete system category error = %v, want ErrSystemCategory", err)
	}
}

func TestDeleteCategoryReassignsExpensesToUncategorized(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "category-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	categories, err := st.ListCategories(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("list initial categories: %v", err)
	}
	uncategorized := categories[0]

	custom, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	expense, err := st.CreateExpense(ctx, user.ID, budget.ID, custom.ID, 1599, "Lunch", "2026-07-11")
	if err != nil {
		t.Fatalf("create expense: %v", err)
	}

	if err := st.DeleteCategory(ctx, user.ID, custom.ID); err != nil {
		t.Fatalf("delete category: %v", err)
	}

	expenses, err := st.ListExpenses(ctx, user.ID, budget.ID, ExpenseFilter{})
	if err != nil {
		t.Fatalf("list expenses: %v", err)
	}

	if len(expenses) != 1 {
		t.Fatalf("expense count = %d, want 1", len(expenses))
	}

	if expenses[0].ID != expense.ID || expenses[0].CategoryID != uncategorized.ID {
		t.Errorf("reassigned expense = %#v, want expense %d in category %d", expenses[0], expense.ID, uncategorized.ID)
	}
}
