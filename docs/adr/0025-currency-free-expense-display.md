# 0025. Currency-free expense display

Date: 2026-07-17
Status: Accepted

## Context

ADR-0024 introduced three persisted expense-row presentations, including `amount-only` for a
compact row that still uses localized currency formatting. Some categories need an even quieter
presentation where users already know the budget currency and want each row to contain only the
numeric value.

## Decision

Offer four expense-row presentations for each category widget: `number-only`, `amount-only`,
`title-amount`, and `title-amount-date`. `number-only` formats the minor-unit value as a localized
two-decimal number without a currency symbol or code. `amount-only` remains the currency-formatted
compact choice, and `title-amount` remains the default for new categories and layouts predating
expense display settings.

Persist `number-only` as another valid `expenseDisplay` value and raise the dashboard layout
contract to version 8. Version-7 layouts preserve their existing choice when upgraded; versions 1
through 6 still receive `title-amount` when the field is absent. Current layouts reject missing or
unknown values and category breakdown widgets continue to reject the field.

All compact modes remain genuine one-line Angular Material list items, and every mode retains the
expense actions button. The category total continues to show currency because this preference only
controls individual expense rows.

## Alternatives considered

- **Change `amount-only` to omit currency** — silently changes an existing persisted choice and
  removes the useful compact currency-formatted presentation.
- **Add a category-level hide-currency toggle** — creates combinations that complicate the menu
  and persistence model for one requested presentation.
- **Render raw integer minor units** — avoids decimal formatting but would expose storage units
  such as `1599` instead of the user-facing value `15.99`.

## Consequences

- (+) Users can choose the most compact, visually quiet expense presentation.
- (+) Existing `amount-only` layouts retain their currency formatting.
- (+) The category total still communicates which currency the numbers use.
- (-) Dashboard version 8 extends the validated display enum and requires another compatibility
  path.
- (-) A number-only row is ambiguous when viewed without the surrounding budget context.
