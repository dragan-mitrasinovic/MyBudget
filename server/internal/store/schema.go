package store

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

	// Expenses hang off a category (and reach their budget through it -- no budget_id
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
