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
  Document current behavior only, including non-obvious invariants such as Uncategorized
  category creation, reassignment, ownership scoping, cascading deletion, and environment-only
  routes.
- Keep examples aligned with the actual JSON tags and use the exact current field casing. When
  an API change also changes an architectural decision, update `API.md` and add or supersede the
  relevant ADR.
- An endpoint change is not complete until its implementation, tests, and `API.md` agree.

## Package layout

- `cmd/mybudget/main.go` is the executable composition root: configuration, dependency
  construction, resource lifecycle, and HTTP startup only.
- `internal/httpapi/` owns HTTP routing, cookies, request validation, response mapping, and
  handlers.
- `internal/store/` owns SQLite setup, schema creation, models, and all SQL queries.
- Dependencies point inward from `cmd/mybudget` to `httpapi` and `store`; `httpapi` may use
  `store`, but `store` must not import the HTTP package.

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

- `gofmt -w .` — formatting. The blank-line rule above is house style layered on top.
- `go vet ./...` — built-in correctness checks.
- `golangci-lint run` — meta-linter with a minimal starter set (`govet`, `staticcheck`,
  `errcheck`, `ineffassign`, `unused`) configured in [`.golangci.yml`](../.golangci.yml).
  Add linters only when they earn their keep.
