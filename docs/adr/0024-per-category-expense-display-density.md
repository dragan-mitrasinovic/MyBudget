# 0024. Per-category expense display density

Date: 2026-07-17
Status: Superseded by ADR-0025

## Context

Category widgets can contain very different numbers and kinds of expenses. Always showing a title,
amount, and date spends vertical space that some categories do not need, while automatically hiding
details based on widget size would make the content unpredictable and remove user control.

## Decision

Let users choose one expense-row presentation for each category widget: `amount-only`,
`title-amount`, or `title-amount-date`. `title-amount` is the default for new categories and existing
layouts. The category actions menu exposes the three choices in an Expense display submenu, and the
expense actions button remains available in every presentation.

Persist the choice as a required `expenseDisplay` field on each `category-expenses` dashboard
widget and raise the dashboard layout contract to version 7. Versions 1 through 6 upgrade missing
values to `title-amount`; current layouts reject missing or unknown values and reject the field on
category breakdown widgets. The browser sends changes through the existing debounced dashboard
save queue.

Keep compact modes as genuine one-line Angular Material list items. The full mode adds the date as
a second list line, so the selected detail level directly controls row height without fixed CSS
height overrides.

## Alternatives considered

- **Store the preference on the category database row** — makes a dashboard presentation choice
  part of the category domain contract and requires a separate category mutation path.
- **Store the preference in browser-local state** — avoids a layout version change, but the choice
  would not follow the signed-in user's budget across browsers.
- **Infer detail from widget height** — adapts automatically, but resizing would unexpectedly hide
  or reveal information and users could not choose a stable presentation.
- **Always show all fields with tighter CSS** — preserves information but does not provide the
  meaningful space reduction requested for high-volume categories.

## Consequences

- (+) Each category can trade detail for visible item count independently.
- (+) The default is denser than the previous title, amount, and date presentation.
- (+) Display choices persist with the dashboard and follow the budget across browsers.
- (+) Existing layouts upgrade deterministically without changing widget geometry.
- (-) Dashboard version 7 adds another validated widget field and legacy defaulting rule.
- (-) Amount-only rows require opening the expense editor to identify an item beyond its amount.
