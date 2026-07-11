# 0015. Air for Go API live reload

Date: 2026-07-11
Status: Accepted

## Context

`make dev-api` previously ran the API with `go run`, so every backend change required a manual
restart. We want fast local feedback without changing the established command that supplies the
development-only authentication, application origin, and Google client configuration.

## Decision

Use Air to rebuild and restart the Go API during local development. Pin Air in the server's Go
module with a `tool` directive, invoke it as `go tool air` from `make dev-api`, and keep its
settings in `server/.air.toml`.

Air watches Go source, embedded SQL migrations, and the module files. It builds
`./cmd/mybudget`, keeps the last successful process running when a later build fails, and writes
temporary build output under the ignored `server/tmp/` directory. This workflow is for local
development only.

## Alternatives considered

- **wgo** — offers a concise `go run`-like interface, but Air has broader adoption and a checked-in
  configuration format that makes the development process explicit.
- **watchexec, Reflex, or entr** — flexible general-purpose watchers, but each requires separate
  installation and more shell-level process configuration.
- **Continue using `go run` directly** — has no file-watching support and retains the manual
  restart loop this decision is intended to remove.

## Consequences

- (+) `make dev-api` now picks up backend changes automatically while preserving its existing
  environment and URL.
- (+) Every clone uses the same Air version without requiring a global installation.
- (+) Compile errors do not take down the last successfully built development server.
- (-) Air and its transitive dependencies enlarge the server module graph and initial download.
- (-) Reloading replaces the process rather than preserving in-memory state or active
  connections; Air must not be used as a production deployment mechanism.
