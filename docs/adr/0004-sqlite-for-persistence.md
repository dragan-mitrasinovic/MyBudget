# 0004. SQLite for persistence

Date: 2026-06-29
Status: Accepted

## Context

The backend needs to store user accounts (and later, budget data). The app's scale is
the author and a few friends — low concurrency — and we want minimal infrastructure to
start.

## Decision

Use SQLite as the database, accessed from Go via `database/sql` with the pure-Go
`modernc.org/sqlite` driver (no CGO). Keep all DB access behind a thin data layer.

## Alternatives considered

- A separate client-server database server — production-grade concurrency, but requires
  running and operating a dedicated DB server in dev and prod from day one; unnecessary at
  our scale.

## Consequences

- (+) Zero infrastructure: the DB is a single file; `go run` just works.
- (+) Simple ops (backup = copy the file) and fast at our scale.
- (−) Single-writer model and one-machine storage; not suited to multi-server deploys.
