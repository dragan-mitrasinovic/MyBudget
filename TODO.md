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
  - [x] Phase 3 — Expenses CRUD (`spent_at` date defaulting to today, quick "save & add another") (2026-07-09) - [x] Added concise optional expense descriptions as the primary UI label (2026-07-11)
  - [x] Phase 4 — Persisted customizable desktop dashboard with movable, resizable
        category-expense widgets (2026-07-12)
  - [x] Phase 5 — Optional all-time category-breakdown pie charts with shared category colours,
        curated palette + custom-hex override, and persisted dashboard placement (2026-07-12)
  - [x] Phase 6 — Bidirectionally extensible dashboard canvas and livelier chart palette
        (2026-07-12)
  - [x] Phase 7 — Configurable Pie, Donut, horizontal Bar, and Treemap presentations for the
        all-time category breakdown (2026-07-12)
  - [x] Phase 8 — One-click budget shortcuts in the sidenav with overflow and active-route
        highlighting, plus richer budget overview cards (2026-07-12)
  - [ ] Later session — spend-over-time charts; add server-side summary aggregation when expense
        pagination or broader analytics makes client-side aggregation incomplete

## Done

- [x] Adaptive System / Light / Dark theme with a browser-local preference (2026-07-12)
- [x] Defined mutable development-stage migration history until prerelease (2026-07-11)
- [x] Simplified budgets to tracking-only containers across UI, API, and database (2026-07-11)
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
