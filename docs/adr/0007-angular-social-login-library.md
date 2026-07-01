# 0007. Google sign-in via @abacritt/angularx-social-login

Date: 2026-07-01
Status: Accepted

## Context

The web client must render the "Sign in with Google" button and obtain a Google ID token
to send to the backend (ADR-0006). The Google Identity Services (GIS) client is an untyped
script loaded at runtime, so we must decide how to load it, type it, and capture the
credential in Angular.

## Decision

Use the `@abacritt/angularx-social-login` library (the de-facto Angular social-login
package). It injects the GIS script, renders the button via the standalone
`asl-google-signin-button` directive, ships its own types, and exposes the ID token through
`SocialAuthService.authState` (`SocialUser.idToken`), which we forward to
`POST /api/auth/google`. It is configured via the `SOCIAL_AUTH_CONFIG` injection token with
One Tap and autoLogin disabled (plain-button flow).

## Alternatives considered

- Hand-written GIS integration (a dynamic `loadGis()` + a small `google.d.ts`) — zero
  dependencies and the same technique the library uses internally, but we maintain the glue
  and typings ourselves.
- Static `<script>` in `index.html` — Google's generic default, but loads GIS on every page
  and handles readiness globally rather than scoped to the login page.

## Consequences

- (+) Standard, recognizable Angular integration; the library owns script loading and types.
- (+) Less code to maintain than the hand-rolled version.
- (−) Adds a dependency (and its abstractions) that spans several providers though we use
  only Google.
- (−) `authState` is a replaying subject, so logout must call `SocialAuthService.signOut()`
  to avoid re-triggering a login.
