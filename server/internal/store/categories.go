package store

import (
	"context"
	"database/sql"
	"time"
)

var categoryChartColorKeys = []string{
	"chart-1",
	"chart-2",
	"chart-3",
	"chart-4",
	"chart-5",
	"chart-6",
	"chart-7",
	"chart-8",
	"chart-9",
	"chart-10",
	"chart-11",
	"chart-12",
}

const categoryNeutralColorKey = "neutral"

// Category is a user-defined per-budget grouping for expenses.
type Category struct {
	ID        int64   `json:"id"`
	BudgetID  int64   `json:"budgetId"`
	Name      string  `json:"name"`
	ColorKey  string  `json:"colorKey"`
	ColorHex  *string `json:"colorHex"`
	CreatedAt string  `json:"createdAt"`
}

// ListCategories returns an owned budget's categories alphabetically. An owned budget may
// have no categories; an absent or unowned budget returns ErrNotFound.
func (s *Store) ListCategories(ctx context.Context, userID, budgetID int64) ([]Category, error) {
	owned, err := s.isBudgetOwnedByUser(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}

	if !owned {
		return nil, ErrNotFound
	}

	const query = `
		SELECT id, budget_id, name, color_key, color_hex, created_at
		FROM categories c
		WHERE budget_id = ?
		ORDER BY name COLLATE NOCASE ASC, id ASC`

	rows, err := s.db.QueryContext(ctx, query, budgetID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.BudgetID, &c.Name, &c.ColorKey, &c.ColorHex, &c.CreatedAt); err != nil {
			return nil, err
		}

		categories = append(categories, c)
	}

	return categories, rows.Err()
}

// CreateCategory adds a user category to a budget the caller owns. It returns ErrNotFound
// when the budget is not the user's, or ErrDuplicateName on a name clash.
func (s *Store) CreateCategory(ctx context.Context, userID, budgetID int64, name string) (*Category, error) {
	owned, err := s.isBudgetOwnedByUser(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}

	if !owned {
		return nil, ErrNotFound
	}

	colorKey, err := s.nextCategoryColorKey(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO categories (budget_id, name, color_key, created_at)
		VALUES (?, ?, ?, ?)`

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, query, budgetID, name, colorKey, now)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateName
		}

		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Category{
		ID:        id,
		BudgetID:  budgetID,
		Name:      name,
		ColorKey:  colorKey,
		CreatedAt: now,
	}, nil
}

// UpdateCategory renames a category the caller owns. It returns ErrNotFound or
// ErrDuplicateName.
func (s *Store) UpdateCategory(ctx context.Context, userID, categoryID int64, name string) (*Category, error) {
	cat, err := s.findCategoryOwnedByUser(ctx, userID, categoryID)
	if err != nil {
		return nil, err
	}

	if cat == nil {
		return nil, ErrNotFound
	}

	const query = `
		UPDATE categories
		SET name = ?
		WHERE id = ?`

	if _, err := s.db.ExecContext(ctx, query, name, categoryID); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateName
		}

		return nil, err
	}

	cat.Name = name
	return cat, nil
}

// UpdateCategoryColor replaces one owned category's chart color.
func (s *Store) UpdateCategoryColor(ctx context.Context, userID, categoryID int64, colorKey string, colorHex *string) (*Category, error) {
	cat, err := s.findCategoryOwnedByUser(ctx, userID, categoryID)
	if err != nil {
		return nil, err
	}

	if cat == nil {
		return nil, ErrNotFound
	}

	const query = `
		UPDATE categories
		SET color_key = ?, color_hex = ?
		WHERE id = ?`

	if _, err := s.db.ExecContext(ctx, query, colorKey, colorHex, categoryID); err != nil {
		return nil, err
	}

	cat.ColorKey = colorKey
	cat.ColorHex = colorHex
	return cat, nil
}

// DeleteCategory removes an owned category. A non-empty category requires another category
// in the same budget as its reassignment target; the move and deletion are atomic.
func (s *Store) DeleteCategory(ctx context.Context, userID, categoryID int64, reassignToCategoryID *int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var budgetID int64
	const lookup = `
		SELECT c.budget_id
		FROM categories c
		JOIN budgets b ON b.id = c.budget_id
		WHERE c.id = ? AND b.user_id = ?`

	err = tx.QueryRowContext(ctx, lookup, categoryID, userID).Scan(&budgetID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}

	if err != nil {
		return err
	}

	if reassignToCategoryID == nil {
		const deleteEmpty = `
			DELETE FROM categories
			WHERE id = ?
			  AND NOT EXISTS (
				SELECT 1 FROM expenses WHERE category_id = ?
			  )`

		res, err := tx.ExecContext(ctx, deleteEmpty, categoryID, categoryID)
		if err != nil {
			return err
		}

		deleted, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if deleted == 0 {
			return ErrCategoryNotEmpty
		}
	} else {
		targetID := *reassignToCategoryID
		if targetID <= 0 || targetID == categoryID {
			return ErrInvalidReassignment
		}

		var one int
		const validateTarget = `SELECT 1 FROM categories WHERE id = ? AND budget_id = ?`
		err := tx.QueryRowContext(ctx, validateTarget, targetID, budgetID).Scan(&one)
		if err == sql.ErrNoRows {
			return ErrInvalidReassignment
		}

		if err != nil {
			return err
		}

		now := time.Now().UTC().Format(time.RFC3339)
		const reassign = `
			UPDATE expenses
			SET category_id = ?, updated_at = ?
			WHERE category_id = ?`

		if _, err := tx.ExecContext(ctx, reassign, targetID, now, categoryID); err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, categoryID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// findCategoryOwnedByUser fetches a category joined to its owning budget, or (nil, nil) if it does
// not exist or belongs to another user.
func (s *Store) findCategoryOwnedByUser(ctx context.Context, userID, categoryID int64) (*Category, error) {
	const query = `
		SELECT c.id, c.budget_id, c.name, c.color_key, c.color_hex, c.created_at
		FROM categories c
		JOIN budgets b ON b.id = c.budget_id
		WHERE c.id = ? AND b.user_id = ?`

	var c Category
	err := s.db.QueryRowContext(ctx, query, categoryID, userID).
		Scan(&c.ID, &c.BudgetID, &c.Name, &c.ColorKey, &c.ColorHex, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (s *Store) nextCategoryColorKey(ctx context.Context, budgetID int64) (string, error) {
	const query = `
		SELECT color_key
		FROM categories
		WHERE budget_id = ?`

	rows, err := s.db.QueryContext(ctx, query, budgetID)
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()

	usedKeys := make(map[string]bool, len(categoryChartColorKeys))
	categoryCount := 0
	for rows.Next() {
		var colorKey string
		if err := rows.Scan(&colorKey); err != nil {
			return "", err
		}

		usedKeys[colorKey] = true
		categoryCount++
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	for _, colorKey := range categoryChartColorKeys {
		if !usedKeys[colorKey] {
			return colorKey, nil
		}
	}

	return categoryChartColorKeys[categoryCount%len(categoryChartColorKeys)], nil
}

// IsValidCategoryColorKey reports whether a key belongs to the persisted chart palette.
func IsValidCategoryColorKey(colorKey string) bool {
	if colorKey == categoryNeutralColorKey {
		return true
	}

	for _, validKey := range categoryChartColorKeys {
		if colorKey == validKey {
			return true
		}
	}

	return false
}

// isBudgetOwnedByUser reports whether the budget exists and belongs to the user.
func (s *Store) isBudgetOwnedByUser(ctx context.Context, userID, budgetID int64) (bool, error) {
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
