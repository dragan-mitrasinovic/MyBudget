# 0006. Google-only authentication (Sign in with Google)

Date: 2026-06-29
Status: Accepted

## Context

MyBudget needs registration and login. We want the simplest secure flow and no password
management of our own, at least to start.

## Decision

Authenticate users exclusively via "Sign in with Google" (Google Identity Services),
using the authentication (ID-token) flow:

1. The web client renders the Google button and receives an ID token (a JWT).
2. It sends the token to `POST /api/auth/google`.
3. The backend verifies the token (signature against Google's public keys, `aud` = our
   client ID, `iss`, `exp`) using `google.golang.org/api/idtoken`, then upserts the user
   keyed by the Google `sub` claim and creates a session (see ADR-0005).

No passwords, and no other identity providers for now.

## Alternatives considered

- Email/password auth — we'd own password storage, reset flows, and breach risk.
- The OAuth authorization-code flow — needed only to call Google APIs on the user's
  behalf; unnecessary for identity-only login.
- Multiple social providers — more surface area than a you-and-friends app needs now.

## Consequences

- (+) No passwords to store; Google handles identity and verification.
- (+) Minimal, well-supported flow; the stable user key is the Google `sub` claim.
- (−) Hard dependency on Google; users without a Google account can't sign in.
- (−) Requires Google Cloud OAuth setup (consent screen + Web client ID), with the client
  ID configured in both the web app and the backend.
