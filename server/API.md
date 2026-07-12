# MyBudget HTTP API

This document is the human-readable contract for the HTTP API implemented by the Go server.
It describes the API as it behaves today. The route registry in
[`internal/httpapi/routes.go`](internal/httpapi/routes.go) and the corresponding handlers remain
the executable source of truth; this file must be updated in the same change whenever that
behavior changes.

## Overview

- API prefix: `/api`
- Production transport: HTTPS
- Local transport: the Angular dev server proxies `/api` to `http://localhost:8080`
- Browser-facing origin: configured by `APP_ORIGIN` (for example, `http://localhost:4200`)
- Request and response field names: camelCase JSON
- Authentication: opaque server-side session in the `session` cookie
- Success bodies: JSON, except `204 No Content` responses
- Error bodies: plain text produced by `http.Error`; clients should primarily use status codes
- Unknown paths return `404 Not Found`; known paths with unsupported methods return `405 Method Not Allowed`

## Endpoint index

| Method        | Path                           | Authentication           | Summary                                                      |
| ------------- | ------------------------------ | ------------------------ | ------------------------------------------------------------ |
| `POST`        | `/api/auth/google`             | Public                   | Exchange a Google ID token for a MyBudget session            |
| `GET`         | `/api/me`                      | Required                 | Return the user represented by the current session           |
| `POST`        | `/api/auth/logout`             | Optional                 | Delete the current session when present and clear its cookie |
| `GET`, `POST` | `/api/dev/login`               | Public, development only | Create a real session for the fixed Dev User                 |
| `GET`         | `/api/budgets`                 | Required                 | List the current user's budgets                              |
| `POST`        | `/api/budgets`                 | Required                 | Create a budget and its permanent Uncategorized category     |
| `GET`         | `/api/budgets/{id}`            | Required                 | Get one owned budget                                         |
| `PATCH`       | `/api/budgets/{id}`            | Required                 | Replace the editable fields of one owned budget              |
| `DELETE`      | `/api/budgets/{id}`            | Required                 | Delete a budget, its categories, and its expenses            |
| `GET`         | `/api/budgets/{id}/dashboard`  | Required                 | Get the reconciled dashboard layout for one owned budget     |
| `PUT`         | `/api/budgets/{id}/dashboard`  | Required                 | Replace the dashboard layout for one owned budget            |
| `GET`         | `/api/budgets/{id}/categories` | Required                 | List categories in one owned budget                          |
| `POST`        | `/api/budgets/{id}/categories` | Required                 | Add a category to one owned budget                           |
| `PATCH`       | `/api/categories/{id}`         | Required                 | Rename one owned category                                    |
| `PUT`         | `/api/categories/{id}/color`   | Required                 | Replace one owned category's chart color                     |
| `DELETE`      | `/api/categories/{id}`         | Required                 | Delete a category after moving its expenses to Uncategorized |
| `GET`         | `/api/budgets/{id}/expenses`   | Required                 | List and optionally filter expenses in one owned budget      |
| `POST`        | `/api/budgets/{id}/expenses`   | Required                 | Add an expense to one owned budget                           |
| `PATCH`       | `/api/expenses/{id}`           | Required                 | Replace the editable fields of one owned expense             |
| `DELETE`      | `/api/expenses/{id}`           | Required                 | Delete one owned expense                                     |

## Authentication and sessions

A successful Google or development login creates an opaque session row in SQLite and sends its
random ID in a cookie named `session`.

Cookie properties:

- `HttpOnly`: browser JavaScript cannot read the credential.
- `SameSite=Lax`: the browser restricts cross-site cookie sending.
- `Path=/`: the cookie applies to every API route.
- `Max-Age`: set to `SESSION_TTL` (720 hours by default) at login and renewal.
- `Secure`: enabled when `IS_DEVELOPMENT=false`; disabled for localhost HTTP when
  `IS_DEVELOPMENT=true`.

A valid session uses throttled sliding expiry. After 24 hours without a renewal, the next
authenticated request atomically moves the SQLite expiry to `now + SESSION_TTL` and reissues the
cookie with a fresh `Max-Age`. For a configured TTL of 24 hours or less, the renewal interval is
half the TTL. With the default configuration, this provides an effective inactivity window of
29–30 days while limiting successful renewal writes to at most one per active session per day.
Requests made before the renewal interval authenticate without writing or reissuing the cookie.

A missing, unknown, or expired session receives `401 Unauthorized` on protected endpoints.
Expired session rows are deleted best-effort when resolved.

