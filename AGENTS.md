# MyBudget

MyBudget is a personal expense-tracking and budgeting app for the author and a few
friends, so it is multi-user from the start. The web client is an Angular + Angular
Material single-page app; a Go HTTP API (`server/`) and a future mobile client are
planned. See `docs/adr/` for the decisions behind these choices.

## Tasks

The task backlog lives in [`TODO.md`](TODO.md) at the repo root — a single Markdown file
with **Inbox**, **In progress**, **Done**, and **Someday / maybe** sections. Capture new
ideas under Inbox, move them along as work happens, and check items off (with a date under
Done) when finished. Keep it current as part of doing the work.

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

## Running & verifying locally

Two processes, then open http://localhost:4200:

- `make dev-api` — the Go API on :8080 in development mode (`IS_DEVELOPMENT=true`),
  with the dev-only dummy-login route and non-Secure localhost session cookies.
- `make dev-web` — the Angular dev server, which proxies `/api` to the API.

**Signing in without Google (dummy user).** The app is behind Google sign-in (ADR-0006), which
can't be completed headlessly. With `make dev-api` running, mint a session for a fixed **Dev
User** by visiting `http://localhost:4200/api/dev/login` once, then open
`http://localhost:4200/`. This shortcut exists only when `IS_DEVELOPMENT=true` (which `make
dev-api` sets). With `IS_DEVELOPMENT=false`, the route is absent and session cookies are
Secure — see [ADR-0012](docs/adr/0012-single-application-environment-mode.md).

**Check UI work in the browser.** A green build and green lint don't prove a screen _looks_
right. After a UI change, open the affected page, read its accessibility snapshot
(structure/interaction) and take a screenshot (layout/spacing/theme), and fix what you see —
overlap, overflow, off-theme colours, broken images. This catches _broken_; taste and
aesthetics still need a human eye.
