# MyBudget server

The Go HTTP API backend for MyBudget. Standard-library `net/http`, SQLite via
`modernc.org/sqlite`. See the root [AGENTS.md](../AGENTS.md) and [docs/adr/](../docs/adr/)
for project-wide context.

## Go coding style

Conventions for this folder, applied on top of `gofmt`:

- **Blank line after an `if` block.** Leave one empty line after an `if` statement's
  closing `}` when more code follows in the same block. Skip it only when the next line
  is the enclosing block's own closing `}`.

  ```go
  if clientID == "" {
      log.Fatal("GOOGLE_CLIENT_ID is required")
  }

  port := getenv("PORT", "8080")
  ```

## Tooling

Run from `server/` (or `make lint` / `make vet` from the repo root). See ADR-0008.

- `gofmt -w .` — formatting. The blank-line rule above is house style layered on top.
- `go vet ./...` — built-in correctness checks.
- `golangci-lint run` — meta-linter with a minimal starter set (`govet`, `staticcheck`,
  `errcheck`, `ineffassign`, `unused`) configured in [`.golangci.yml`](../.golangci.yml).
  Add linters only when they earn their keep.
