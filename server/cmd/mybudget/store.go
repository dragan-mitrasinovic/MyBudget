package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// User is a registered MyBudget account. The Google sub and created_at are
// internal and never serialized to clients.
type User struct {
	ID        int64  `json:"id"`
	GoogleSub string `json:"-"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	CreatedAt string `json:"-"`
}

// Store is the thin data layer over SQLite; all database queries live here.
type Store struct {
	db *sql.DB
}

// OpenStore opens (creating if needed) the SQLite database and applies the schema.
func OpenStore(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) migrate() error {
	const usersTable = `
		CREATE TABLE IF NOT EXISTS users (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			google_sub  TEXT NOT NULL UNIQUE,
			email       TEXT NOT NULL,
			name        TEXT NOT NULL,
			picture     TEXT NOT NULL,
			created_at  TEXT NOT NULL
		)`

	if _, err := s.db.Exec(usersTable); err != nil {
		return err
	}

	const sessionsTable = `
		CREATE TABLE IF NOT EXISTS sessions (
			id          TEXT PRIMARY KEY,
			user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at  TEXT NOT NULL,
			expires_at  TEXT NOT NULL
		)`

	if _, err := s.db.Exec(sessionsTable); err != nil {
		return err
	}

	const budgetsTable = `
		CREATE TABLE IF NOT EXISTS budgets (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name         TEXT NOT NULL,
			currency     TEXT NOT NULL,
			target_minor INTEGER,
			archived     INTEGER NOT NULL DEFAULT 0,
			created_at   TEXT NOT NULL,
			updated_at   TEXT NOT NULL
		)`

	if _, err := s.db.Exec(budgetsTable); err != nil {
		return err
	}

	// Categories are per-budget. Each budget owns exactly one is_system=1 row
	// ("Uncategorized"); the UNIQUE(budget_id, name) guard also blocks duplicate names.
	const categoriesTable = `
		CREATE TABLE IF NOT EXISTS categories (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			budget_id  INTEGER NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
			name       TEXT NOT NULL,
			is_system  INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			UNIQUE(budget_id, name)
		)`

	if _, err := s.db.Exec(categoriesTable); err != nil {
		return err
	}

	// Expenses hang off a category (and reach their budget through it — no budget_id
	// column). Deleting a budget cascades budgets -> categories -> expenses.
	const expensesTable = `
		CREATE TABLE IF NOT EXISTS expenses (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			category_id  INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			amount_minor INTEGER NOT NULL,
			note         TEXT NOT NULL,
			spent_at     TEXT NOT NULL,
			created_at   TEXT NOT NULL,
			updated_at   TEXT NOT NULL
		)`

	_, err := s.db.Exec(expensesTable)
	return err
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
