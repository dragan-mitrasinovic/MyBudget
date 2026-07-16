# 0011. Budgets, categories, and expenses data model

Date: 2026-07-05
Status: Superseded by ADR-0023

## Context

The first core feature is budgeting: a user creates budgets and records expenses in them,
grouped by their own custom categories. The app is multi-user from the start, so every row
must be owned by a user, and access must be scoped accordingly. We need a data model that is
simple to start with but leaves room for the visual/reporting features (charts) that come
later. This ADR records the shape of that model and the choices behind it; charts and
category colours are deliberately out of scope here (a later ADR will cover the charting
library and colour palette).

## Decision

Three new tables, owned transitively by a user via the budget:

- **budgets** — `id`, `user_id` (FK → `users`, `ON DELETE CASCADE`), `name`, `currency`,
  `archived`, `created_at`, `updated_at`.
- **categories** — `id`, `budget_id` (FK → `budgets`, `ON DELETE CASCADE`), `name`
  (unique per budget), `is_system` (`0`/`1`), `created_at`.
- **expenses** — `id`, `category_id` (FK → `categories`, `ON DELETE CASCADE`),
  `amount_minor`, `description`, `spent_at`, `created_at`, `updated_at`.

Supporting rules:

- **Money is stored as integer minor units** (`amount_minor`) to avoid floating-point rounding
  error. Currency is a per-budget ISO-4217 code; all expenses in a budget share it (no
  per-expense currency or conversion yet).
- **Categories are scoped to a single budget** (not a shared per-user library).
- **Every budget has a permanent "Uncategorized" category**, created together with the
  budget in one transaction and flagged with `is_system = 1`. It cannot be renamed or
  deleted. Every expense therefore always has a category (`category_id` is `NOT NULL`).
- **Deleting a category** reassigns its expenses to that budget's Uncategorized category
  (in a transaction), then deletes the category — expenses are never lost. **Deleting a
  budget** cascades to its categories and, through them, to their expenses.
- **Expenses carry their own `spent_at` date**, distinct from `created_at` (when the row was
  entered). The UI defaults `spent_at` to today but lets the user change it, so the spend
  date is accurate even when logged later. Time-based filters/queries key off `spent_at`.
- **Expenses may have an optional `description`** as their primary human-readable label, such
  as "Burgers". Descriptions are trimmed and limited to 120 Unicode characters. The category
  remains the expense's classification and is the primary UI fallback when the description is
  empty.
- **Expenses do not store `budget_id`**; the owning budget is reached through the category.
  Booleans (`archived`, `is_system`) are `0`/`1` integers, since SQLite has no boolean type.

## Alternatives considered

- **First-class recurring periods** (each month a tracked budget instance with its own
  limits/rollover) — more powerful, but a lot of machinery. Budgets are kept as neutral
  containers; the user decides what a budget represents and slices by date as needed.
- **Shared per-user category library** reused across budgets — deferred; per-budget
  categories are simpler, and "copy categories from another budget"/templates can be added
  later without changing the model.
- **Nullable `category_id` for uncategorized spend** — rejected in favour of an explicit
  permanent Uncategorized category, so "uncategorized" is a real, chartable bucket and the
  add flow always has a valid default.
- **Denormalised `budget_id` on expenses** — rejected as redundant with `category → budget`
  and able to drift out of sync; the join is cheap at our scale.
- **Floating-point money** — rejected; classic source of rounding bugs.
- **Multi-currency expenses with conversion** — deferred; single currency per budget covers
  the current use cases. Revisit if needed.

## Consequences

- (+) Simple, normalised schema; ownership is easy to enforce by scoping through the budget,
  and money is exact.
- (+) Every expense is always categorised, and no deletion path silently loses expenses.
- (+) Accurate spend dates via `spent_at`, independent of when a row was entered.
- (−) Listing a budget's expenses requires a join through `categories` (no `budget_id`
  shortcut).
- (−) The Uncategorized row is special and needs guard logic (no rename/delete, auto-created
  with every budget).
- (−) Categories don't yet carry colours; per-budget scope means no reuse across budgets. These
  are intentional omissions for now (colours arrive with charts).
