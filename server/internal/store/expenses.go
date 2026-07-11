package store

import (
	"context"
	"database/sql"
	"time"
)

// Expense is a single dated spend within a category. The owning budget is reached through
// the category (there is no budget_id column) -- see ADR-0011.
type Expense struct {
	ID          int64  `json:"id"`
	CategoryID  int64  `json:"categoryId"`
	AmountMinor int64  `json:"amountMinor"`
	Note        string `json:"note"`
	SpentAt     string `json:"spentAt"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ExpenseFilter optionally narrows a listing by spent-at date range and/or category.
type ExpenseFilter struct {
	From       string
	To         string
	CategoryID *int64
}

// ListExpenses returns a budget's expenses (newest first) scoped to the owner, optionally
// filtered. It returns ErrNotFound if the budget is not the caller's.
func (s *Store) ListExpenses(ctx context.Context, userID, budgetID int64, filter ExpenseFilter) ([]Expense, error) {
	owned, err := s.isBudgetOwnedByUser(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}

	if !owned {
		return nil, ErrNotFound
	}

	query := `
		SELECT e.id, e.category_id, e.amount_minor, e.note, e.spent_at, e.created_at, e.updated_at
		FROM expenses e
		JOIN categories c ON c.id = e.category_id
		WHERE c.budget_id = ?`
	args := []any{budgetID}

	if filter.From != "" {
		query += " AND e.spent_at >= ?"
		args = append(args, filter.From)
	}

	if filter.To != "" {
		query += " AND e.spent_at <= ?"
		args = append(args, filter.To)
	}

	if filter.CategoryID != nil {
		query += " AND e.category_id = ?"
		args = append(args, *filter.CategoryID)
	}

	query += " ORDER BY e.spent_at DESC, e.id DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	expenses := []Expense{}
	for rows.Next() {
		var e Expense
		if err := rows.Scan(&e.ID, &e.CategoryID, &e.AmountMinor, &e.Note, &e.SpentAt, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}

		expenses = append(expenses, e)
	}

	return expenses, rows.Err()
}

// CreateExpense adds an expense to a category that belongs to a budget the caller owns.
func (s *Store) CreateExpense(ctx context.Context, userID, budgetID, categoryID, amountMinor int64, note, spentAt string) (*Expense, error) {
	if err := s.validateCategoryForOwnedBudget(ctx, userID, budgetID, categoryID); err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO expenses (category_id, amount_minor, note, spent_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, query, categoryID, amountMinor, note, spentAt, now, now)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.getExpenseByID(ctx, id)
}

// UpdateExpense edits an expense the caller owns; the new category must be in the same
// budget as the expense.
func (s *Store) UpdateExpense(ctx context.Context, userID, expenseID, categoryID, amountMinor int64, note, spentAt string) (*Expense, error) {
	budgetID, err := s.getExpenseBudgetIDForUser(ctx, userID, expenseID)
	if err != nil {
		return nil, err
	}

	if err := s.validateCategoryForOwnedBudget(ctx, userID, budgetID, categoryID); err != nil {
		return nil, err
	}

	const query = `
		UPDATE expenses
		SET category_id = ?, amount_minor = ?, note = ?, spent_at = ?, updated_at = ?
		WHERE id = ?`

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, query, categoryID, amountMinor, note, spentAt, now, expenseID); err != nil {
		return nil, err
	}

	return s.getExpenseByID(ctx, expenseID)
}

// DeleteExpense removes an expense the caller owns, reporting ErrNotFound otherwise.
func (s *Store) DeleteExpense(ctx context.Context, userID, expenseID int64) error {
	const query = `
		DELETE FROM expenses
		WHERE id = ?
		  AND category_id IN (
			SELECT c.id FROM categories c
			JOIN budgets b ON b.id = c.budget_id
			WHERE b.user_id = ?
		  )`

	res, err := s.db.ExecContext(ctx, query, expenseID, userID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrNotFound
	}

	return nil
}

// validateCategoryForOwnedBudget verifies the budget belongs to the user and the category is
// in it, returning ErrNotFound or ErrInvalidCategory respectively.
func (s *Store) validateCategoryForOwnedBudget(ctx context.Context, userID, budgetID, categoryID int64) error {
	owned, err := s.isBudgetOwnedByUser(ctx, userID, budgetID)
	if err != nil {
		return err
	}

	if !owned {
		return ErrNotFound
	}

	var one int
	err = s.db.QueryRowContext(ctx, `SELECT 1 FROM categories WHERE id = ? AND budget_id = ?`, categoryID, budgetID).
		Scan(&one)
	if err == sql.ErrNoRows {
		return ErrInvalidCategory
	}

	if err != nil {
		return err
	}

	return nil
}

// getExpenseBudgetIDForUser returns the budget ID of an expense owned by the user, or ErrNotFound.
func (s *Store) getExpenseBudgetIDForUser(ctx context.Context, userID, expenseID int64) (int64, error) {
	const query = `
		SELECT c.budget_id
		FROM expenses e
		JOIN categories c ON c.id = e.category_id
		JOIN budgets b ON b.id = c.budget_id
		WHERE e.id = ? AND b.user_id = ?`

	var budgetID int64
	err := s.db.QueryRowContext(ctx, query, expenseID, userID).Scan(&budgetID)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}

	if err != nil {
		return 0, err
	}

	return budgetID, nil
}

// getExpenseByID loads a full expense row by ID (ownership is checked by the caller).
func (s *Store) getExpenseByID(ctx context.Context, id int64) (*Expense, error) {
	const query = `
		SELECT id, category_id, amount_minor, note, spent_at, created_at, updated_at
		FROM expenses
		WHERE id = ?`

	var e Expense
	err := s.db.QueryRowContext(ctx, query, id).
		Scan(&e.ID, &e.CategoryID, &e.AmountMinor, &e.Note, &e.SpentAt, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &e, nil
}
