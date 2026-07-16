# MyBudget server

The Go HTTP API backend for MyBudget. Standard-library `net/http`, SQLite via
`modernc.org/sqlite`. See the root [AGENTS.md](../AGENTS.md) and [docs/adr/](../docs/adr/)
for project-wide context. The complete current HTTP contract is documented in
[`API.md`](API.md).

## API documentation

- **Read `API.md` before working on backend behavior.** Use it together with
  `internal/httpapi/routes.go`, the relevant handler, and the relevant store method to understand
  the existing contract and domain side effects.
- **Update `API.md` in the same change as any API change.** This includes adding or removing an
  endpoint; changing a method, path, authentication rule, request field, query parameter,
  response shape, status code, validation rule, error meaning, ordering rule, or side effect.
- Add every implemented endpoint to both the endpoint index and its detailed resource section.
  Document current behavior only, including non-obvious invariants such as category
  reassignment, ownership scoping, cascading deletion, and environment-only routes.
- Keep examples aligned with the actual JSON tags and use the exact current field casing. When
  an API change also changes an architectural decision, update `API.md` and add or supersede the
  relevant ADR.
- An endpoint change is not complete until its implementation, tests, and `API.md` agree.

## Package layout

- `cmd/mybudget/main.go` is the executable composition root: configuration, dependency
  construction, resource lifecycle, and HTTP startup only.
- `internal/httpapi/` owns HTTP routing, cookies, request validation, response mapping, and
  handlers.
- `internal/store/` owns SQLite setup, embedded ordered migrations, models, and all SQL queries.
- Dependencies point inward from `cmd/mybudget` to `httpapi` and `store`; `httpapi` may use
  `store`, but `store` must not import the HTTP package.

## Database migrations

Schema history lives in `internal/store/migrations/` and follows ADR-0016.

**Current lifecycle: active development, not prerelease.** The app has never been deployed and
has no real users, so migration history is not yet a compatibility contract. Until the owner
explicitly declares the project prerelease:

- Give database additions, such as new tables, columns, constraints, or indexes, their own
  contiguous migration.
- Fold destructive or corrective schema changes into the existing migration that defines that
  schema. Remove obsolete columns and tables from the baseline instead of adding drop migrations;
  delete an obsolete later migration when appropriate and keep the remaining sequence contiguous.
- Local development databases may need to be deleted and recreated after a rewritten migration.
  Do not add compatibility migrations solely to preserve local development data.

When the owner declares the project prerelease, this exception ends: freeze all existing
migrations and use a new forward migration for every subsequent schema change, including drops,
renames, and corrections.

- **Read all existing migrations before changing the database schema.** Do not recreate schema
  definitions in Go or rely on `CREATE TABLE IF NOT EXISTS` outside the baseline migration.
- Name new additive migrations using `NNNN_lowercase_description.sql` (for example,
  `0002_add_expense_index.sql`).
- Migrations are forward-only and run automatically during `store.Open`, one transaction per
  file. Do not put `BEGIN`, `COMMIT`, `ROLLBACK`, or `PRAGMA user_version` in migration SQL; the
  runner owns transaction and version management.
- Keep migration SQL deterministic and safe for SQLite. When SQLite requires a table rebuild,
  copy data explicitly, preserve foreign keys/indexes, and verify the resulting constraints.
- Add or update focused migration tests for fresh setup, upgrades from the previous version,
  data preservation, and failure rollback. Run `go test ./internal/store` for every migration.
- After prerelease begins, back up persistent databases with SQLite's backup mechanism before
  applying schema changes. An application binary must not be deployed unless it contains every
  migration needed by the target database.
- Update the relevant ADR when migration strategy changes. Update `API.md` only when the schema
  migration also changes observable HTTP behavior.

## Go coding style

Conventions for this folder, applied on top of `gofmt`:

- **Boolean names are predicates.** Name booleans so they read as yes/no questions, using
  prefixes such as `is`, `has`, `can`, `should`, `use`, or `enable` (for example,
  `isDevelopment` and `useSecureCookies`), rather than vague state names.

- **Function names describe their action and contract.** Lead with a specific verb such as
  `get`, `find`, `list`, `create`, `validate`, `require`, or `write`; reserve `handle` for HTTP
  handlers. Include relevant scope or return shape when it is not obvious (for example,
  `getExpenseBudgetIDForUser`). Use `find` when absence is an expected result, `get` when a
  value is expected, and compound names when a function has meaningful combined behavior
  (`normalizeAndValidate`). Keep standard initialisms capitalized (`ID`, `HTTP`, `JSON`). Avoid
  vague names that hide lookups, validation, response writes, or other significant side effects.

- **Blank line after an `if` block.** Leave one empty line after an `if` statement's
  closing `}` when more code follows in the same block. Skip it only when the next line
  is the enclosing block's own closing `}`.

  ```go
  if clientID == "" {
      log.Fatal("GOOGLE_CLIENT_ID is required")
  }

  port := envOrDefault("PORT", "8080")
  ```

## Tooling

Run from `server/` (or `make lint` / `make vet` from the repo root). See ADR-0008.

- `make dev-api` from the repository root — run the development API through the pinned Air
  tool, rebuilding and restarting it when backend source files change. Prefer this over calling
  Air directly because the Make target supplies the required development environment.
- `gofmt -w .` — formatting. The blank-line rule above is house style layered on top.
- `go vet ./...` — built-in correctness checks.
- `golangci-lint run` — meta-linter with a minimal starter set (`govet`, `staticcheck`,
  `errcheck`, `ineffassign`, `unused`) configured in [`.golangci.yml`](../.golangci.yml).
  Add linters only when they earn their keep.

## Testing

- Prefer small table-driven unit tests for validation, parsing, and middleware control flow.
- Test Store ownership and data-integrity invariants against a real SQLite database in
  `t.TempDir()`; do not mock SQL behavior that the database can verify directly.
- Use `httptest` through the complete router only when route registration, authentication,
  Origin protection, or HTTP response mapping is the behavior under test.
- Keep tests deterministic and local: no external services, wall-clock sleeps, or shared
  databases. Use `t.Helper()` and `t.Cleanup()` in repeated setup.
- Run a focused test while iterating, then `go test ./internal/store`,
  `go test ./internal/httpapi`, and finally `go test ./...`.
