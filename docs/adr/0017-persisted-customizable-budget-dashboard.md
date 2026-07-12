# 0017. Persisted customizable budget dashboard

Date: 2026-07-12
Status: Superseded by ADR-0019

## Context

Budget details need to evolve from fixed category and expense lists into a user-arranged
dashboard. Every category must currently have an expense-list section, users must be able to move
and resize those sections on desktop, and the arrangement must follow the signed-in user across
browsers. The design also needs to admit optional charts and other widget types later without
replacing the layout contract.

## Decision

Use `angular-gridster2` 22.x for the Angular dashboard grid. It matches Angular 22 and provides
standalone components, collision handling, dragging, resizing, grid pushing, and responsive
stacking. The budget detail page uses a 36-column desktop grid with 32-pixel rows, an explicit
drag handle, and resize handles along all four edges. Resizing into an adjacent widget shrinks
that widget down to its configured minimum instead of moving it; once the neighbor reaches its
minimum, resizing stops. At the standard desktop content width this produces approximately
32-by-32-pixel grid cells: granular enough for deliberate placement without making the movement
preview too busy. Expense content remains interactive and scrolls inside each widget.

Persist one versioned dashboard layout per budget in SQLite and expose it through
`GET /api/budgets/{id}/dashboard` and `PUT /api/budgets/{id}/dashboard`. Budgets currently have a
single owner, so `budget_id` is both the layout key and the ownership boundary; deleting a budget
cascades to its layout. The versioned JSON document stores typed widget identity plus `x`, `y`,
`cols`, and `rows` geometry. The server validates the document rather than accepting arbitrary
Gridster state.

The current contract is layout version 3. Version 1 used 12 columns and 64-pixel rows; the
short-lived version 2 used 72 columns and 16-pixel rows. The server converts both formats into
version 3 while preserving widget boundaries as closely as the 32-pixel grid allows. Default
widgets remain one-third of the dashboard width and 320 pixels high.

For now, every category maps to one permanent `category-expenses:{categoryId}` widget. On every
read and write, the server removes widgets for deleted categories and places missing category
widgets into available default positions. The browser saves after drag or resize completion with
a short debounce and serializes overlapping save attempts. The dashboard does not expose a reset
command; users directly arrange the persistent canvas.

Desktop editing is the current product scope. Narrow screens use Gridster's derived stacked
presentation rather than persisting a separate mobile arrangement. Category management will
move to a separate budget-details surface later and is not part of the dashboard grid.

## Alternatives considered

- **Angular CDK drag-and-drop with custom resizing and packing** — avoids another dependency but
  requires us to implement collision detection, resize behavior, pushing, and grid math.
- **GridStack** — mature and framework-neutral, but its Angular integration is less direct than
  Gridster's Angular 22 standalone components for this application.
- **Browser-local persistence** — simpler, but a signed-in user's layout would differ across
  browsers and devices and could disappear with local storage.
- **Normalized widget rows and coordinate columns** — easier to query, but adds schema and
  migration cost for data that is always loaded and replaced as one small layout document. The
  versioned JSON contract remains server-validated and can be migrated when its shape changes.
- **12-column, 64-pixel grid** — too coarse for deliberate movement and resizing.
- **72-column, 16-pixel grid** — precise, but the movement preview and snapping felt too fine.

## Consequences

- (+) Users can arrange and resize category sections, and the layout survives reloads and follows
  the owned budget across browsers.
- (+) Future widget types can extend the versioned layout without coupling all geometry to the
  categories table.
- (+) Server reconciliation keeps category creation and deletion from corrupting an existing
  layout.
- (+) Existing saved layouts retain approximately the same geometry through deterministic
  version conversion.
- (-) `angular-gridster2` becomes a user-interface dependency that must stay compatible with
  Angular upgrades.
- (-) SQLite cannot enforce the full JSON shape; Go validation and focused contract tests are
  required.
- (-) The versioned layout contract requires upgrade logic when grid geometry changes.
- (-) Pointer-based editing is not a complete accessibility solution. Keyboard move and resize
  controls need a later interaction design.
- (-) A separate mobile layout is not remembered; adding one later will require an explicit
  layout-scope or breakpoint contract.
