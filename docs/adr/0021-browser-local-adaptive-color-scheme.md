# 0021. Browser-local adaptive color scheme

Date: 2026-07-12
Status: Accepted

## Context

The web client has a token-based Angular Material 3 theme but previously forced its light
scheme. Users need a dark appearance that can follow their operating-system preference or be
overridden without duplicating component styles. The preference is visual and device-specific,
while some Chart.js colors require an Angular signal because canvas does not react to CSS token
changes after drawing.

## Decision

Offer System, Light, and Dark choices from the application toolbar. System is the default and is
represented by the absence of a saved value; explicit Light or Dark choices are stored under
`mybudget.theme` in browser `localStorage`.

Keep the single global `mat.theme()` definition and switch its generated light/dark system-token
values with the root CSS `color-scheme` property. A root `data-theme` attribute forces explicit
choices, while no attribute lets `prefers-color-scheme` select the scheme. A small guarded script
in the document head restores an explicit choice before first paint.

An Angular root service owns the preference signal, effective light/dark signal, CDK
`MediaMatcher` observation, root attribute, storage access, and cross-tab storage events. Canvas
charts depend on the effective signal when resolving theme tokens so they redraw after either an
explicit change or an operating-system change.

## Alternatives considered

- Follow the operating system only -- simplest, but gives users no override when their preferred
  app appearance differs from the device setting.
- Use a Light/Dark toggle -- fewer controls, but an explicit choice leaves no way to return to
  automatic system behavior.
- Store the preference on the user account -- synchronizes across devices, but adds an API,
  database migration, and network behavior for a setting that is commonly device-specific.
- Compile separate light and dark themes -- duplicates theme output and ownership when Material
  already generates both schemes through `light-dark()` system tokens.

## Consequences

- (+) Material components, app surfaces, browser controls, and token-based custom styles switch
  from one global source of truth.
- (+) System mode reacts live to operating-system changes, explicit choices survive browser
  restarts, and open tabs stay synchronized.
- (+) Applying the saved attribute before Angular starts prevents a wrong-scheme flash on reload.
- (-) The preference does not follow a user to another browser or device and is shared by users
  who use the same browser profile and origin.
- (-) The head bootstrap duplicates the storage key and valid explicit values; a future strict
  Content Security Policy will need to authorize or replace the inline script.
- (-) Canvas renderers must explicitly observe the effective scheme when snapshotting CSS colors.
