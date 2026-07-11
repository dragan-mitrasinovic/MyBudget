# 0013. Embedded ordered SQLite migrations

Date: 2026-07-11
Status: Superseded by ADR-0016

## Context

The store previously created tables at startup with `CREATE TABLE IF NOT EXISTS` statements in
Go. That initialized fresh databases but could not evolve existing columns, constraints,
indexes, or data, and it recorded no schema version. MyBudget has one SQLite database and one
server binary, so it needs deterministic schema history without requiring a separate deployment
tool or migration service.

The existing development database already contains the current schema but has SQLite
`user_version = 0`; the migration system must adopt it without losing data.

## Decision

Store forward-only SQL migrations under `server/internal/store/migrations/`, embedded into the
server binary with `go:embed`. Name files with contiguous four-digit versions:

```text
0001_initial_schema.sql
0002_example_change.sql
```

At `store.Open`, the migration runner:

1. Loads and validates all embedded migration filenames.
2. Requires versions to start at `0001` and remain contiguous.
3. Reads the current version from SQLite's `PRAGMA user_version`.
4. Rejects a database newer than the latest migration known to the binary.
5. Applies each pending migration in order, one transaction per migration.
6. Runs `PRAGMA foreign_key_check` before committing each migration.
7. Updates `PRAGMA user_version` in the same transaction as the schema/data change.
8. Fails server startup if any migration fails.

Migrations are immutable once applied. Schema changes add the next numbered migration; they do
not edit or delete an earlier one. Migrations are forward-only: recovery from a bad production
migration uses a SQLite backup or a new corrective migration, not an automatic down migration.

`0001_initial_schema.sql` is the baseline and uses `CREATE TABLE IF NOT EXISTS`. On a fresh
database it creates the complete initial schema. On the existing unversioned database it leaves
the already matching tables and data intact, verifies foreign keys, and records version 1. The
current development schema was inspected and integrity-checked before this adoption.

## Alternatives considered

- **Continue `CREATE TABLE IF NOT EXISTS` in Go** — simple for fresh databases, but cannot
  perform ordered upgrades or record which schema an existing database has.
- **A migration library such as Goose or golang-migrate** — provides CLIs, down migrations, and
  broader database support, but adds a dependency and deployment workflow that are unnecessary
  for one embedded SQLite database and one server process.
- **A `schema_migrations` table** — supports richer metadata such as names, timestamps, and
  checksums. `PRAGMA user_version` is sufficient for the current single linear migration stream
  and avoids another internal table.
- **Go functions for every migration** — useful when complex data transformation needs Go, but
  makes ordinary schema history harder to inspect than standalone SQL. A future migration can
  add narrowly scoped Go assistance if SQL is genuinely insufficient.
- **Down migrations** — convenient in development but risky for destructive production changes.
  Forward fixes plus verified backups are simpler and safer at this scale.

## Consequences

- (+) Fresh and existing databases follow one deterministic, versioned startup path.
- (+) Migration SQL ships inside the binary and cannot be omitted during deployment.
- (+) Each migration and its version update are atomic; failures roll back and stop startup.
- (+) Future constraints, indexes, and data transformations can be introduced deliberately.
- (+) The server rejects databases created by a newer binary instead of running against an
  unknown schema.
- (-) Applied migrations must remain immutable and version numbers must stay contiguous.
- (-) `user_version` records only the latest integer, not migration checksums or timestamps.
- (-) Rollback requires a backup or a forward corrective migration.
- (-) The baseline can safely adopt only the already inspected schema it was designed for;
  arbitrary unversioned database shapes are not supported.
