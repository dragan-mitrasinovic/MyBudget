# 0003. Backend is a standalone Go HTTP API

Date: 2026-06-29
Status: Accepted

## Context

The web app (and a future mobile app) need a backend for authentication, per-user data,
and sharing between users. We must decide how the backend is built and how clients reach
it.

## Decision

Build the backend as a standalone HTTP API in Go, living in `server/`. Clients (web,
future mobile) are independent network clients that talk to it over HTTP/JSON — the
backend is not shared as code. Start with the standard library (`net/http`,
`http.ServeMux`) and a flat layout (`main.go`), growing into `cmd/` + `internal/` only
when needed.

## Alternatives considered

- A Node/TypeScript backend (shared language with Angular) — the API boundary makes
  language-sharing low-value, and Go is preferred for the backend.
- A Backend-as-a-Service (Firebase/Supabase) — faster initially, but means adopting an
  auth/data model we'd later migrate off when building the Go backend.
- A web framework (chi/echo/gin) — deferred; the standard library is enough to start and
  is trivially swappable behind handlers.

## Consequences

- (+) One backend serves every client; clear network boundary; portable choices.
- (+) Minimal dependencies to start.
- (−) We own more plumbing (routing, sessions) than a framework or BaaS would provide.