Every `POST`, `PUT`, `PATCH`, and `DELETE` route requires an `Origin` header that exactly matches
`APP_ORIGIN`. A missing or different Origin receives `403 Forbidden` with `forbidden`; the
handler is not invoked. Read-only routes do not require Origin. The development-only
`GET /api/dev/login` is the sole state-changing GET exception and remains directly visitable in
a browser; its `POST` variant is Origin-protected like every other POST route.

Every protected query is scoped to the authenticated user. For individual resources, a resource
that does not exist and a resource owned by another user are both generally reported as
`404 Not Found`, avoiding disclosure across users.

## Common representations

### User

```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "Example User",
  "picture": "https://example.com/avatar.png"
}
```

Google's stable subject identifier and the user's creation timestamp are internal and are not
returned.

### Budget

```json
{
  "id": 12,
  "name": "Household",
  "currency": "EUR",
  "archived": false,
  "createdAt": "2026-07-11T08:30:00Z",
  "updatedAt": "2026-07-11T08:30:00Z"
}
```

`createdAt` and `updatedAt` are UTC RFC 3339 timestamps. `userId` is internal and is not
returned.

### Category

```json
{
  "id": 34,
  "budgetId": 12,
  "name": "Food",
  "isSystem": false,
  "colorKey": "chart-1",
  "colorHex": null,
  "createdAt": "2026-07-11T08:31:00Z"
}
```

`isSystem=true` identifies the permanent Uncategorized category. `colorKey` selects one of the
persisted chart palette tokens (`chart-1` through `chart-12`, or `neutral`); a non-null
`colorHex` overrides that token with a normalized six-digit RGB hex color.

### Expense

```json
{
  "id": 56,
  "categoryId": 34,
  "amountMinor": 1599,
  "description": "Burgers",
  "spentAt": "2026-07-11",
  "createdAt": "2026-07-11T08:32:00Z",
  "updatedAt": "2026-07-11T08:32:00Z"
}
```

Expenses do not store or return `budgetId`; their budget is reached through their category.

### Dashboard layout

```json
{
  "version": 6,
  "widgets": [
    {
      "id": "category-expenses:34",
      "type": "category-expenses",
      "categoryId": 34,
      "x": 0,
      "y": 0,
      "cols": 12,
      "rows": 10
    },
    {
      "id": "category-breakdown:11111111-1111-4111-8111-111111111111",
      "type": "category-breakdown",
      "visualization": "pie",
      "x": 12,
      "y": 10,
      "cols": 18,
      "rows": 12
    }
  ]
}
```

Version 6 uses fixed 32-pixel rows and columns on a canvas that begins with a 36-column viewport
and can extend to 20000 columns and 20000 rows. `x` and `y` are zero-based coordinates; `cols` and
`rows` are the widget dimensions. A category expense widget's ID is always
`category-expenses:{categoryId}`. A category breakdown widget has no `categoryId` and uses
`category-breakdown:{uuid}`. Its `visualization` is `pie`, `donut`, `bar`, or `treemap`. The server
accepts versions 1 through 5 and returns them upgraded to version 6. Version 1 used 12 columns and
64-pixel rows; version 2 used 72 columns and 16-pixel rows; versions 3 and 4 used 32-pixel geometry
but limited the canvas to 36 columns; version 5 introduced the current bidirectional canvas.

## Shared data rules

### IDs

Path IDs and `categoryId` values must be positive base-10 integers. An invalid path ID returns
`400 Bad Request` with `invalid id`.

### Money and currency

Money is represented as integer minor units, never floating-point major units. For example,
`1599` represents EUR 15.99 for a two-decimal currency.

Each budget has one currency shared by all its expenses. Currency input is trimmed and converted
to uppercase. The current validation checks exactly three ASCII letters (`A-Z`); it does not
validate against a complete ISO 4217 table.

- `amountMinor` is required and must be greater than zero.

### Names and descriptions

Budget and category names are trimmed and must be 1-100 bytes in the current implementation.
Expense descriptions are trimmed, optional, and limited to 120 Unicode characters. Category
names are unique within a budget using SQLite's current case-sensitive uniqueness, so exact
duplicates conflict while case-only variants are currently distinct.

### Category colors

Category colors are shared by every chart that displays that category. `colorKey` must be one of
`chart-1` through `chart-12` or `neutral`. `colorHex` is either JSON `null` or a six-digit RGB hex
color beginning with `#`; accepted hex is normalized to uppercase. A custom hex overrides the
palette token without discarding the selected key, so resetting to the palette needs no inferred
fallback.

### Dates and timestamps

