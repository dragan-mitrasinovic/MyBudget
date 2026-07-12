# 0022. Browser-local selectable visual themes

Date: 2026-07-12
Status: Accepted

## Context

A three-direction design study showed that MyBudget can gain personality without replacing
Angular Material, but no single visual identity suits every user. The app already separates
Material tokens from components and independently adapts to System, Light, or Dark, so visual
identity can be another browser-local preference rather than a parallel component system.

## Decision

Offer eight selectable app styles: Soft Ledger (the default), Color-Coded Studio, Bright Utility,
Signal Red, Midnight Terminal, Candy Pop, Velvet Night, and Vaporwave. Each style defines a
distinct typography, Material color and shape system, app surfaces, dashboard treatment, and
category-chart palette, with explicit light and dark values.

Keep the single base `mat.theme()` call from ADR-0009. Centralize the selectable identities in
`web/src/styles/_app-themes.scss` as root `data-app-theme` selectors using scoped
`mat.theme-overrides()`. Components continue to consume Material system variables and semantic app
tokens; they do not branch on theme names.

An Angular root service owns the selected-theme signal, the root attribute, guarded storage
access, and cross-tab storage events. Store the choice under `mybudget.app-theme` in browser
`localStorage`, independently of the `mybudget.theme` System/Light/Dark preference from ADR-0021.
A guarded document-head bootstrap restores both preferences before Angular starts to avoid a
wrong-theme first paint. Canvas charts observe both theme signals when resolving CSS colors so
they redraw immediately.

Use self-hosted Fontsource packages for theme typography. Theme selection remains device-local;
it is not part of the user or budget API.

## Alternatives considered

- Promote one design-study direction for everyone -- simpler, but discards the useful variety the
  study exposed and makes visual taste an application-wide compromise.
- Let users choose arbitrary colors and fonts -- maximally flexible, but creates inaccessible and
  incoherent combinations, complicates support, and weakens the curated Material system.
- Store the choice on the user account -- synchronizes across devices, but adds API and database
  work for a preference commonly expected to be local to a browser or device.
- Couple each visual identity to light or dark mode -- reduces the test matrix, but prevents users
  from combining their preferred identity with the color scheme needed by their environment.

## Consequences

- (+) Users can compare identities spanning calm editorial, playful candy, crisp utility,
  industrial signage, terminal phosphor, luxury editorial, and retro-futurist neon without losing
  shared Material behavior or accessibility primitives.
- (+) A theme applies consistently to the login flow, shell, overlays, dashboard, and Chart.js
  visualizations while System/Light/Dark remains independently selectable.
- (+) New themes have one centralized implementation boundary and cannot require component-level
  theme branches.
- (-) Visual verification covers sixteen primary combinations: eight styles in light and dark,
  across desktop and narrow layouts.
- (-) Additional palettes and self-hosted font assets increase global CSS and maintenance cost.
- (-) The preference is shared by users of the same browser profile and origin and does not follow
  a user to another device.
- (-) The head bootstrap duplicates the valid theme names and default; a future strict Content
  Security Policy will need to authorize or replace the inline script.
