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

func TestListCategoriesDistinguishesOwnedEmptyBudgetFromUnownedBudget(t *testing.T) {
	st := openTestStore(t)
	owner := createStoreTestUser(t, st, "category-owner")
	other := createStoreTestUser(t, st, "other-user")
	budget := createStoreTestBudget(t, st, owner.ID, "Household")
	ctx := context.Background()

	categories, err := st.ListCategories(ctx, owner.ID, budget.ID)
	if err != nil {
		t.Fatalf("list owned categories: %v", err)
	}

	if categories == nil || len(categories) != 0 {
		t.Fatalf("owned empty categories = %#v, want allocated empty slice", categories)
	}

	if _, err := st.ListCategories(ctx, other.ID, budget.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("list unowned categories error = %v, want ErrNotFound", err)
	}
}

func TestUpdateCategoryColorChecksOwnership(t *testing.T) {
	st := openTestStore(t)
	owner := createStoreTestUser(t, st, "category-owner")
	other := createStoreTestUser(t, st, "other-user")
	budget := createStoreTestBudget(t, st, owner.ID, "Household")
	ctx := context.Background()

	category, err := st.CreateCategory(ctx, owner.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	customHex := "#AABBCC"

	updated, err := st.UpdateCategoryColor(ctx, owner.ID, category.ID, "chart-3", &customHex)
	if err != nil {
		t.Fatalf("update category color: %v", err)
	}

	if updated.ColorKey != "chart-3" || updated.ColorHex == nil || *updated.ColorHex != customHex {
		t.Errorf("updated category = %#v, want chart-3 with %s override", updated, customHex)
	}

	if _, err := st.UpdateCategoryColor(ctx, other.ID, category.ID, "chart-1", nil); !errors.Is(err, ErrNotFound) {
		t.Errorf("other user's update error = %v, want ErrNotFound", err)
	}
}

func TestDeleteEmptyCategoryWithoutReassignment(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "category-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	category, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	if err := st.DeleteCategory(ctx, user.ID, category.ID, nil); err != nil {
		t.Fatalf("delete empty category: %v", err)
	}

	categories, err := st.ListCategories(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	if len(categories) != 0 {
		t.Fatalf("category count = %d, want 0", len(categories))
	}
}

func TestDeleteCategoryReassignsExpensesToSelectedCategory(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "category-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	source, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create source category: %v", err)
	}

	target, err := st.CreateCategory(ctx, user.ID, budget.ID, "Other")
	if err != nil {
		t.Fatalf("create target category: %v", err)
	}

	expense, err := st.CreateExpense(ctx, user.ID, budget.ID, source.ID, 1599, "Lunch", "2026-07-11")
	if err != nil {
		t.Fatalf("create expense: %v", err)
	}

	const previousUpdatedAt = "2020-01-01T00:00:00Z"
	if _, err := st.db.ExecContext(ctx, `UPDATE expenses SET updated_at = ? WHERE id = ?`, previousUpdatedAt, expense.ID); err != nil {
		t.Fatalf("set previous expense timestamp: %v", err)
	}

	if err := st.DeleteCategory(ctx, user.ID, source.ID, &target.ID); err != nil {
		t.Fatalf("delete category: %v", err)
	}

	expenses, err := st.ListExpenses(ctx, user.ID, budget.ID, ExpenseFilter{})
	if err != nil {
		t.Fatalf("list expenses: %v", err)
	}

	if len(expenses) != 1 {
		t.Fatalf("expense count = %d, want 1", len(expenses))
	}

	if expenses[0].ID != expense.ID || expenses[0].CategoryID != target.ID {
		t.Errorf("reassigned expense = %#v, want expense %d in category %d", expenses[0], expense.ID, target.ID)
	}

	if expenses[0].UpdatedAt == previousUpdatedAt {
		t.Errorf("reassigned expense updatedAt = %q, want a refreshed timestamp", expenses[0].UpdatedAt)
	}
}

func TestDeleteCategoryRequiresReassignmentWhenNotEmpty(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "category-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	category, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	if _, err := st.CreateExpense(ctx, user.ID, budget.ID, category.ID, 1599, "Lunch", "2026-07-11"); err != nil {
		t.Fatalf("create expense: %v", err)
	}

	if err := st.DeleteCategory(ctx, user.ID, category.ID, nil); !errors.Is(err, ErrCategoryNotEmpty) {
		t.Fatalf("delete category error = %v, want ErrCategoryNotEmpty", err)
	}

	categories, err := st.ListCategories(ctx, user.ID, budget.ID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	if len(categories) != 1 || categories[0].ID != category.ID {
		t.Fatalf("categories after rejected deletion = %#v, want source category", categories)
	}

	expenses, err := st.ListExpenses(ctx, user.ID, budget.ID, ExpenseFilter{})
	if err != nil {
		t.Fatalf("list expenses: %v", err)
	}

	if len(expenses) != 1 || expenses[0].CategoryID != category.ID {
		t.Fatalf("expenses after rejected deletion = %#v, want expense in source category", expenses)
	}
}

func TestDeleteCategoryRejectsInvalidReassignmentTargets(t *testing.T) {
	st := openTestStore(t)
	owner := createStoreTestUser(t, st, "category-owner")
	other := createStoreTestUser(t, st, "other-user")
	budget := createStoreTestBudget(t, st, owner.ID, "Household")
	otherBudget := createStoreTestBudget(t, st, owner.ID, "Travel")
	ctx := context.Background()

	source, err := st.CreateCategory(ctx, owner.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create source category: %v", err)
	}

	foreign, err := st.CreateCategory(ctx, owner.ID, otherBudget.ID, "Travel")
	if err != nil {
		t.Fatalf("create foreign category: %v", err)
	}

	if _, err := st.CreateExpense(ctx, owner.ID, budget.ID, source.ID, 1599, "Lunch", "2026-07-11"); err != nil {
		t.Fatalf("create expense: %v", err)
	}

	zero := int64(0)
	missing := int64(999999)
	tests := []struct {
		name     string
		targetID *int64
	}{
		{name: "same category", targetID: &source.ID},
		{name: "non-positive category", targetID: &zero},
		{name: "missing category", targetID: &missing},
		{name: "category in another budget", targetID: &foreign.ID},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := st.DeleteCategory(ctx, owner.ID, source.ID, test.targetID); !errors.Is(err, ErrInvalidReassignment) {
				t.Fatalf("delete category error = %v, want ErrInvalidReassignment", err)
			}
		})
	}

	if err := st.DeleteCategory(ctx, other.ID, source.ID, nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("other user's delete error = %v, want ErrNotFound", err)
	}

	categories, err := st.ListCategories(ctx, owner.ID, budget.ID)
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	if len(categories) != 1 || categories[0].ID != source.ID {
		t.Fatalf("categories after invalid reassignments = %#v, want source category", categories)
	}
}
