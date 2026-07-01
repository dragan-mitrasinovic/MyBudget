# MyBudget

MyBudget is a personal expense-tracking and budgeting app for the author and a few
friends, so it is multi-user from the start. The web client is an Angular + Angular
Material single-page app; a Go HTTP API (`server/`) and a future mobile client are
planned. See `docs/adr/` for the decisions behind these choices.

## Architecture decisions

Significant architectural and tech-stack decisions are recorded as ADRs (Architecture
Decision Records) in `docs/adr/` — one Markdown file per decision.

**Whenever an architectural or tech-stack decision is made, add a new ADR.** Copy
`docs/adr/template.md`, give it the next number, and fill in the context, decision,
alternatives considered, and consequences. Decisions are immutable: to change or reverse
one, add a new ADR and mark the old one `Superseded by ADR-NNNN` — don't edit or delete
the original.

## Coding conventions

Language-specific coding conventions live in per-folder `AGENTS.md` files next to the
code they govern — e.g. [`server/AGENTS.md`](server/AGENTS.md) for Go and
[`web/AGENTS.md`](web/AGENTS.md) for Angular/TypeScript. Add new conventions to the
relevant folder's file.

## Code quality

Formatting, linting, and vetting are wired into a top-level `Makefile` (the single source
of truth) and enforced in layers — editor, agent hook, and a lefthook pre-commit hook. See
[ADR-0008](docs/adr/0008-code-quality-tooling.md).

- `make fmt` — format everything (Prettier + gofmt).
- `make lint` — angular-eslint + golangci-lint.
- `make vet` — `go vet`.
- `make test` — web (Vitest) + Go tests.
- `make check` — fmt-check + lint; run before committing.

First-time setup per clone: install the tools (`golangci-lint`, `lefthook`, plus the web
deps via `npm ci`) and run `lefthook install` so the commit hook is active.
