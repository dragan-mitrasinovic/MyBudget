package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const migrationsDirectory = "migrations"

var migrationFilenamePattern = regexp.MustCompile(`^(\d{4})_[a-z0-9_]+\.sql$`)

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

type migration struct {
	version int
	name    string
	sql     string
}

func (s *Store) applyMigrations(ctx context.Context) error {
	return applyMigrations(ctx, s.db, embeddedMigrations)
}

func applyMigrations(ctx context.Context, db *sql.DB, migrationFS fs.FS) error {
	migrations, err := loadMigrations(migrationFS)
	if err != nil {
		return err
	}

	currentVersion, err := databaseVersion(ctx, db)
	if err != nil {
		return err
	}

	latestVersion := migrations[len(migrations)-1].version
	if currentVersion > latestVersion {
		return fmt.Errorf("database version %d is newer than latest migration %d", currentVersion, latestVersion)
	}

	for _, migration := range migrations {
		if migration.version <= currentVersion {
			continue
		}

		if err := applyMigration(ctx, db, migration); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.name, err)
		}
	}

	return nil
}

func loadMigrations(migrationFS fs.FS) ([]migration, error) {
	entries, err := fs.ReadDir(migrationFS, migrationsDirectory)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, fmt.Errorf("unexpected directory in migrations: %s", entry.Name())
		}

		matches := migrationFilenamePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			return nil, fmt.Errorf("invalid migration filename %q; expected NNNN_description.sql", entry.Name())
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", matches[1], err)
		}

		expectedVersion := len(migrations) + 1
		if version != expectedVersion {
			return nil, fmt.Errorf("migration %q has version %d; expected %04d", entry.Name(), version, expectedVersion)
		}

		migrationPath := path.Join(migrationsDirectory, entry.Name())
		script, err := fs.ReadFile(migrationFS, migrationPath)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}

		if strings.TrimSpace(string(script)) == "" {
			return nil, fmt.Errorf("migration %q is empty", entry.Name())
		}

		migrations = append(migrations, migration{
			version: version,
			name:    entry.Name(),
			sql:     string(script),
		})
	}

	if len(migrations) == 0 {
		return nil, fmt.Errorf("no migrations found")
	}

	return migrations, nil
}

func databaseVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return 0, fmt.Errorf("read database version: %w", err)
	}

	return version, nil
}

func applyMigration(ctx context.Context, db *sql.DB, migration migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, migration.sql); err != nil {
		return err
	}

	if err := verifyForeignKeys(ctx, tx); err != nil {
		return err
	}

	setVersion := fmt.Sprintf("PRAGMA user_version = %d", migration.version)
	if _, err := tx.ExecContext(ctx, setVersion); err != nil {
		return fmt.Errorf("set database version: %w", err)
	}

	return tx.Commit()
}

func verifyForeignKeys(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return fmt.Errorf("check foreign keys: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		return fmt.Errorf("foreign key check failed")
	}

	return rows.Err()
}
