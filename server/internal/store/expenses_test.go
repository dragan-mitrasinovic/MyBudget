package store

import (
	"context"
	"errors"
	"testing"
)

func TestCreateExpenseRequiresCategoryInOwnedBudget(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "expense-owner")
	otherUser := createStoreTestUser(t, st, "other-user")
	firstBudget := createStoreTestBudget(t, st, user.ID, "First")
	secondBudget := createStoreTestBudget(t, st, user.ID, "Second")
	foreignBudget := createStoreTestBudget(t, st, otherUser.ID, "Foreign")
	ctx := context.Background()

	secondCategory, err := st.CreateCategory(ctx, user.ID, secondBudget.ID, "Second category")
	if err != nil {
		t.Fatalf("create second-budget category: %v", err)
	}

	if _, err := st.CreateExpense(ctx, user.ID, firstBudget.ID, secondCategory.ID, 100, "", "2026-07-11"); !errors.Is(err, ErrInvalidCategory) {
		t.Errorf("create with category from another budget error = %v, want ErrInvalidCategory", err)
	}

	foreignCategory, err := st.CreateCategory(ctx, otherUser.ID, foreignBudget.ID, "Foreign category")
	if err != nil {
		t.Fatalf("create foreign category: %v", err)
	}

	if _, err := st.CreateExpense(ctx, user.ID, foreignBudget.ID, foreignCategory.ID, 100, "", "2026-07-11"); !errors.Is(err, ErrNotFound) {
		t.Errorf("create in foreign budget error = %v, want ErrNotFound", err)
	}
}

func TestExpenseMutationsAreScopedToOwner(t *testing.T) {
	st := openTestStore(t)
	owner := createStoreTestUser(t, st, "expense-owner")
	otherUser := createStoreTestUser(t, st, "other-user")
	budget := createStoreTestBudget(t, st, owner.ID, "Household")
	ctx := context.Background()

	category, err := st.CreateCategory(ctx, owner.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	expense, err := st.CreateExpense(ctx, owner.ID, budget.ID, category.ID, 1599, "Lunch", "2026-07-11")
	if err != nil {
		t.Fatalf("create expense: %v", err)
	}

	if _, err := st.UpdateExpense(ctx, otherUser.ID, expense.ID, category.ID, 9999, "Changed", "2026-07-10"); !errors.Is(err, ErrNotFound) {
		t.Errorf("update as other user error = %v, want ErrNotFound", err)
	}

	if err := st.DeleteExpense(ctx, otherUser.ID, expense.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("delete as other user error = %v, want ErrNotFound", err)
	}

	expenses, err := st.ListExpenses(ctx, owner.ID, budget.ID, ExpenseFilter{})
	if err != nil {
		t.Fatalf("list expenses as owner: %v", err)
	}

	if len(expenses) != 1 || expenses[0].ID != expense.ID || expenses[0].AmountMinor != 1599 {
		t.Fatalf("owner's expenses after foreign mutations = %#v, want original expense", expenses)
	}
}

func TestListExpensesAppliesFiltersAndOrdering(t *testing.T) {
	st := openTestStore(t)
	user := createStoreTestUser(t, st, "expense-owner")
	budget := createStoreTestBudget(t, st, user.ID, "Household")
	ctx := context.Background()

	food, err := st.CreateCategory(ctx, user.ID, budget.ID, "Food")
	if err != nil {
		t.Fatalf("create food category: %v", err)
	}

	travel, err := st.CreateCategory(ctx, user.ID, budget.ID, "Travel")
	if err != nil {
		t.Fatalf("create travel category: %v", err)
	}

	oldFood, err := st.CreateExpense(ctx, user.ID, budget.ID, food.ID, 100, "Old food", "2026-07-09")
	if err != nil {
		t.Fatalf("create old food expense: %v", err)
	}

	newFood, err := st.CreateExpense(ctx, user.ID, budget.ID, food.ID, 200, "New food", "2026-07-11")
	if err != nil {
		t.Fatalf("create new food expense: %v", err)
	}

	travelExpense, err := st.CreateExpense(ctx, user.ID, budget.ID, travel.ID, 300, "Travel", "2026-07-10")
	if err != nil {
		t.Fatalf("create travel expense: %v", err)
	}

	all, err := st.ListExpenses(ctx, user.ID, budget.ID, ExpenseFilter{})
	if err != nil {
		t.Fatalf("list all expenses: %v", err)
	}

	if len(all) != 3 || all[0].ID != newFood.ID || all[1].ID != travelExpense.ID || all[2].ID != oldFood.ID {
		t.Fatalf("ordered expenses = %#v, want newest dates first", all)
	}

	filtered, err := st.ListExpenses(ctx, user.ID, budget.ID, ExpenseFilter{
		From:       "2026-07-10",
		To:         "2026-07-11",
		CategoryID: &food.ID,
	})
	if err != nil {
		t.Fatalf("list filtered expenses: %v", err)
	}

	if len(filtered) != 1 || filtered[0].ID != newFood.ID {
		t.Fatalf("filtered expenses = %#v, want new food expense", filtered)
	}
}
