# MyBudget — Tasks

Quick-capture backlog. Drop new ideas under **Inbox** the moment they strike — no
triage required. Groom them into the sections below when you have a moment, and check
items off as you finish them. Big architectural or tech-stack decisions belong in an
ADR (`docs/adr/`), not here — this file is "what to do," ADRs are "why we chose how."

## Inbox

_Unsorted, capture-anything. Groom later._

- [ ] Define and implement currency safety: validate supported currencies/minor units and
      prevent or explicitly handle changing a budget's currency after expenses exist

## In progress

- [ ] Budgets feature — users create budgets and track expenses in custom categories, with a
      per-budget currency and a permanent Uncategorized category. See
      [ADR-0011](docs/adr/0011-budgets-categories-expenses-data-model.md).
  - [x] Phase 1 — Budgets CRUD (auto-creates Uncategorized) + app-shell refactor (2026-07-05)
  - [x] Phase 2 — Categories CRUD (deleting one reassigns its expenses to Uncategorized) (2026-07-09)
  - [x] Phase 3 — Expenses CRUD (`spent_at` date defaulting to today, quick "save & add another") (2026-07-09)
  - [ ] Later session — charts + category colours (option C): category breakdown & spend
        over time, `GET /api/budgets/{id}/summary`, chart-library ADR, token palette +
        custom-hex override

## Done

- [x] Basic backend test suite covering request validation, data ownership and integrity, and
      route security (2026-07-11)
- [x] Automatic Go API rebuild and restart with Air (2026-07-11)
- [x] Throttled sliding session expiry with daily renewal (2026-07-11)
- [x] Home page app shell: top bar (app name + log out) and left sidenav for future menus (2026-07-04)
- [x] Dev dummy-user login (`IS_DEVELOPMENT=true`) + in-browser verification workflow for agents (2026-07-02)
- [x] Design system foundation: tokens, Material-first pages, stylelint design lint (2026-07-02)
- [x] Code-quality tooling: Prettier/ESLint, golangci-lint, lefthook, Makefile (2026-07-01)
- [x] Google sign-in: Go HTTP API, Angular auth, monorepo split (2026-07-01)
- [x] Architecture decision records and agent guide (2026-06-29)
- [x] Angular 22 + Angular Material scaffold (2026-06-28)

## Someday / maybe

_Parked ideas, not committed to._

- [ ] Storybook component catalog + visual-regression testing (Chromatic / Playwright), once
      there are enough reusable components to justify it
