# 0018. Dashboard charts and category colors

Date: 2026-07-12
Status: Superseded by ADR-0023

## Context

The persisted budget dashboard needs optional visual summaries in addition to its mandatory
category expense lists. The first summary is an all-time pie chart of spending by category, and a
category should retain the same recognizable color across every chart while still allowing users
to choose a custom color. The design must extend the existing versioned layout without coupling
calculated expense totals to saved widget state.

## Decision

Use Chart.js 4 through ng2-charts 10 for Angular chart rendering. Register only the pie controller,
arc element, and tooltip from the chart widget's provider so Chart.js remains in the lazy budget
detail bundle. Render an HTML legend beside the canvas with category names, formatted amounts,
percentages, and operable color swatches; the canvas is not the sole accessible representation.

Add `category-breakdown` as an optional dashboard widget and raise the layout contract to version 4. Category breakdown IDs are `category-breakdown:{uuid}` and contain only identity and geometry;
users may add duplicate instances with distinct IDs. Versions 1 through 3 remain accepted and are
upgraded to version 4. Mandatory `category-expenses` widgets continue to be reconciled with current
categories, while optional charts are preserved independently.

The first chart covers all expenses in its budget across all dates, includes categories with
positive spending, and omits zero-spend categories. Calculate its values in Angular from the
complete category and expense signals already required by the category-list widgets. Do not save
calculated totals in the layout and do not add a summary endpoint until expense pagination or more
complex analytics makes the client list incomplete or inefficient.

Store category-global colors as `color_key` plus nullable `color_hex` columns. `color_key` selects
one of 12 stable curated palette tokens or the neutral token; `color_hex` is an optional custom
override. New categories receive the first unused palette key in their budget and cycle after the
palette is exhausted. Uncategorized starts neutral and may be recolored even though it remains
protected from rename and deletion. A dedicated color endpoint updates this shared identity, so all
chart instances redraw consistently.

## Alternatives considered

- **Apache ECharts with ngx-echarts** — provides a broader visualization toolbox, but the first
  pie-chart scope does not justify its larger API and package surface.
- **Custom SVG or canvas rendering** — avoids a dependency but makes responsive drawing, tooltips,
  resize lifecycle, and future chart types application-owned concerns.
- **Server-side `/summary` aggregation immediately** — avoids client aggregation but duplicates
  data already loaded in full and requires coordinated refreshes after every expense mutation.
- **Per-chart colors** — allows independent styling but makes the same category visually unstable
  across duplicate and future chart types.
- **Curated colors only** — guarantees theme-aligned choices but does not meet the requirement for
  full user customization; arbitrary custom hex remains an explicit advanced override.

## Consequences

- (+) Charts are optional, movable, resizable, duplicable, and persisted without storing stale
  calculated data.
- (+) A category has one recognizable color across all current and future visualizations.
- (+) Selective lazy registration keeps the application's initial bundle below its warning budget.
- (+) The HTML legend makes values and color editing available without relying on canvas pixels.
- (-) Custom colors can clash or have poor contrast; chart text and focus indicators therefore use
  Material tokens rather than being drawn over user colors.
- (-) All-time client aggregation assumes the dashboard holds the complete expense list. Adding
  pagination will require a server-side aggregate contract before charts can remain accurate.
- (-) Layout version 4 adds another legacy conversion path that server validation and tests must
  preserve.
