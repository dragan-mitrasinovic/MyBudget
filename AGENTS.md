# MyBudget

MyBudget is a personal expense-tracking and budgeting app for the author and a few
friends, so it is multi-user from the start. The web client is an Angular + Angular
Material single-page app; a Go HTTP API (`server/`) and a future mobile client are
planned. See `docs/adr/` for the decisions behind these choices.

## Architecture decisions

Significant architectural and tech-stack decisions are recorded as ADRs (Architecture
Decision Records) in `docs/adr/` — one Markdown file per decision.

**Whenever an architectural or tech-stack decision is made, add a new ADR.** Copy
`docs/adr/template.md`, give it the next number, and fill in the context, decision,
alternatives considered, and consequences. Decisions are immutable: to change or reverse
one, add a new ADR and mark the old one `Superseded by ADR-NNNN` — don't edit or delete
the original.

## Coding conventions

Language-specific coding conventions live in per-folder `AGENTS.md` files next to the
code they govern — e.g. [`server/AGENTS.md`](server/AGENTS.md) for Go. Add new
conventions to the relevant folder's file (create `web/AGENTS.md` for Angular/TypeScript
when the need arises).
