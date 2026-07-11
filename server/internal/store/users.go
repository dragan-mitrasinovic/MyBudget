package store

import (
	"context"
	"time"
)

// User is a registered MyBudget account. GoogleSub and CreatedAt are internal and never
// serialized to clients.
type User struct {
	ID        int64  `json:"id"`
	GoogleSub string `json:"-"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	CreatedAt string `json:"-"`
}

// UpsertUser creates the user on first login or refreshes their profile on return,
// keyed by the stable Google sub, and returns the stored row.
func (s *Store) UpsertUser(ctx context.Context, sub, email, name, picture string) (*User, error) {
	const upsertQuery = `
		INSERT INTO users (google_sub, email, name, picture, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(google_sub) DO UPDATE SET
			email = excluded.email,
			name = excluded.name,
			picture = excluded.picture`

	const selectQuery = `
		SELECT id, google_sub, email, name, picture, created_at
		FROM users
		WHERE google_sub = ?`

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, upsertQuery, sub, email, name, picture, now); err != nil {
		return nil, err
	}

	u := &User{}
	err := s.db.QueryRowContext(ctx, selectQuery, sub).
		Scan(&u.ID, &u.GoogleSub, &u.Email, &u.Name, &u.Picture, &u.CreatedAt)
	if err != nil {
		return nil, err
	}

	return u, nil
}
