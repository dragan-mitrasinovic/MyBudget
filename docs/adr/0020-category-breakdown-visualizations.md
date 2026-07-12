# 0020. Configurable category breakdown visualizations

Date: 2026-07-12
Status: Accepted

## Context

The first category breakdown widget introduced by ADR-0018 renders an all-time budget breakdown as
a pie chart. Users need alternative visual forms of the same totals without introducing date
ranges, new measures, category filters, or separate analytics contracts. Switching presentation
should not discard a widget's identity, position, size, or shared category colors.

## Decision

Keep one `category-breakdown` widget type and persist a required `visualization` value: `pie`,
`donut`, `bar`, or `treemap`. Raise the dashboard layout contract to version 6. Version-4 and
version-5 breakdown widgets without the field upgrade to `pie`; supported values already present
in a legacy document are preserved. Current layouts reject missing or unknown visualization values
and reject visualization fields on category expense widgets.

The Add chart menu offers all four presentations directly. A chart-type submenu on each widget
also switches its visualization in place through the existing debounced layout save queue. Every
presentation consumes the same live, all-time positive category totals sorted by amount, uses the
same category-global colors, and retains the shared HTML legend with amounts, percentages, and
color controls.

Use Chart.js's built-in Pie, Doughnut, and Bar controllers. Horizontal bars use a zero-based linear
value axis and a scrollable plot height when category count grows. Add `chartjs-chart-treemap` 4.x
for Treemap rendering and selectively register its controller and element in the lazy budget
detail widget. Treemap rectangle area represents `amountMinor`; arbitrary user colors fill the
rectangles, while category names and values remain in tooltips and the shared HTML legend rather
than being drawn over colors with unknown contrast.

Render each Chart.js type in its own Angular template branch so changing between incompatible
dataset shapes destroys and recreates the canvas directive cleanly. The HTML legend remains the
authoritative accessible representation for all chart types, and each canvas exposes a textual
summary naming its current presentation.

## Alternatives considered

- **Separate widget types per visualization** — duplicates persistence and component behavior and
  makes changing presentation require deleting and recreating a widget.
- **A single dynamically mutated Chart.js instance** — reduces template branches but risks stale
  controller state when moving between arc, Cartesian, and Treemap dataset shapes.
- **Custom Treemap rendering** — avoids a small dependency but makes spatial layout, hit testing,
  tooltips, resizing, and future maintenance application-owned concerns.
- **Pie and bars only** — avoids the Treemap plugin but does not provide the approved spatial
  part-to-whole presentation.

## Consequences

- (+) Users can create or switch among four views without changing the underlying analysis.
- (+) Widget identity and geometry remain stable when presentation changes.
- (+) Shared colors, legends, accessibility, and live updates stay consistent across all types.
- (+) Selective registration keeps the chart implementation in the lazy budget-detail bundle.
- (-) The Treemap plugin becomes an additional visualization dependency that must remain compatible
  with Chart.js.
- (-) Four Chart.js dataset/option shapes require separate focused tests and browser pixel checks.
- (-) Layout version 6 adds legacy defaulting and another strict validation path.
