# 0004. SQLite for persistence

Date: 2026-06-29
Status: Accepted

## Context

The backend needs to store user accounts (and later, budget data). The app's scale is
the author and a few friends — low concurrency — and we want minimal infrastructure to
start.

## Decision

Use SQLite as the database, accessed from Go via `database/sql` with the pure-Go
`modernc.org/sqlite` driver (no CGO). Keep all DB access behind a thin data layer so a
future swap to PostgreSQL is localized.

## Alternatives considered

- PostgreSQL — production-grade concurrency and scales past "friends" with no migration,
  but requires running a DB server in dev and prod from day one; overkill for now.

## Consequences

- (+) Zero infrastructure: the DB is a single file; `go run` just works.
- (+) Simple ops (backup = copy the file) and fast at our scale.
- (−) Single-writer model and one-machine storage; not suited to multi-server deploys.
- (−) If we outgrow it, we migrate to PostgreSQL (contained by the thin data layer).
