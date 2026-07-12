# 0019. Bidirectionally extensible dashboard canvas

Date: 2026-07-12
Status: Accepted

Supersedes ADR-0017.

## Context

ADR-0017 established a persisted Gridster dashboard with a 36-column desktop boundary and an
effectively unbounded vertical axis. That makes additional rows possible but prevents users from
building a wider workspace or arranging related widgets side by side beyond the initial viewport.
The dashboard should extend in either direction while retaining predictable cell dimensions and
the existing saved arrangements.

## Decision

Keep angular-gridster2 and the existing persisted widget model, but use its fixed-grid mode with
32-by-32-pixel cells and scrolling on both axes. The initial viewport remains 36 columns wide so
existing layouts keep their familiar scale. Users may drag widgets right or down and the canvas
expands with their content, up to the defensive server limit of 20000 columns by 20000 rows.
Widgets remain limited to 9-36 columns and 6-24 rows; this changes workspace reach, not individual
widget size.

Raise the dashboard layout contract to version 5. Accept versions 1 through 4 and preserve their
geometry during upgrade after applying their historical validation rules. Continue placing new
automatic category widgets and user-added charts within the initial 36-column region; horizontal
expansion happens only when the user deliberately moves a widget there. Desktop uses the same
persisted canvas across viewport sizes, while narrow screens retain Gridster's derived stacked
presentation.

The layout remains one validated JSON document per budget. Category widgets are still reconciled
against current categories, optional chart widgets remain independently persisted, and drag or
resize completion continues to use the serialized debounced save queue established previously.

## Alternatives considered

- **Keep 36 columns and add horizontal pages** — creates separate navigation and persistence
  semantics for what users expect to be one spatial canvas.
- **Use horizontal-fixed mode** — extends right but constrains vertical behavior, reversing rather
  than removing the one-axis limitation.
- **Use responsive cell widths on an unbounded grid** — makes coordinates change physical scale as
  the viewport changes and weakens the 32-pixel placement model.
- **Remove all coordinate limits** — integers still need defensive bounds for validation, Gridster
  configuration, and pathological payload protection; 20000 cells per axis is effectively
  unbounded for normal use.

## Consequences

- (+) Users can build dashboards wider or taller than the viewport and scroll to every widget.
- (+) Existing layouts retain their positions and approximately 32-pixel cell dimensions.
- (+) Server and client share the same explicit two-axis bounds.
- (-) Wide dashboards require horizontal scrolling and can place widgets far from the origin.
- (-) The fixed-width initial grid may horizontally scroll on narrower desktop containers; mobile
  continues to use its stacked presentation.
- (-) Layout version 5 adds another legacy validation and upgrade path.