- `spentAt` is a required calendar date in `YYYY-MM-DD` format.
- `createdAt` and `updatedAt` are server-generated UTC RFC 3339 timestamps.
- Expense `from` and `to` query filters are expected to use `YYYY-MM-DD`, but the current list
  endpoint does not independently validate their format.

### Update semantics

The current `PATCH` endpoints use replacement-style request bodies, not partial merge semantics.
Clients must send every required editable field. An omitted non-pointer field assumes its Go zero
value; for example, omitting `archived` from a budget update sets it to `false`.

## Uncategorized category

Every budget is created with exactly one permanent category named `Uncategorized`. Budget and
Uncategorized creation occur in the same database transaction, so a successfully created budget
always has the fallback category.

Specific behavior:

- Its category response has `isSystem: true`.
- It cannot be renamed or deleted; either attempt returns `403 Forbidden`.
- It can be recolored like any other category and initially uses `colorKey: "neutral"`.
- New expenses may be assigned to it like any other category.
- Deleting a custom category moves all of that category's expenses to the same budget's
  Uncategorized category before deleting the category. The reassignment and deletion occur in
  one transaction, and reassigned expenses receive a new `updatedAt` timestamp.
- Deleting a whole budget cascades to all of its categories, including Uncategorized, and then to
  all expenses in those categories.

## Authentication endpoints

### `POST /api/auth/google`

Exchanges a Google ID token for a MyBudget session. Registration is implicit: the Google user is
created on first login and their email, name, and picture are refreshed on later logins. The
stable Google subject (`sub`) is the account key; email is not used as identity.

Request:

```json
{
  "credential": "<Google ID token>"
}
```

Behavior:

- Validates the token signature, audience, issuer, and expiry through Google's ID-token library.
- Requires the token's `email_verified` claim to be `true`.
- Upserts the user, creates a new session, and sets the `session` cookie.

Success: `200 OK` with a User body.

Errors:

- `400 Bad Request`: malformed JSON or missing/empty `credential`.
- `401 Unauthorized`: invalid Google credential or unverified email.
- `500 Internal Server Error`: user or session persistence failed.

### `GET /api/me`

Returns the User represented by the current valid session cookie. The web client uses this to
restore authentication state after startup or page reload.

Success: `200 OK` with a User body.

Errors:

- `401 Unauthorized`: missing, invalid, or expired session.
- `500 Internal Server Error`: session lookup failed.

### `POST /api/auth/logout`

Deletes the current session row when a session cookie is present, then sends an expired cookie.
The endpoint is idempotent and does not require a valid session; missing or unknown cookies still
receive success. Session deletion is best-effort.

Success: `204 No Content`.

### `GET /api/dev/login` and `POST /api/dev/login`

Available only when `IS_DEVELOPMENT=true`; otherwise the route is not registered and returns
`404 Not Found`. Both methods perform the same state-changing operation. `GET` exists only so a
developer or test agent can open the URL directly in a browser. The `POST` variant requires the
configured app Origin; the development-only `GET` variant does not.

Behavior:

- Upserts one fixed user (`dev-dummy-user`, `dev@mybudget.local`, `Dev User`).
- Creates a real server-side session using the same session mechanism as Google login.
- Sets a non-Secure localhost session cookie in development.
- Bypasses only Google token verification; all later session, ownership, and data behavior is
  real backend behavior.

Success: `200 OK` with a User body.

Errors:

- `500 Internal Server Error`: user or session persistence failed.

## Budget endpoints

All budget endpoints require authentication and operate only on the current user's budgets.

### `GET /api/budgets`

Lists all budgets owned by the current user, ordered newest first by `createdAt`, then by `id`.
Archived budgets are currently included. An empty result is `[]`.

Success: `200 OK` with an array of Budget objects.

Errors: `401 Unauthorized`; `500 Internal Server Error` if the list cannot be loaded.

### `POST /api/budgets`

Creates a budget and its permanent Uncategorized category in one transaction.

Request:

```json
{
  "name": "Household",
  "currency": "EUR"
}
```

Fields:

- `name`: required; trimmed; 1-100 bytes.
- `currency`: required; trimmed, uppercased, and exactly three ASCII letters.

A new budget always has `archived: false`; `archived` is not a create-time field.

Success: `201 Created` with the created Budget.

Errors:

- `400 Bad Request`: malformed JSON or validation failure.
- `401 Unauthorized`: invalid session.
- `500 Internal Server Error`: transactional creation failed.

### `GET /api/budgets/{id}`

Returns one budget when it belongs to the current user.

Success: `200 OK` with a Budget.

Errors:

- `400 Bad Request`: invalid path ID.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing budget or budget owned by another user.
- `500 Internal Server Error`: lookup failed.

### `PATCH /api/budgets/{id}`

