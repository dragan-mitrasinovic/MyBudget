# 0012. Single explicit development-mode flag

Date: 2026-07-11
Status: Accepted

## Context

The server originally exposed two boolean environment variables: `DEV_AUTH` registered the
dummy-login route, while `SECURE_COOKIES` controlled whether session cookies were restricted to
HTTPS. Our supported configurations always use these settings together: local development needs
the dummy login and plain-HTTP localhost cookies; non-development operation must omit the auth
bypass and use Secure cookies. Independent flags therefore exposed invalid combinations and
required an extra startup guard. This supersedes ADR-0010's configuration mechanism while
retaining its dev-login behavior.

## Decision

Require one `IS_DEVELOPMENT` boolean environment variable:

- `IS_DEVELOPMENT=true` registers `/api/dev/login` and sets session cookies without `Secure`,
  so they work over `http://localhost`.
- `IS_DEVELOPMENT=false` does not register `/api/dev/login` and sets session cookies with
  `Secure`, so browsers send them only over HTTPS.

The executable parses the value with Go's standard boolean parser and fails at startup when it
is missing or invalid. It passes one `IsDevelopment` value into the HTTP layer, which derives
both behaviors; the invalid mixed state is no longer representable. `make dev-api` supplies
`IS_DEVELOPMENT=true`. Independent settings such as `PORT`, `DB_PATH`, `SESSION_TTL`, and
`GOOGLE_CLIENT_ID` remain separate configuration.

## Alternatives considered

- **Keep `DEV_AUTH` and `SECURE_COOKIES` plus the startup guard** — technically flexible, but
  exposes combinations this application does not support and makes deployment configuration
  harder to reason about.
- **Use an `APP_ENV=development|production` enum** — more extensible for future environment
  types, but adds states and string handling that current behavior does not need. Introduce it
  later if staging or testing requires distinct runtime behavior.
- **Default a missing flag to false** — fail-safe for the auth bypass, but still hides
  a deployment mistake. Requiring an explicit value makes the mode visible and testable.

## Consequences

- (+) One source of truth controls all development-only behavior.
- (+) Production cannot accidentally combine Secure cookies with the dummy-login route, and
  development cannot accidentally disable both.
- (+) Cookie transport still follows ADR-0005: plain HTTP locally, HTTPS-only in production.
- (-) Every invocation must set `IS_DEVELOPMENT`; the Makefile handles this for normal local
  work.
- (-) The boolean distinguishes only development from non-development. If a future staging or
  testing environment needs distinct behavior, replace it with a richer model in a new decision.
