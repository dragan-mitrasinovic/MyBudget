package store

import (
	"context"
	"database/sql"
	"time"
)

// uncategorizedCategoryName is the permanent, per-budget fallback category created with
// every budget. It is flagged is_system and cannot be renamed or deleted; deleting any
// other category reassigns its expenses here.
const uncategorizedCategoryName = "Uncategorized"

// Budget is a user-owned container for categories and expenses. UserID is internal and
// never serialized to clients.
type Budget struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"-"`
	Name        string `json:"name"`
	Currency    string `json:"currency"`
	TargetMinor *int64 `json:"targetMinor"`
	Archived    bool   `json:"archived"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// CreateBudget inserts a budget together with its permanent Uncategorized category in a
// single transaction, so a budget never exists without the fallback category.
func (s *Store) CreateBudget(ctx context.Context, userID int64, name, currency string, targetMinor *int64) (*Budget, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	const insertBudget = `
		INSERT INTO budgets (user_id, name, currency, target_minor, archived, created_at, updated_at)
		VALUES (?, ?, ?, ?, 0, ?, ?)`

	res, err := tx.ExecContext(ctx, insertBudget, userID, name, currency, targetMinor, now, now)
	if err != nil {
		return nil, err
	}

	budgetID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	const insertCategory = `
		INSERT INTO categories (budget_id, name, is_system, created_at)
		VALUES (?, ?, 1, ?)`

	if _, err := tx.ExecContext(ctx, insertCategory, budgetID, uncategorizedCategoryName, now); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &Budget{
		ID:          budgetID,
		UserID:      userID,
		Name:        name,
		Currency:    currency,
		TargetMinor: targetMinor,
		Archived:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// ListBudgets returns all of the user's budgets, newest first.
func (s *Store) ListBudgets(ctx context.Context, userID int64) ([]Budget, error) {
	const query = `
		SELECT id, user_id, name, currency, target_minor, archived, created_at, updated_at
		FROM budgets
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	budgets := []Budget{}
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.Currency, &b.TargetMinor, &b.Archived, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}

		budgets = append(budgets, b)
	}

	return budgets, rows.Err()
}

// GetBudget returns one budget scoped to its owner, or (nil, nil) if it does not exist or
// belongs to another user.
func (s *Store) GetBudget(ctx context.Context, userID, budgetID int64) (*Budget, error) {
	const query = `
		SELECT id, user_id, name, currency, target_minor, archived, created_at, updated_at
		FROM budgets
		WHERE id = ? AND user_id = ?`

	var b Budget
	err := s.db.QueryRowContext(ctx, query, budgetID, userID).
		Scan(&b.ID, &b.UserID, &b.Name, &b.Currency, &b.TargetMinor, &b.Archived, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &b, nil
}

// UpdateBudget replaces the editable fields of a user's budget and returns the saved row,
// or (nil, nil) if no matching budget is owned by the user.
func (s *Store) UpdateBudget(ctx context.Context, userID, budgetID int64, name, currency string, targetMinor *int64, archived bool) (*Budget, error) {
	const query = `
		UPDATE budgets
		SET name = ?, currency = ?, target_minor = ?, archived = ?, updated_at = ?
		WHERE id = ? AND user_id = ?`

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, query, name, currency, targetMinor, archived, now, budgetID, userID)
	if err != nil {
		return nil, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, nil
	}

	return s.GetBudget(ctx, userID, budgetID)
}

// DeleteBudget removes a user's budget (cascading to its categories and expenses) and
// reports whether a row was deleted.
func (s *Store) DeleteBudget(ctx context.Context, userID, budgetID int64) (bool, error) {
	const query = `
		DELETE FROM budgets
		WHERE id = ? AND user_id = ?`

	res, err := s.db.ExecContext(ctx, query, budgetID, userID)
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}