Replaces the editable fields of one owned budget. This is currently replacement-style despite
using `PATCH`.

Request:

```json
{
  "name": "Household 2026",
  "currency": "EUR",
  "archived": false
}
```

`name` and `currency` are required; `archived` is a boolean. The server refreshes `updatedAt`.

Success: `200 OK` with the updated Budget.

Errors:

- `400 Bad Request`: invalid path ID, malformed JSON, or validation failure.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing budget or budget owned by another user.
- `500 Internal Server Error`: update failed.

### `DELETE /api/budgets/{id}`

Deletes one owned budget. Foreign-key cascades delete its categories and all expenses belonging
to those categories, as well as its saved dashboard layout.

Success: `204 No Content`.

Errors:

- `400 Bad Request`: invalid path ID.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing budget or budget owned by another user.
- `500 Internal Server Error`: deletion failed.

## Dashboard endpoints

Dashboard layouts belong to one budget and therefore transitively to its owner. Every current
category has exactly one `category-expenses` widget in the response. Users may also add, move,
resize, duplicate, and remove optional `category-breakdown` widgets.

The server reconciles layouts with category changes on both reads and writes. Widgets for
categories that no longer belong to the budget are discarded, while missing current categories
receive the first available default 12-by-10 position. This makes category creation and deletion
safe even when a client is holding an older layout. Reconciliation returned by `GET` is not
itself a database write.

Version 6 validation rules:

- `version` must be `6`. Versions `1` through `5` are accepted and upgraded in the response;
  version-3 through version-5 geometry is unchanged.
- A `category-expenses` widget must have a positive `categoryId`, use the exact derived ID, and be
  the only widget for that category. It must omit `visualization`.
- A `category-breakdown` widget must omit `categoryId` and use a lowercase
  `category-breakdown:{uuid}` ID. It must set `visualization` to `pie`, `donut`, `bar`, or
  `treemap`. Multiple widgets of this type are allowed when their IDs differ.
- Version-4 and version-5 breakdown widgets that omit `visualization` are upgraded to `pie`; a
  supported value already present in a legacy document is preserved.
- Unknown widget types and duplicate widget IDs are rejected.
- `x` and `y` must be non-negative. `cols` must be 9-36, `rows` must be 6-24,
  `x + cols` and `y + rows` must not exceed 20000.
- Retained widgets cannot overlap.

### `GET /api/budgets/{id}/dashboard`

Returns the saved layout reconciled with the budget's current categories. When no layout has
been saved, returns a deterministic default layout with category widgets placed left to right,
then top to bottom.

Success: `200 OK` with a Dashboard layout.

Errors:

- `400 Bad Request`: invalid budget ID.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing budget or budget owned by another user.
- `500 Internal Server Error`: the layout could not be loaded or decoded.

### `PUT /api/budgets/{id}/dashboard`

Validates, reconciles, and replaces the saved layout. The response is authoritative and includes
any category reconciliation or legacy layout upgrade performed by the server.

Request: a complete Dashboard layout object.

Success: `200 OK` with the saved, reconciled Dashboard layout.

Errors:

- `400 Bad Request`: invalid budget ID, malformed JSON, unsupported layout version, invalid
  widget type, identity, or geometry, duplicate widget/category identity, or overlapping widgets.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing budget or budget owned by another user.
- `500 Internal Server Error`: persistence failed.

## Category endpoints

Categories belong to one budget. The owning budget must belong to the current user.

### `GET /api/budgets/{id}/categories`

Lists a budget's custom categories alphabetically using case-insensitive ordering, followed by
the system Uncategorized category. The current implementation relies on the invariant that every
valid budget has Uncategorized: an empty query result is interpreted as `404 Not Found`.

Success: `200 OK` with an array of Category objects.

Errors:

- `400 Bad Request`: invalid budget ID.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing budget or budget owned by another user.
- `500 Internal Server Error`: query failed.

### `POST /api/budgets/{id}/categories`

Adds a custom category to one owned budget.

Request:

```json
{
  "name": "Food"
}
```

`name` is required, trimmed, and limited to 1-100 bytes.

Success: `201 Created` with the created Category (`isSystem: false`). New custom categories receive
the first palette key not already used by a custom category in that budget, cycling through the
12-key palette when all keys are in use.

Errors:

- `400 Bad Request`: invalid budget ID, malformed JSON, or invalid name.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing budget or budget owned by another user.
- `409 Conflict`: an exact category name already exists in the budget.
- `500 Internal Server Error`: creation failed.

### `PATCH /api/categories/{id}`

Renames one custom category owned through the current user's budget.

