# 0009. Design tokens and Material-first UI, enforced by stylelint

Date: 2026-07-02
Status: Accepted

## Context

We are starting UI work. The app must look consistent across many screens (and, later, a
mobile client), and much of the UI is written by AI agents — so consistency has to come from
the tools and structure, not from every author (human or agent) remembering the rules.
Angular Material 3 is already configured via a single `mat.theme()` call, which exposes
`--mat-sys-*` system variables for colour, typography, and shape.

## Decision

Standardise the UI around a single source of truth, enforced mechanically:

- **Material-first** — build from Angular Material components and CDK primitives; don't
  hand-roll equivalents.
- **Design tokens** — colour, typography, and shape come from Material's `--mat-sys-*`
  system variables; a small app-level spacing scale (`--space-1 … --space-8`, 4px base) lives
  in `web/src/styles/_tokens.scss`. Component styles reference tokens, never literals.
- **Enforcement** — `stylelint` (with `stylelint-declaration-strict-value`, `color-no-hex`,
  and `color-named: never`) rejects hardcoded colours/spacing. It is wired into `make lint`,
  `make check`, and the lefthook pre-commit hook, alongside the existing gates from ADR-0008.
  stylelint's formatting rules are disabled so Prettier remains the sole formatter.

## Alternatives considered

- **Instructions/review only** (an `AGENTS.md` note, no lint) — relies on vigilance; drifts
  silently as the app and the number of contributors/agents grow. We keep the `AGENTS.md`
  guidance *and* enforce it.
- **A bespoke design system / component library** — redundant; Angular Material already is a
  themeable Material 3 design system. Building our own is disproportionate for this app.
- **Storybook + visual-regression testing now** — valuable but heavier, and there are no
  reusable components yet. Deferred to the backlog (`TODO.md`, Someday / maybe).
- **Tailwind / utility CSS** — introduces a second styling and token system to reconcile with
  Material's theme; more friction than benefit here.

## Consequences

- (+) Consistency is enforced by construction — colour/spacing drift fails lint and the
  commit, rather than surfacing in review or in production.
- (+) Restyling is a single-point change: edit the theme or a token, and the whole app follows.
- (+) No new runtime dependencies; stylelint is dev-only.
- (−) Spacing enforcement currently covers `color`/`fill`/`stroke`/`gap`; `padding`/`margin`
  remain convention (documented, not linted) to avoid shorthand false positives — this can be
  tightened later.
- (−) One more dev tool and config to maintain, and contributors must learn the token names
  (documented in `web/AGENTS.md`).
