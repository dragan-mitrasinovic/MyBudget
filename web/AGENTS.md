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
- `npm run typecheck` — `tsc --noEmit` for a fast type pass. The full `ng build` also
  type-checks templates.

Prettier owns layout; ESLint owns correctness. `eslint-config-prettier` disables ESLint's
formatting-related rules so the two don't fight — don't re-add stylistic ESLint rules; adjust
`.prettierrc` instead.
