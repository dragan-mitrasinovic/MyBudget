# 0005. Server-side sessions via HttpOnly cookies (not JWT)

Date: 2026-06-29
Status: Accepted

## Context

After a user authenticates, the backend must keep them logged in across requests. The
app is web-first (Angular SPA), holds financial data, and a native mobile client is
planned later. The main options are an HttpOnly session cookie or a JWT bearer token
held by the SPA.

## Decision

Use opaque, server-side session IDs stored in SQLite, delivered to the web client as a
`Secure`, `HttpOnly`, `SameSite=Lax` cookie. The `sessions` table is the source of
truth. In development, `Secure` is disabled (served over `http://localhost`) and enabled
in production.

## Alternatives considered

- JWT bearer token in the SPA — uniform for a future mobile client and CSRF-free, but the
  token is readable by injected JS (an XSS risk for a money app), revocation needs a
  denylist, and doing it safely pulls in refresh-token machinery now.

## Consequences

- (+) Credential is unreadable by JavaScript; revocation is a single row delete.
- (+) No refresh-token complexity for the basic flow.
- (−) CSRF must be mitigated (`SameSite=Lax` + Origin check on a JSON-only API).
- (−) Dev cross-origin (4200 ↔ 8080) requires an Angular dev proxy so requests are same-origin.
- Mobile later: add an `Authorization: Bearer` transport over the **same** sessions table
  (token in the OS keychain) — additive, not a migration.
