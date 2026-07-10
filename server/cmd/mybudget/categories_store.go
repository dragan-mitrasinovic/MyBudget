package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// Sentinel errors from the category data methods, mapped to HTTP status codes by the
// handlers.
var (
	errNotFound       = errors.New("not found")
	errSystemCategory = errors.New("cannot modify the system category")
	errDuplicateName  = errors.New("name already exists")
)

// Category is a per-budget grouping for expenses. Every budget owns one is_system row
// ("Uncategorized") that cannot be renamed or deleted.
type Category struct {
	ID        int64  `json:"id"`
	BudgetID  int64  `json:"budgetId"`
	Name      string `json:"name"`
	IsSystem  bool   `json:"isSystem"`
	CreatedAt string `json:"createdAt"`
}

// ListCategories returns a budget's categories (user categories first, Uncategorized last)
// scoped to the owner. An empty result means the budget does not exist or is not the
// caller's — every real budget has at least its Uncategorized category.
func (s *Store) ListCategories(ctx context.Context, userID, budgetID int64) ([]Category, error) {
	const query = `
		SELECT c.id, c.budget_id, c.name, c.is_system, c.created_at
		FROM categories c
		JOIN budgets b ON b.id = c.budget_id
		WHERE c.budget_id = ? AND b.user_id = ?
		ORDER BY c.is_system ASC, c.name COLLATE NOCASE ASC`

	rows, err := s.db.QueryContext(ctx, query, budgetID, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.BudgetID, &c.Name, &c.IsSystem, &c.CreatedAt); err != nil {
			return nil, err
		}

		categories = append(categories, c)
	}

	return categories, rows.Err()
}

// CreateCategory adds a user category to a budget the caller owns. It returns errNotFound
// when the budget is not the user's, or errDuplicateName on a name clash.
func (s *Store) CreateCategory(ctx context.Context, userID, budgetID int64, name string) (*Category, error) {
	owned, err := s.budgetOwnedByUser(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}

	if !owned {
		return nil, errNotFound
	}

	const query = `
		INSERT INTO categories (budget_id, name, is_system, created_at)
		VALUES (?, ?, 0, ?)`

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, query, budgetID, name, now)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, errDuplicateName
		}

		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Category{ID: id, BudgetID: budgetID, Name: name, IsSystem: false, CreatedAt: now}, nil
}

// UpdateCategory renames a user category the caller owns. It returns errNotFound,
// errSystemCategory (the Uncategorized row is immutable), or errDuplicateName.
func (s *Store) UpdateCategory(ctx context.Context, userID, categoryID int64, name string) (*Category, error) {
	cat, err := s.categoryForUser(ctx, userID, categoryID)
	if err != nil {
		return nil, err
	}

	if cat == nil {
		return nil, errNotFound
	}

	if cat.IsSystem {
		return nil, errSystemCategory
	}

	const query = `
		UPDATE categories
		SET name = ?
		WHERE id = ?`

	if _, err := s.db.ExecContext(ctx, query, name, categoryID); err != nil {
		if isUniqueViolation(err) {
			return nil, errDuplicateName
		}

		return nil, err
	}

	cat.Name = name
	return cat, nil
}

// DeleteCategory removes a user category, first moving its expenses to the budget's
// Uncategorized category so none are lost. The system category cannot be deleted.
func (s *Store) DeleteCategory(ctx context.Context, userID, categoryID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var budgetID int64
	var isSystem bool
	const lookup = `
		SELECT c.budget_id, c.is_system
		FROM categories c
		JOIN budgets b ON b.id = c.budget_id
		WHERE c.id = ? AND b.user_id = ?`

	err = tx.QueryRowContext(ctx, lookup, categoryID, userID).Scan(&budgetID, &isSystem)
	if err == sql.ErrNoRows {
		return errNotFound
	}

	if err != nil {
		return err
	}

	if isSystem {
		return errSystemCategory
	}

	var uncategorizedID int64
	const findUncategorized = `
		SELECT id FROM categories
		WHERE budget_id = ? AND is_system = 1`

	if err := tx.QueryRowContext(ctx, findUncategorized, budgetID).Scan(&uncategorizedID); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	const reassign = `
		UPDATE expenses
		SET category_id = ?, updated_at = ?
		WHERE category_id = ?`

	if _, err := tx.ExecContext(ctx, reassign, uncategorizedID, now, categoryID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, categoryID); err != nil {
		return err
	}

	return tx.Commit()
}

// categoryForUser fetches a category joined to its owning budget, or (nil, nil) if it does
// not exist or belongs to another user.
func (s *Store) categoryForUser(ctx context.Context, userID, categoryID int64) (*Category, error) {
	const query = `
		SELECT c.id, c.budget_id, c.name, c.is_system, c.created_at
		FROM categories c
		JOIN budgets b ON b.id = c.budget_id
		WHERE c.id = ? AND b.user_id = ?`

	var c Category
	err := s.db.QueryRowContext(ctx, query, categoryID, userID).
		Scan(&c.ID, &c.BudgetID, &c.Name, &c.IsSystem, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

// budgetOwnedByUser reports whether the budget exists and belongs to the user.
func (s *Store) budgetOwnedByUser(ctx context.Context, userID, budgetID int64) (bool, error) {
	const query = `SELECT 1 FROM budgets WHERE id = ? AND user_id = ?`

	var one int
	err := s.db.QueryRowContext(ctx, query, budgetID, userID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// isUniqueViolation reports whether err is a SQLite UNIQUE constraint failure.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