Request:

```json
{
  "name": "Groceries"
}
```

Success: `200 OK` with the updated Category.

Errors:

- `400 Bad Request`: invalid category ID, malformed JSON, or invalid name.
- `401 Unauthorized`: invalid session.
- `403 Forbidden`: attempted to rename Uncategorized.
- `404 Not Found`: missing category or category owned by another user.
- `409 Conflict`: an exact category name already exists in the budget.
- `500 Internal Server Error`: update failed.

### `PUT /api/categories/{id}/color`

Replaces one category's shared chart color. This operation is also allowed for Uncategorized;
its system protection applies only to renaming and deletion.

Request using the curated palette:

```json
{
  "colorKey": "chart-5",
  "colorHex": null
}
```

Request using a custom override:

```json
{
  "colorKey": "chart-5",
  "colorHex": "#A1B2C3"
}
```

Success: `200 OK` with the updated Category.

Errors:

- `400 Bad Request`: invalid category ID, malformed JSON, unknown `colorKey`, or invalid
  `colorHex`.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing category or category owned by another user.
- `500 Internal Server Error`: update failed.

### `DELETE /api/categories/{id}`

Deletes one custom category. Before deletion, all of its expenses are reassigned to the same
budget's Uncategorized category in the same transaction. No expenses are deleted.

Success: `204 No Content`.

Errors:

- `400 Bad Request`: invalid category ID.
- `401 Unauthorized`: invalid session.
- `403 Forbidden`: attempted to delete Uncategorized.
- `404 Not Found`: missing category or category owned by another user.
- `500 Internal Server Error`: reassignment or deletion failed.

## Expense endpoints

Expenses belong to categories, and categories belong to budgets. An expense has no independent
`budgetId`; ownership and budget membership are derived through its category.

### `GET /api/budgets/{id}/expenses`

Lists expenses in one owned budget, ordered by `spentAt` descending and then by `id` descending.
An empty result is `[]`.

Optional query parameters:

| Parameter    | Meaning                                            |
| ------------ | -------------------------------------------------- |
| `from`       | Include expenses with `spentAt >= from`            |
| `to`         | Include expenses with `spentAt <= to`              |
| `categoryId` | Include only expenses in this positive category ID |

`from` and `to` are expected to be `YYYY-MM-DD` strings; they are currently compared as stored
ISO date strings without separate request validation. If `categoryId` names a category outside
the budget, the result is an empty array rather than an error.

Success: `200 OK` with an array of Expense objects.

Errors:

- `400 Bad Request`: invalid budget ID or non-positive/non-integer `categoryId`.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing budget or budget owned by another user.
- `500 Internal Server Error`: query failed.

### `POST /api/budgets/{id}/expenses`

Adds an expense to one owned budget. The selected category must belong to the same budget.

Request:

```json
{
  "categoryId": 34,
  "amountMinor": 1599,
  "description": "Burgers",
  "spentAt": "2026-07-11"
}
```

Fields:

- `categoryId`: required positive integer; must belong to the path budget.
- `amountMinor`: required integer greater than zero.
- `description`: optional string; trimmed; at most 120 Unicode characters; omitted becomes `""`.
- `spentAt`: required valid date in `YYYY-MM-DD` format.

Success: `201 Created` with the created Expense.

Errors:

- `400 Bad Request`: invalid budget ID, malformed JSON, validation failure, or category not in
  the budget.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing budget or budget owned by another user.
- `500 Internal Server Error`: creation failed.

### `PATCH /api/expenses/{id}`

Replaces the editable fields of one owned expense. This is currently replacement-style despite
using `PATCH`. The new category must belong to the expense's existing budget, so this endpoint
cannot move an expense between budgets.

Request:

```json
{
  "categoryId": 35,
  "amountMinor": 1899,
  "description": "Dinner",
  "spentAt": "2026-07-11"
}
```

`categoryId`, `amountMinor`, and `spentAt` are required; `description` may be empty. The server
refreshes `updatedAt`.
`updatedAt`.

Success: `200 OK` with the updated Expense.

Errors:

- `400 Bad Request`: invalid expense ID, malformed JSON, validation failure, or category not in
  the expense's budget.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing expense or expense owned by another user.
- `500 Internal Server Error`: update failed.

### `DELETE /api/expenses/{id}`

Deletes one expense when it belongs to the current user through its category and budget.

Success: `204 No Content`.

Errors:

- `400 Bad Request`: invalid expense ID.
- `401 Unauthorized`: invalid session.
- `404 Not Found`: missing expense or expense owned by another user.
- `500 Internal Server Error`: deletion failed.
