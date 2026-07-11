package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
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

// GetSessionUser returns the user for a valid, unexpired session, or (nil, nil) if
// the session is missing or expired (expired sessions are deleted best-effort).
func (s *Store) GetSessionUser(ctx context.Context, sessionID string) (*User, error) {
	const query = `
		SELECT u.id, u.google_sub, u.email, u.name, u.picture, u.created_at, s.expires_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = ?`

	u := &User{}
	var expiresAt string
	err := s.db.QueryRowContext(ctx, query, sessionID).
		Scan(&u.ID, &u.GoogleSub, &u.Email, &u.Name, &u.Picture, &u.CreatedAt, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	exp, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return nil, err
	}

	if time.Now().After(exp) {
		_ = s.DeleteSession(ctx, sessionID)
		return nil, nil
	}

	return u, nil
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
