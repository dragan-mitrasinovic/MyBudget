# 0023. Remove the default Uncategorized category

Date: 2026-07-16
Status: Accepted

Supersedes ADR-0011 and ADR-0018.

## Context

ADR-0011 made a protected Uncategorized row the category fallback for every budget, and ADR-0018
made that row a visible dashboard and chart category. Once the dashboard began showing all expense
lists, the automatic section became visual noise rather than useful structure. The application is
still in active development with no deployed users, so ADR-0016 permits correcting the baseline
schema and recreating local databases instead of preserving this special case indefinitely.

## Decision

Keep budgets, categories, and expenses as normalized user-owned data. Categories remain scoped to
one budget, every expense continues to require a non-null `category_id`, and expense ownership is
still reached through `expense -> category -> budget`; expenses do not gain a redundant
`budget_id`.

Remove the persisted Uncategorized concept completely:

- New budgets begin with zero categories and do not seed a fallback row.
- Every category is user-created. Remove `categories.is_system` and the API `isSystem` field, and
  reserve no category name.
- A budget with zero categories is valid. Its category list is `[]`, and its unsaved dashboard is
  layout version 6 with no widgets. Expense creation remains unavailable until a category exists.
- The dashboard continues to reconcile exactly one `category-expenses:{categoryId}` widget for
  every current category. Optional category-breakdown widgets remain independently persisted, so
  no dashboard layout version change is needed.
- Deleting an empty category requires no replacement. Deleting a populated category requires the
  user to choose a different category in the same budget; the server moves every expense, refreshes
  its `updated_at`, and deletes the source category in one transaction. A populated category is
  never removed without a valid destination, and category deletion never deletes expenses.

Retain the category visualization decisions from ADR-0018. Category-breakdown widgets use Chart.js
through ng2-charts, derive all-time positive-spend totals in Angular from the complete loaded
category and expense signals, and provide an HTML legend. Categories keep shared `color_key` and
nullable `color_hex` values. Automatic colors use the 12 chart palette keys, while `neutral`
remains an ordinary manually selectable palette value rather than a system-category default.

Fold the destructive schema correction into the existing baseline migrations as allowed by
ADR-0016. Local development databases must be recreated; no compatibility migration converts old
Uncategorized rows or saved layouts.

## Alternatives considered

- **Keep the row but hide its dashboard widget** — leaves the special schema, API flag, mutation
  guards, fallback behavior, and chart identity in place while only masking one symptom.
- **Allow category-less expenses** — requires nullable `category_id`, a new direct budget ownership
  path, synthetic UI grouping, and broader query changes even though explicit categorization is
  still desired.
- **Automatically choose the first remaining category on deletion** — preserves data but silently
  changes its classification without user intent.
- **Block every populated-category deletion** — safe but needlessly forces users to edit expenses
  one at a time when an atomic move has clear semantics.
- **Cascade-delete a category's expenses** — makes a classification action destructive and risks
  losing financial records.
- **Preserve local databases with a forward conversion** — adds compatibility machinery for data
  that ADR-0016 explicitly treats as disposable during this stage.

## Consequences

- (+) The schema and API contain only user-meaningful categories and no protected row or transport
  flag.
- (+) The dashboard no longer creates a default expense-list section, while custom category widgets
  and charts keep their existing layout contract.
- (+) Category deletion preserves every expense and makes its destination explicit.
- (+) Empty budgets have unambiguous API behavior because ownership is checked independently of
  category existence.
- (-) A new budget requires one category-creation step before its first expense.
- (-) Category deletion needs a target-selection dialog and stale-state handling when expenses or
  categories change concurrently.
- (-) Local development databases and their saved dashboard layouts must be recreated after the
  baseline migration rewrite.
