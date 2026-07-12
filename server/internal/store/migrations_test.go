package store

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"testing/fstest"
)

func TestOpenAppliesMigrations(t *testing.T) {
	st := openTestStore(t)

	if version := readDatabaseVersion(t, st.db); version != 3 {
		t.Fatalf("database version = %d, want 3", version)
	}

	for _, table := range []string{"users", "sessions", "budgets", "categories", "expenses", "dashboard_layouts"} {
		if !tableExists(t, st.db, table) {
			t.Errorf("table %q does not exist", table)
		}
	}

	if !columnExists(t, st.db, "expenses", "description") {
		t.Error("expenses.description column does not exist")
	}

	for _, column := range []string{"color_key", "color_hex"} {
		if !columnExists(t, st.db, "categories", column) {
			t.Errorf("categories.%s column does not exist", column)
		}
	}
}

func TestOpenAdoptsExistingUnversionedSchemaWithoutDataLoss(t *testing.T) {
	path := testDatabasePath(t)
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open unversioned database: %v", err)
	}

	migrations, err := loadMigrations(embeddedMigrations)
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}

	if _, err := db.Exec(migrations[0].sql); err != nil {
		t.Fatalf("create unversioned schema: %v", err)
	}

	const insertUser = `
		INSERT INTO users (google_sub, email, name, picture, created_at)
		VALUES ('existing-user', 'existing@example.com', 'Existing User', '', '2026-07-11T00:00:00Z')`
	if _, err := db.Exec(insertUser); err != nil {
		t.Fatalf("insert existing user: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("close unversioned database: %v", err)
	}

	st, err := Open(path)
	if err != nil {
		t.Fatalf("reopen unversioned store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if version := readDatabaseVersion(t, st.db); version != 3 {
		t.Fatalf("database version = %d, want 3", version)
	}

	var userCount int
	if err := st.db.QueryRow("SELECT COUNT(*) FROM users WHERE google_sub = 'existing-user'").Scan(&userCount); err != nil {
		t.Fatalf("count existing user: %v", err)
	}

	if userCount != 1 {
		t.Fatalf("existing user count = %d, want 1", userCount)
	}
}

func TestOpenAlreadyMigratedDatabaseIsNoOp(t *testing.T) {
	path := testDatabasePath(t)
	st, err := Open(path)
	if err != nil {
		t.Fatalf("open initial store: %v", err)
	}

	const insertUser = `
		INSERT INTO users (google_sub, email, name, picture, created_at)
		VALUES ('persisted-user', 'persisted@example.com', 'Persisted User', '', '2026-07-11T00:00:00Z')`
	if _, err := st.db.Exec(insertUser); err != nil {
		t.Fatalf("insert persisted user: %v", err)
	}

	if err := st.Close(); err != nil {
		t.Fatalf("close initial store: %v", err)
	}

	st, err = Open(path)
	if err != nil {
		t.Fatalf("reopen migrated store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if version := readDatabaseVersion(t, st.db); version != 3 {
		t.Fatalf("database version = %d, want 3", version)
	}

	var userCount int
	if err := st.db.QueryRow("SELECT COUNT(*) FROM users WHERE google_sub = 'persisted-user'").Scan(&userCount); err != nil {
		t.Fatalf("count persisted user: %v", err)
	}

	if userCount != 1 {
		t.Fatalf("persisted user count = %d, want 1", userCount)
	}
}

func TestCategoryColorMigrationBackfillsExistingCategories(t *testing.T) {
	db := openTestDatabase(t)
	migrations, err := loadMigrations(embeddedMigrations)
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}

	for _, migration := range migrations[:2] {
		if err := applyMigration(t.Context(), db, migration); err != nil {
			t.Fatalf("apply migration %s: %v", migration.name, err)
		}
	}

	const insert = `
		INSERT INTO users (id, google_sub, email, name, picture, created_at)
		VALUES (1, 'existing-user', 'existing@example.com', 'Existing User', '', '2026-07-11T00:00:00Z');
		INSERT INTO budgets (id, user_id, name, currency, created_at, updated_at)
		VALUES (1, 1, 'Household', 'EUR', '2026-07-11T00:00:00Z', '2026-07-11T00:00:00Z');
		INSERT INTO categories (id, budget_id, name, is_system, created_at)
		VALUES
			(1, 1, 'Uncategorized', 1, '2026-07-11T00:00:00Z'),
			(2, 1, 'Transport', 0, '2026-07-11T00:00:00Z'),
			(3, 1, 'Food', 0, '2026-07-11T00:00:00Z');`
	if _, err := db.Exec(insert); err != nil {
		t.Fatalf("insert existing categories: %v", err)
	}

	if err := applyMigration(t.Context(), db, migrations[2]); err != nil {
		t.Fatalf("apply category color migration: %v", err)
	}

	rows, err := db.Query(`SELECT name, color_key, color_hex FROM categories ORDER BY id`)
	if err != nil {
		t.Fatalf("query migrated categories: %v", err)
	}
	defer func() { _ = rows.Close() }()

	wantColors := map[string]string{
		"Uncategorized": "neutral",
		"Food":          "chart-1",
		"Transport":     "chart-2",
	}
	for rows.Next() {
		var name, colorKey string
		var colorHex *string
		if err := rows.Scan(&name, &colorKey, &colorHex); err != nil {
			t.Fatalf("scan migrated category: %v", err)
		}

		if colorKey != wantColors[name] || colorHex != nil {
			t.Errorf("category %q color = %q, %v; want %q, nil", name, colorKey, colorHex, wantColors[name])
		}
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterate migrated categories: %v", err)
	}
}

func TestFailedMigrationRollsBackAndDoesNotAdvanceVersion(t *testing.T) {
	db := openTestDatabase(t)
	migrationFS := fstest.MapFS{
		"migrations/0001_create_stable.sql": {
			Data: []byte("CREATE TABLE stable (id INTEGER PRIMARY KEY);"),
		},
		"migrations/0002_fail.sql": {
			Data: []byte(`
				CREATE TABLE rolled_back (id INTEGER PRIMARY KEY);
				INSERT INTO table_that_does_not_exist (id) VALUES (1);
			`),
		},
	}

	err := applyMigrations(context.Background(), db, migrationFS)
	if err == nil {
		t.Fatal("applyMigrations() error = nil, want migration failure")
	}

	if version := readDatabaseVersion(t, db); version != 1 {
		t.Fatalf("database version = %d, want 1", version)
	}

	if !tableExists(t, db, "stable") {
		t.Error("stable table was not committed")
	}

	if tableExists(t, db, "rolled_back") {
		t.Error("failed migration was not rolled back")
	}
}

func TestDatabaseNewerThanMigrationsIsRejected(t *testing.T) {
	db := openTestDatabase(t)
	if _, err := db.Exec("PRAGMA user_version = 2"); err != nil {
		t.Fatalf("set future database version: %v", err)
	}

	migrationFS := fstest.MapFS{
		"migrations/0001_initial.sql": {
			Data: []byte("CREATE TABLE example (id INTEGER PRIMARY KEY);"),
		},
	}

	err := applyMigrations(context.Background(), db, migrationFS)
	if err == nil {
		t.Fatal("applyMigrations() error = nil, want future-version error")
	}
}

func TestMigrationVersionsMustBeContiguous(t *testing.T) {
	migrationFS := fstest.MapFS{
		"migrations/0001_initial.sql": {
			Data: []byte("SELECT 1;"),
		},
		"migrations/0003_gap.sql": {
			Data: []byte("SELECT 1;"),
		},
	}

	if _, err := loadMigrations(migrationFS); err == nil {
		t.Fatal("loadMigrations() error = nil, want version-gap error")
	}
}

func openTestStore(t *testing.T) *Store {
	t.Helper()

	st, err := Open(testDatabasePath(t))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	return st
}

func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", testDatabasePath(t))
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Fatalf("ping database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return db
}

func testDatabasePath(t *testing.T) string {
	t.Helper()
	return t.TempDir() + "/test.db"
}

func readDatabaseVersion(t *testing.T, db *sql.DB) int {
	t.Helper()

	version, err := databaseVersion(context.Background(), db)
	if err != nil {
		t.Fatalf("read database version: %v", err)
	}

	return version
}

func tableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()

	var count int
	const query = `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`
	if err := db.QueryRow(query, table).Scan(&count); err != nil {
		t.Fatalf("check table %q: %v", table, err)
	}

	return count == 1
}

func columnExists(t *testing.T, db *sql.DB, table, column string) bool {
	t.Helper()

	var count int
	const query = `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`
	if err := db.QueryRow(query, table, column).Scan(&count); err != nil {
		t.Fatalf("check column %q.%q: %v", table, column, err)
	}

	return count == 1
}
