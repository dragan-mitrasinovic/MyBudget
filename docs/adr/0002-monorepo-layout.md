# 0002. Single monorepo with web/, server/, and future mobile/

Date: 2026-06-29
Status: Accepted

## Context

MyBudget will have multiple parts: an Angular web client, a Go HTTP API, and likely a
mobile client later. They are written in different languages, and the backend is a
network service rather than shared code.

## Decision

Keep everything in one Git monorepo with top-level folders:

- `web/` — the Angular app
- `server/` — the Go HTTP API, with its own `go.mod` inside the folder
- `mobile/` — future client

No monorepo tooling (Nx/Turborepo) — plain folders, since the languages are separate.
The Angular app stays at the repo root until the Go backend actually starts; that is the
moment we `git mv` it into `web/` and add `server/`.

## Alternatives considered

- Separate repos per component — more overhead to coordinate, harder to share an API
  contract later, and unnecessary for a small project.

## Consequences

- (+) One place for code, issues, and a future API contract; atomic cross-cutting changes.
- (+) Module boundary = folder boundary (`server/go.mod`), so Go tooling ignores the JS tree.
- (−) The web/server split is a pending restructure, done when the backend is introduced.
