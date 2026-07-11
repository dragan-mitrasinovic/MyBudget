package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"
)

// CreateSession issues a new opaque session for the user and returns its id.
func (s *Store) CreateSession(ctx context.Context, userID int64, ttl time.Duration) (string, error) {
	const query = `
		INSERT INTO sessions (id, user_id, created_at, expires_at)
		VALUES (?, ?, ?, ?)`

	id, err := newSessionID()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx, query, id, userID, now.Format(time.RFC3339), now.Add(ttl).Format(time.RFC3339))
	if err != nil {
		return "", err
	}

	return id, nil
}

// GetAndRenewSessionUser returns the user for a valid session and renews sessions that have
// passed the renewal interval. Missing and expired sessions return a nil user; expired rows are
// deleted best-effort.
func (s *Store) GetAndRenewSessionUser(
	ctx context.Context,
	sessionID string,
	ttl time.Duration,
	renewalInterval time.Duration,
) (*User, bool, error) {
	if ttl <= 0 {
		return nil, false, fmt.Errorf("session TTL must be positive")
	}

	if renewalInterval <= 0 || renewalInterval >= ttl {
		return nil, false, fmt.Errorf("session renewal interval must be positive and less than TTL")
	}

	const query = `
		SELECT u.id, u.google_sub, u.email, u.name, u.picture, u.created_at, s.expires_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = ?`

	u := &User{}
	var storedExpiresAt string
	err := s.db.QueryRowContext(ctx, query, sessionID).
		Scan(&u.ID, &u.GoogleSub, &u.Email, &u.Name, &u.Picture, &u.CreatedAt, &storedExpiresAt)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	expiresAt, err := time.Parse(time.RFC3339, storedExpiresAt)
	if err != nil {
		return nil, false, err
	}

	now := time.Now().UTC()
	if !now.Before(expiresAt) {
		_ = s.DeleteSession(ctx, sessionID)
		return nil, false, nil
	}

	if expiresAt.Sub(now) > ttl-renewalInterval {
		return u, false, nil
	}

	const renewQuery = `
		UPDATE sessions
		SET expires_at = ?
		WHERE id = ? AND expires_at = ? AND expires_at > ?`
	newExpiresAt := now.Add(ttl).Format(time.RFC3339)
	result, err := s.db.ExecContext(
		ctx,
		renewQuery,
		newExpiresAt,
		sessionID,
		storedExpiresAt,
		now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, false, err
	}

	return u, rowsAffected == 1, nil
}

// DeleteSession removes an opaque session id.
func (s *Store) DeleteSession(ctx context.Context, sessionID string) error {
	const query = `
		DELETE FROM sessions
		WHERE id = ?`

	_, err := s.db.ExecContext(ctx, query, sessionID)
	return err
}

func newSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
