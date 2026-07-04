# MyBudget web

The Angular + Angular Material single-page client. See the root [AGENTS.md](../AGENTS.md)
and [docs/adr/](../docs/adr/) for project-wide context, and ADR-0008 for the quality tooling
described here.

## Formatting & linting

Run from `web/` (or use the repo-root `make` targets):

- `npm run format` — Prettier writes `.ts`, `.html` (Angular parser), `.scss`, `.json`, and
  `.md`. `npm run format:check` verifies without writing.
- `npm run lint` / `npm run lint:fix` — `angular-eslint` (flat config in `eslint.config.js`),
  including Angular template rules.
- `npm run lint:css` / `npm run lint:css:fix` — `stylelint` on `.scss`, enforcing the design
  conventions below (no hardcoded colours/spacing). Config in `.stylelintrc.json`.
- `npm run typecheck` — `tsc --noEmit` for a fast type pass. The full `ng build` also
  type-checks templates.

Prettier owns layout; ESLint owns correctness; stylelint owns design-token discipline.
`eslint-config-prettier` disables ESLint's formatting-related rules, and stylelint's own
formatting rules are turned off, so nothing fights Prettier — don't re-add stylistic ESLint
or stylelint rules; adjust `.prettierrc` instead.

## Design & UI conventions

Consistency comes from a single source of truth, not from vigilance. See ADR-0009.

- **Material-first.** Build UI from Angular Material components (`@angular/material`) and CDK
  primitives — don't hand-roll buttons, inputs, cards, dialogs, menus, etc. If a pattern
  isn't in Material, wrap a Material/CDK primitive rather than starting from a bare element.
- **Design tokens, never literals.** Colours come from Material 3 system variables —
  `var(--mat-sys-primary)`, `--mat-sys-error`, `--mat-sys-surface`, `--mat-sys-on-surface`, …
  (see the [system variables guide](https://material.angular.dev/guide/system-variables)).
  Spacing comes from the scale in `src/styles/_tokens.scss` — `var(--space-1 … --space-8)`,
  a 4px base. No hex, named colours, or ad-hoc values for colour/spacing; stylelint rejects
  them (currently on `color`/`fill`/`stroke`/`gap`; use `--space-*` for padding/margin too).
- **Typography is Material's.** Use its type scale (`var(--mat-sys-*)` font tokens, e.g.
  `font: var(--mat-sys-headline-small)`) instead of setting `font-size` by hand.
- **Theming lives in one place.** `src/styles.scss` holds the single `mat.theme()` call;
  app-level tokens live in `src/styles/_tokens.scss`. Light theme only for now — don't scatter
  per-component colour overrides; adjust the theme or add a token.
- **Class names are kebab-case** (`login-card`, not `login__card`) — stylelint's
  `selector-class-pattern`.
- **Accessibility.** Label controls, give images meaningful `alt`, keep focus order sane.
  Material components handle most ARIA — don't undo it.

Building a shared component? Keep it token-driven and Material-based so it inherits the theme
for free, and add it under `src/app/` following the existing standalone-component pattern.

**Verify in the browser.** Don't stop at a green build/lint — run the app (`make dev-api` +
`make dev-web`, sign in via the dummy user) and actually look at the page you changed. See
[Running & verifying locally](../AGENTS.md) in the root guide.
