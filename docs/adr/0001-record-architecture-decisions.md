# 0001. Record architecture decisions

Date: 2026-06-29
Status: Accepted

## Context

MyBudget is early and making foundational choices (repo layout, backend language,
storage, auth) whose rationale is easy to forget later. We want the "why" preserved
alongside the code, not lost in chat logs or scattered notes.

## Decision

Record significant architectural and tech-stack decisions as Architecture Decision
Records (ADRs) — one Markdown file per decision in `docs/adr/`, named
`NNNN-short-title.md`, using the lean template in `docs/adr/template.md`.

ADRs are immutable: to reverse or change a past decision, add a **new** ADR and mark the
old one `Superseded by ADR-NNNN` rather than editing it. A decision merits an ADR when
it is costly to reverse or shapes the architecture or tech stack (not routine code
choices).

## Alternatives considered

- A single `DECISIONS.md` file — lighter, but grows unwieldy and decisions can't carry
  individual status as cleanly.
- The richer MADR template — more structure (explicit options + pros/cons per option),
  but more ceremony than we want per entry.

## Consequences

- (+) Decision history and rationale live with the code and survive over time.
- (+) New contributors (and future-self) can see why things are the way they are.
- (−) Requires the discipline to write an ADR when a real decision is made (see `AGENTS.md`).
