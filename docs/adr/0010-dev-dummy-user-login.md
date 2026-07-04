# 0010. Dev-only dummy-user login for local testing

Date: 2026-07-02
Status: Accepted

## Context

Authentication is Google-only (ADR-0006): the only way to obtain a session is to complete a
real Google sign-in and post the resulting ID token. That is impossible to drive headlessly —
in local development, automated/agent UI checks, or CI — so every authenticated screen (i.e.
almost the whole app) is unreachable without a human clicking through Google each time. We want
a way to reach those screens locally without weakening production auth.

## Decision

Add a dev-only `/api/dev/login` endpoint that upserts a single fixed **Dev User** and mints a
real session cookie for it — reusing the exact session path that Google login uses, so `/api/me`,
the route guard, and all future authenticated API calls behave normally.

Gate it on a **runtime flag**, `DEV_AUTH`, which defaults to `false`. The route is registered
only when `DEV_AUTH=true`. As a guard against enabling it in production, the server refuses to
start (`log.Fatal`) if `DEV_AUTH=true` is combined with `SECURE_COOKIES=true` (the setting that
marks a production deployment). `make dev-api` sets `DEV_AUTH=true`; usage is documented in the
root `AGENTS.md`.

## Alternatives considered

- **Compile-time build tag** (`//go:build dev` for the route + a `!dev` no-op) — the strongest
  guarantee, since the dummy-login code is physically absent from a production binary and no
  configuration can enable it. Rejected for this project because it adds a two-file no-op seam
  that reads as indirect, the tagged file escapes the default `go vet` / `golangci-lint` / test
  run, and it needs `-tags dev` to build/run. For a small, single-operator personal app we
  preferred simpler, lint-covered, single-file code with a runtime guard.
- **Frontend-only mock user** — inject a fake `User` into `AuthService` and skip the API. It
  never creates a session, so every real API call would 401; you couldn't exercise authenticated
  data flows, which defeats the purpose.
- **Seed a real Google account for testing** — still requires completing real OAuth; brittle and
  unusable for agents/CI.

## Consequences

- (+) Local and automated testing can reach authenticated UI through the real session mechanism,
  so behaviour matches production.
- (+) Simple, linear code in one always-compiled file, covered by the normal vet/lint/test gate.
- (−) The dummy-login handler ships in the production binary (dormant). It is inert unless
  `DEV_AUTH=true`, and the startup guard refuses that in a production config — defense by
  configuration plus a guard, not a compile-time guarantee. Accepted for a personal project.
- (−) `GET /api/dev/login` mutates state (mints a session) for browser convenience — fine for a
  dev-only route, but not a pattern to copy into real endpoints.
- (−) A single hard-coded identity; multi-user dev scenarios would need more (deferred until
  needed).
